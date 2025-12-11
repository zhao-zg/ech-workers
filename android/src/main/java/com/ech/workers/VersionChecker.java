package com.ech.workers;

import android.os.AsyncTask;
import android.util.Log;

import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;

/**
 * 版本检查工具类
 */
public class VersionChecker {
    private static final String TAG = "VersionChecker";
    private static final String GITHUB_API_URL = "https://api.github.com/repos/zhao-zg/ech-workers/releases/latest";
    
    public interface VersionCheckListener {
        void onUpdateAvailable(UpdateInfo updateInfo);
        void onNoUpdate();
        void onError(String error);
    }
    
    public static class UpdateInfo {
        public String currentVersion;
        public String latestVersion;
        public String publishedAt;
        public String body;
        public String downloadUrl;
        
        public UpdateInfo(String current, String latest, String published, String body, String url) {
            this.currentVersion = current;
            this.latestVersion = latest;
            this.publishedAt = published;
            this.body = body;
            this.downloadUrl = url;
        }
    }
    
    /**
     * 获取当前版本号
     */
    public static String getCurrentVersion() {
        return BuildConfig.VERSION_NAME;
    }
    
    /**
     * 比较版本号
     * @return 1表示v1>v2, -1表示v1<v2, 0表示相等
     */
    public static int compareVersion(String v1, String v2) {
        // 移除 'v' 前缀
        v1 = v1.replaceAll("^v", "");
        v2 = v2.replaceAll("^v", "");
        
        String[] parts1 = v1.split("\\.");
        String[] parts2 = v2.split("\\.");
        
        int maxLen = Math.max(parts1.length, parts2.length);
        
        for (int i = 0; i < maxLen; i++) {
            int n1 = 0;
            int n2 = 0;
            
            if (i < parts1.length) {
                try {
                    n1 = Integer.parseInt(parts1[i].replaceAll("[^0-9]", ""));
                } catch (Exception e) {
                    n1 = 0;
                }
            }
            
            if (i < parts2.length) {
                try {
                    n2 = Integer.parseInt(parts2[i].replaceAll("[^0-9]", ""));
                } catch (Exception e) {
                    n2 = 0;
                }
            }
            
            if (n1 > n2) return 1;
            if (n1 < n2) return -1;
        }
        
        return 0;
    }
    
    /**
     * 检查更新(异步)
     */
    public static void checkUpdate(final VersionCheckListener listener) {
        new AsyncTask<Void, Void, UpdateInfo>() {
            private String errorMessage = null;
            
            @Override
            protected UpdateInfo doInBackground(Void... voids) {
                HttpURLConnection connection = null;
                try {
                    URL url = new URL(GITHUB_API_URL);
                    connection = (HttpURLConnection) url.openConnection();
                    connection.setRequestMethod("GET");
                    connection.setConnectTimeout(10000);
                    connection.setReadTimeout(10000);
                    connection.setRequestProperty("User-Agent", "ECH-Workers-Android/" + getCurrentVersion());
                    
                    int responseCode = connection.getResponseCode();
                    if (responseCode != HttpURLConnection.HTTP_OK) {
                        errorMessage = "GitHub API 返回错误: " + responseCode;
                        return null;
                    }
                    
                    BufferedReader reader = new BufferedReader(
                            new InputStreamReader(connection.getInputStream()));
                    StringBuilder response = new StringBuilder();
                    String line;
                    while ((line = reader.readLine()) != null) {
                        response.append(line);
                    }
                    reader.close();
                    
                    JSONObject json = new JSONObject(response.toString());
                    String latestVersion = json.optString("tag_name", "");
                    String publishedAt = json.optString("published_at", "");
                    String body = json.optString("body", "");
                    String htmlUrl = json.optString("html_url", "");
                    
                    if (latestVersion.isEmpty()) {
                        errorMessage = "无法获取版本信息";
                        return null;
                    }
                    
                    String currentVersion = getCurrentVersion();
                    
                    // 只有新版本时才返回UpdateInfo
                    if (compareVersion(latestVersion, currentVersion) > 0) {
                        return new UpdateInfo(currentVersion, latestVersion, publishedAt, body, htmlUrl);
                    }
                    
                    return null;
                    
                } catch (Exception e) {
                    Log.e(TAG, "检查更新失败", e);
                    errorMessage = "检查更新失败: " + e.getMessage();
                    return null;
                } finally {
                    if (connection != null) {
                        connection.disconnect();
                    }
                }
            }
            
            @Override
            protected void onPostExecute(UpdateInfo updateInfo) {
                if (errorMessage != null) {
                    listener.onError(errorMessage);
                } else if (updateInfo != null) {
                    listener.onUpdateAvailable(updateInfo);
                } else {
                    listener.onNoUpdate();
                }
            }
        }.execute();
    }
    
    /**
     * 格式化发布时间
     */
    public static String formatPublishedDate(String publishedAt) {
        try {
            SimpleDateFormat inputFormat = new SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US);
            SimpleDateFormat outputFormat = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault());
            Date date = inputFormat.parse(publishedAt);
            return outputFormat.format(date);
        } catch (Exception e) {
            return publishedAt;
        }
    }
}
