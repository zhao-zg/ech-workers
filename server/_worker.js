const WS_OPEN = 1, WS_CLOSING = 2;
const encoder = new TextEncoder();

import { connect } from 'cloudflare:sockets';

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

  const connectRemote = async (addr, data) => {
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

    const hosts = [host, ...fallbackHosts];
    
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
          const pipeIndex = payload.indexOf('|');
          const addr = pipeIndex === -1 ? payload : payload.slice(0, pipeIndex);
          const extra = pipeIndex === -1 ? '' : payload.slice(pipeIndex + 1);
          await connectRemote(addr, extra);
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