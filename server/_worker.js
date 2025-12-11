const WS_OPEN = 1, WS_CLOSING = 2;
const encoder = new TextEncoder();

import { connect } from 'cloudflare:sockets';

/**
 * ECH-Workers WebSocket Proxy Server
 * 
 * 环境变量:
 * - TOKEN: 可选，客户端认证令牌
 * - FALLBACK_HOSTS: 可选，备用连接IP列表，逗号分隔，默认为 2a00:1098:2b::1:6815:5881
 * 
 * 客户端协议:
 * - CONNECT:addr|data#fallbackHost - 连接目标，可选携带初始数据和fallbackHost
 *   - addr: 目标地址（支持 [IPv6]:port 或 host:port 格式）
 *   - data: 可选，连接后立即发送的数据
 *   - fallbackHost: 可选，客户端指定的反代Host（支持域名和IP），优先使用此Host连接目标
 * - DATA:text - 发送文本数据
 * - Binary data - 发送二进制数据
 * - CLOSE - 关闭连接
 */

export default {
  async fetch(req, env) {
    const token = env.TOKEN || '';
    const fallbackHosts = (env.FALLBACK_HOSTS || '2a00:1098:2b::1:6815:5881')
      .split(',')
      .map(h => h.trim())
      .filter(h => h);
    
    const upgrade = req.headers.get('Upgrade');
    
    if (upgrade !== 'websocket') {
      return req.url.endsWith('/') 
        ? new Response('WebSocket Proxy', { status: 200 })
        : new Response('WebSocket required', { status: 426 });
    }

    if (token && req.headers.get('Sec-WebSocket-Protocol') !== token) {
      return new Response('Unauthorized', { status: 401 });
    }

    const [client, server] = Object.values(new WebSocketPair());
    server.accept();
    handleConn(server, fallbackHosts).catch(() => closeWS(server));
    
    const init = { status: 101, webSocket: client };
    if (token) init.headers = { 'Sec-WebSocket-Protocol': token };
    return new Response(null, init);
  }
};

async function handleConn(ws, fallbackHosts) {
  let remote, writer, reader;
  let closed = false;

  const cleanup = () => {
    if (closed) return;
    closed = true;
    [writer, reader, remote].forEach(s => s?.close?.());
    writer = reader = remote = null;
    closeWS(ws);
  };

  const relay = async () => {
    try {
      while (!closed && reader) {
        const { done, value } = await reader.read();
        if (done || ws.readyState !== WS_OPEN) break;
        if (value?.length) ws.send(value);
      }
    } catch {}
    !closed && cleanup();
  };

  const connectRemote = async (addr, data, fallbackHost) => {
    // 支持 [IPv6]:port 或 host:port 格式
    let host, port;
    if (addr.startsWith('[')) {
      const end = addr.indexOf(']');
      if (end === -1) throw new Error('Invalid IPv6 format');
      host = addr.slice(1, end);
      port = addr.slice(end + 2);
    } else {
      const lastColon = addr.lastIndexOf(':');
      if (lastColon === -1) throw new Error('Missing port');
      host = addr.slice(0, lastColon);
      port = addr.slice(lastColon + 1);
    }
    port = parseInt(port, 10);
    if (!port || port < 1 || port > 65535) throw new Error('Invalid port');

    // 如果客户端传入了fallbackHost，优先使用fallbackHost，否则使用原始host和env.FALLBACK_HOSTS
    const hosts = fallbackHost ? [fallbackHost] : [host, ...fallbackHosts];
    
    for (const h of hosts) {
      try {
        remote = connect({ hostname: h, port });
        await remote.opened;
        writer = remote.writable.getWriter();
        reader = remote.readable.getReader();
        
        if (data) await writer.write(encoder.encode(data));
        ws.send('CONNECTED');
        relay();
        return;
      } catch (e) {
        [writer, reader, remote].forEach(s => s?.close?.());
        writer = reader = remote = null;
        if (!e.message?.toLowerCase().includes('cloudflare')) throw e;
      }
    }
  };

  ws.addEventListener('message', async ({ data }) => {
    if (closed) return;
    
    try {
      if (typeof data === 'string') {
        if (data.startsWith('CONNECT:')) {
          const payload = data.slice(8);
          // 格式: CONNECT:addr|extra#fallbackHost 或 CONNECT:addr|extra 或 CONNECT:addr#fallbackHost
          const fallbackIndex = payload.indexOf('#');
          const pipeIndex = payload.indexOf('|');
          
          let addr, extra, fallbackHost;
          if (fallbackIndex !== -1) {
            // 有fallbackHost
            const beforeFallback = payload.slice(0, fallbackIndex);
            fallbackHost = payload.slice(fallbackIndex + 1);
            const pipe = beforeFallback.indexOf('|');
            addr = pipe === -1 ? beforeFallback : beforeFallback.slice(0, pipe);
            extra = pipe === -1 ? '' : beforeFallback.slice(pipe + 1);
          } else {
            // 没有fallbackHost，使用原逻辑
            addr = pipeIndex === -1 ? payload : payload.slice(0, pipeIndex);
            extra = pipeIndex === -1 ? '' : payload.slice(pipeIndex + 1);
            fallbackHost = '';
          }
          await connectRemote(addr, extra, fallbackHost);
        } else if (data.startsWith('DATA:')) {
          writer?.write(encoder.encode(data.slice(5)));
        } else if (data === 'CLOSE') cleanup();
      } else if (data instanceof ArrayBuffer) {
        writer?.write(new Uint8Array(data));
      }
    } catch (e) {
      ws.send(`ERROR:${e.message}`);
      cleanup();
    }
  });

  ws.addEventListener('close', cleanup);
  ws.addEventListener('error', cleanup);
}

function closeWS(ws) {
  try {
    if (ws.readyState === WS_OPEN || ws.readyState === WS_CLOSING) {
      ws.close(1000, 'Server closed');
    }
  } catch {}
}