/*
 ============================================================================
 Name        : ChinaIpListManager.java
 Author      : ECH Workers
 Description : 中国IP列表管理器
 ============================================================================
 */

package com.ech.workers;

import android.content.Context;
import android.util.Log;
import com.ech.workers.tunnel.Tunnel;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;

public class ChinaIpListManager {
    private static final String TAG = "ChinaIpListManager";
    
    // IP列表文件名
    private static final String IPV4_FILE = "chn_ip.txt";
    private static final String IPV6_FILE = "chn_ip_v6.txt";
    
    // GitHub 原始文件 URL
    private static final String IPV4_URL = "https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt";
    private static final String IPV6_URL = "https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip_v6.txt";
    
    // 更新间隔（7天）
    private static final long UPDATE_INTERVAL = 7 * 24 * 60 * 60 * 1000L;
    
    private Context context;
    private Preferences prefs;
    
    public ChinaIpListManager(Context context) {
        this.context = context;
        this.prefs = new Preferences(context);
    }
    
    /**
     * 检查是否需要更新IP列表
     */
    public boolean needsUpdate() {
        if (!prefs.isChinaIpListLoaded()) {
            return true;
        }
        
        long lastUpdate = prefs.getChinaIpListUpdateTime();
        long now = System.currentTimeMillis();
        return (now - lastUpdate) > UPDATE_INTERVAL;
    }
    
    /**
     * 下载并加载中国IP列表
     */
    public void downloadAndLoad(final DownloadCallback callback) {
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    Log.i(TAG, "开始下载中国IP列表...");
                    
                    // 下载 IPv4 列表
                    File ipv4File = new File(context.getFilesDir(), IPV4_FILE);
                    if (!downloadFile(IPV4_URL, ipv4File)) {
                        if (callback != null) {
                            callback.onError("下载 IPv4 列表失败");
                        }
                        return;
                    }
                    
                    // 下载 IPv6 列表
                    File ipv6File = new File(context.getFilesDir(), IPV6_FILE);
                    if (!downloadFile(IPV6_URL, ipv6File)) {
                        Log.w(TAG, "下载 IPv6 列表失败，将继续使用 IPv4");
                    }
                    
                    // 加载到 Go 层
                    if (loadIpListToGo(ipv4File, ipv6File)) {
                        prefs.setChinaIpListLoaded(true);
                        prefs.setChinaIpListUpdateTime(System.currentTimeMillis());
                        Log.i(TAG, "中国IP列表下载并加载成功");
                        
                        if (callback != null) {
                            callback.onSuccess();
                        }
                    } else {
                        if (callback != null) {
                            callback.onError("加载IP列表到Go层失败");
                        }
                    }
                    
                } catch (Exception e) {
                    Log.e(TAG, "下载IP列表时出错", e);
                    if (callback != null) {
                        callback.onError(e.getMessage());
                    }
                }
            }
        }).start();
    }
    
    /**
     * 从本地文件加载IP列表
     */
    public boolean loadFromLocal() {
        try {
            File ipv4File = new File(context.getFilesDir(), IPV4_FILE);
            File ipv6File = new File(context.getFilesDir(), IPV6_FILE);
            
            if (!ipv4File.exists()) {
                Log.w(TAG, "本地 IPv4 文件不存在");
                return false;
            }
            
            return loadIpListToGo(ipv4File, ipv6File);
            
        } catch (Exception e) {
            Log.e(TAG, "加载本地IP列表时出错", e);
            return false;
        }
    }
    
    /**
     * 下载文件
     */
    private boolean downloadFile(String urlStr, File destFile) {
        HttpURLConnection connection = null;
        InputStream input = null;
        FileOutputStream output = null;
        
        try {
            URL url = new URL(urlStr);
            connection = (HttpURLConnection) url.openConnection();
            connection.setConnectTimeout(30000);
            connection.setReadTimeout(30000);
            connection.setRequestMethod("GET");
            
            int responseCode = connection.getResponseCode();
            if (responseCode != HttpURLConnection.HTTP_OK) {
                Log.e(TAG, "下载失败，HTTP " + responseCode + ": " + urlStr);
                return false;
            }
            
            input = connection.getInputStream();
            output = new FileOutputStream(destFile);
            
            byte[] buffer = new byte[8192];
            int bytesRead;
            while ((bytesRead = input.read(buffer)) != -1) {
                output.write(buffer, 0, bytesRead);
            }
            
            Log.i(TAG, "下载成功: " + destFile.getName());
            return true;
            
        } catch (Exception e) {
            Log.e(TAG, "下载文件时出错: " + urlStr, e);
            return false;
            
        } finally {
            try {
                if (output != null) output.close();
                if (input != null) input.close();
                if (connection != null) connection.disconnect();
            } catch (IOException e) {
                // Ignore
            }
        }
    }
    
    /**
     * 读取文件内容
     */
    private String readFile(File file) throws IOException {
        if (!file.exists()) {
            return "";
        }
        
        StringBuilder sb = new StringBuilder();
        BufferedReader reader = null;
        
        try {
            reader = new BufferedReader(new InputStreamReader(new FileInputStream(file)));
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line).append("\n");
            }
        } finally {
            if (reader != null) {
                try {
                    reader.close();
                } catch (IOException e) {
                    // Ignore
                }
            }
        }
        
        return sb.toString();
    }
    
    /**
     * 加载IP列表到Go层
     */
    private boolean loadIpListToGo(File ipv4File, File ipv6File) {
        try {
            String ipv4Data = readFile(ipv4File);
            String ipv6Data = ipv6File.exists() ? readFile(ipv6File) : "";
            
            Log.i(TAG, "IPv4 数据长度: " + ipv4Data.length());
            Log.i(TAG, "IPv6 数据长度: " + ipv6Data.length());
            
            // 调用 Go 层的加载函数
            Tunnel.loadChinaIPList(ipv4Data, ipv6Data);
            
            return true;
            
        } catch (Exception e) {
            Log.e(TAG, "加载IP列表到Go层时出错", e);
            return false;
        }
    }
    
    /**
     * 下载回调接口
     */
    public interface DownloadCallback {
        void onSuccess();
        void onError(String message);
    }
}
