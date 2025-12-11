package com.ech.workers;

import android.app.Activity;
import android.content.Intent;
import android.content.SharedPreferences;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.VpnService;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.View;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.ImageView;
import android.widget.Spinner;
import android.widget.TextView;
import android.widget.Toast;

import java.net.InetSocketAddress;
import java.net.Socket;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.Set;

import javax.net.ssl.SSLSocket;
import javax.net.ssl.SSLSocketFactory;

public class MainActivity extends Activity {
    private SharedPreferences prefs;
    private Spinner profileSpinner;
    private Button btnSettings;
    private Button btnToggle;
    private ImageView flagIcon;
    private TextView statusText;
    private TextView ipText;
    private TextView latencyText;
    
    private ArrayList<String> profiles;
    private ArrayAdapter<String> profileAdapter;
    private String currentProfile;
    private boolean isRunning = false;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        
        prefs = getSharedPreferences("profiles", MODE_PRIVATE);
        
        initViews();
        loadProfiles();
        setupListeners();
    }

    @Override
    protected void onResume() {
        super.onResume();
        loadProfiles();
        checkServiceStatus();
        updateUI();
    }

    private void initViews() {
        profileSpinner = findViewById(R.id.profile_spinner);
        btnSettings = findViewById(R.id.btn_settings);
        btnToggle = findViewById(R.id.btn_toggle);
        flagIcon = findViewById(R.id.flag_icon);
        statusText = findViewById(R.id.status_text);
        ipText = findViewById(R.id.ip_text);
        latencyText = findViewById(R.id.latency_text);
    }

    private void loadProfiles() {
        Set<String> profileSet = prefs.getStringSet("profile_list", new HashSet<>());
        profiles = new ArrayList<>(profileSet);
        
        if (profiles.isEmpty()) {
            profiles.add(getString(R.string.default_profile_name));
            Set<String> newSet = new HashSet<>(profiles);
            prefs.edit().putStringSet("profile_list", newSet).apply();
        }
        
        profileAdapter = new ArrayAdapter<>(this, android.R.layout.simple_spinner_item, profiles);
        profileAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        profileSpinner.setAdapter(profileAdapter);
        
        String lastProfile = prefs.getString("last_profile", profiles.get(0));
        int position = profiles.indexOf(lastProfile);
        if (position >= 0) {
            profileSpinner.setSelection(position);
            currentProfile = lastProfile;
        }
    }

    private void setupListeners() {
        profileSpinner.setOnItemSelectedListener(new AdapterView.OnItemSelectedListener() {
            @Override
            public void onItemSelected(AdapterView<?> parent, View view, int position, long id) {
                currentProfile = profiles.get(position);
                prefs.edit().putString("last_profile", currentProfile).apply();
                
                if (isRunning) {
                    stopService();
                }
                updateUI();
            }

            @Override
            public void onNothingSelected(AdapterView<?> parent) {}
        });

        btnSettings.setOnClickListener(v -> {
            Intent intent = new Intent(MainActivity.this, SettingsActivity.class);
            startActivity(intent);
        });

        btnToggle.setOnClickListener(v -> {
            if (!hasValidConfig()) {
                Toast.makeText(this, R.string.no_profile_configured, Toast.LENGTH_LONG).show();
                Intent intent = new Intent(MainActivity.this, SettingsActivity.class);
                startActivity(intent);
                return;
            }
            
            if (isRunning) {
                stopService();
            } else {
                startService();
            }
        });
    }

    private boolean hasValidConfig() {
        if (currentProfile == null) return false;
        String prefix = "profile_" + currentProfile + "_";
        String serverAddr = prefs.getString(prefix + "server_addr", "");
        return !serverAddr.isEmpty();
    }

    private void startService() {
        if (currentProfile == null) return;
        
        String prefix = "profile_" + currentProfile + "_";
        String serverAddr = prefs.getString(prefix + "server_addr", "");
        
        if (serverAddr.isEmpty()) {
            Toast.makeText(this, R.string.no_profile_configured, Toast.LENGTH_SHORT).show();
            return;
        }
        
        // 如果 serverAddr 没有端口号，自动添加 :443
        if (!serverAddr.contains(":") || serverAddr.indexOf("://") > serverAddr.lastIndexOf(":")) {
            serverAddr = serverAddr + ":443";
        }
        
        // 保存当前配置到 Preferences 用于 TProxyService
        Preferences appPrefs = new Preferences(this);
        appPrefs.setWssAddr(serverAddr);
        appPrefs.setPrefIp(prefs.getString(prefix + "server_ip", "mfa.gov.ua"));
        appPrefs.setFallbackHosts(prefs.getString(prefix + "fallback_hosts", ""));
        appPrefs.setToken(prefs.getString(prefix + "token", ""));
        appPrefs.setEchDns(prefs.getString(prefix + "dns_server", "dns.alidns.com/dns-query"));
        appPrefs.setEchDomain(prefs.getString(prefix + "ech_domain", "cloudflare-ech.com"));
        appPrefs.setRoutingMode(prefs.getString(prefix + "routing_mode", "bypass_cn"));
        
        statusText.setText(R.string.connecting);
        btnToggle.setEnabled(false);
        
        // 启动 VPN 服务
        Intent intent = VpnService.prepare(this);
        if (intent != null) {
            startActivityForResult(intent, 100);
        } else {
            startVpnService();
        }
    }
    
    private void startVpnService() {
        Intent intent = new Intent(this, TProxyService.class);
        intent.setAction(TProxyService.ACTION_CONNECT);
        startService(intent);
        
        new Handler(Looper.getMainLooper()).postDelayed(() -> {
            isRunning = true;
            updateUI();
            testLatency();
        }, 1000);
    }
    
    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode == 100) {
            if (resultCode == RESULT_OK) {
                startVpnService();
            } else {
                Toast.makeText(this, "VPN权限被拒绝", Toast.LENGTH_SHORT).show();
                btnToggle.setEnabled(true);
                statusText.setText(R.string.not_connected);
            }
        }
    }

    private void stopService() {
        Intent intent = new Intent(this, TProxyService.class);
        intent.setAction(TProxyService.ACTION_DISCONNECT);
        startService(intent);
        
        isRunning = false;
        updateUI();
    }

    private void checkServiceStatus() {
        // 检查 VPN 服务是否真的在运行
        ConnectivityManager cm = (ConnectivityManager) getSystemService(CONNECTIVITY_SERVICE);
        Network activeNetwork = cm.getActiveNetwork();
        if (activeNetwork != null) {
            NetworkCapabilities caps = cm.getNetworkCapabilities(activeNetwork);
            if (caps != null && caps.hasTransport(NetworkCapabilities.TRANSPORT_VPN)) {
                isRunning = true;
                return;
            }
        }
        isRunning = false;
    }
    
    private void updateUI() {
        if (isRunning) {
            statusText.setText(R.string.connected);
            btnToggle.setText(R.string.stop_service);
            btnToggle.setEnabled(true);
            flagIcon.setVisibility(View.VISIBLE);
            ipText.setVisibility(View.VISIBLE);
            latencyText.setVisibility(View.VISIBLE);
            
            // 获取真实 IP 和国家信息
            fetchIpInfo();
        } else {
            statusText.setText(R.string.not_connected);
            btnToggle.setText(R.string.start_service);
            btnToggle.setEnabled(true);
            flagIcon.setVisibility(View.GONE);
            ipText.setVisibility(View.GONE);
            latencyText.setVisibility(View.GONE);
            latencyText.setText("");
        }
    }

    private void testLatency() {
        if (currentProfile == null) return;
        
        String prefix = "profile_" + currentProfile + "_";
        String serverIp = prefs.getString(prefix + "server_ip", "mfa.gov.ua");
        
        latencyText.setText(R.string.testing);
        
        new Thread(() -> {
            try {
                long startTime = System.currentTimeMillis();
                SSLSocketFactory factory = (SSLSocketFactory) SSLSocketFactory.getDefault();
                SSLSocket socket = (SSLSocket) factory.createSocket();
                socket.connect(new InetSocketAddress(serverIp, 443), 5000);
                socket.close();
                long latency = System.currentTimeMillis() - startTime;
                
                runOnUiThread(() -> {
                    latencyText.setText(getString(R.string.latency_format, latency));
                });
            } catch (Exception e) {
                runOnUiThread(() -> {
                    latencyText.setText(R.string.latency_timeout);
                });
            }
        }).start();
    }
    
    private void fetchIpInfo() {
        ipText.setText(getString(R.string.ip_format, "检测中..."));
        
        new Thread(() -> {
            try {
                java.net.URL url = new java.net.URL("https://api.ip.sb/geoip");
                java.net.HttpURLConnection conn = (java.net.HttpURLConnection) url.openConnection();
                conn.setRequestMethod("GET");
                conn.setConnectTimeout(5000);
                conn.setReadTimeout(5000);
                
                int responseCode = conn.getResponseCode();
                if (responseCode == 200) {
                    java.io.BufferedReader reader = new java.io.BufferedReader(
                        new java.io.InputStreamReader(conn.getInputStream()));
                    StringBuilder response = new StringBuilder();
                    String line;
                    while ((line = reader.readLine()) != null) {
                        response.append(line);
                    }
                    reader.close();
                    
                    // 解析 JSON (简单解析)
                    String json = response.toString();
                    String ip = extractJsonValue(json, "ip");
                    String country = extractJsonValue(json, "country");
                    String countryCode = extractJsonValue(json, "country_code");
                    
                    final String displayText = ip + " (" + country + ")";
                    final String flagEmoji = getFlagEmoji(countryCode);
                    
                    runOnUiThread(() -> {
                        ipText.setText(getString(R.string.ip_format, displayText));
                        // 可以设置国旗表情符号或图标
                        if (!flagEmoji.isEmpty()) {
                            statusText.setText(getString(R.string.connected) + " " + flagEmoji);
                        }
                    });
                } else {
                    runOnUiThread(() -> {
                        ipText.setText(getString(R.string.ip_format, "检测失败"));
                    });
                }
                conn.disconnect();
            } catch (Exception e) {
                runOnUiThread(() -> {
                    ipText.setText(getString(R.string.ip_format, "检测失败"));
                });
            }
        }).start();
    }
    
    private String extractJsonValue(String json, String key) {
        try {
            String searchKey = "\"" + key + "\":";
            int startIndex = json.indexOf(searchKey);
            if (startIndex == -1) return "";
            
            startIndex += searchKey.length();
            // Skip whitespace and quotes
            while (startIndex < json.length() && 
                   (json.charAt(startIndex) == ' ' || json.charAt(startIndex) == '\"')) {
                startIndex++;
            }
            
            int endIndex = startIndex;
            while (endIndex < json.length() && 
                   json.charAt(endIndex) != '\"' && 
                   json.charAt(endIndex) != ',' && 
                   json.charAt(endIndex) != '}') {
                endIndex++;
            }
            
            return json.substring(startIndex, endIndex).trim();
        } catch (Exception e) {
            return "";
        }
    }
    
    private String getFlagEmoji(String countryCode) {
        if (countryCode == null || countryCode.length() != 2) return "";
        
        countryCode = countryCode.toUpperCase();
        int firstChar = Character.codePointAt(countryCode, 0) - 0x41 + 0x1F1E6;
        int secondChar = Character.codePointAt(countryCode, 1) - 0x41 + 0x1F1E6;
        
        return new String(Character.toChars(firstChar)) + new String(Character.toChars(secondChar));
    }
}
