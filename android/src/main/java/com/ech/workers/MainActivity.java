package com.ech.workers;

import android.app.Activity;
import android.content.Intent;
import android.content.SharedPreferences;
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
        
        // Request VPN permission
        Intent intent = VpnService.prepare(this);
        if (intent != null) {
            startActivityForResult(intent, 0);
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        loadProfiles();
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
        
        statusText.setText(R.string.connecting);
        btnToggle.setEnabled(false);
        
        // TODO: 启动TunnelService
        // 简化版本：直接更新UI
        new Handler(Looper.getMainLooper()).postDelayed(() -> {
            isRunning = true;
            updateUI();
            testLatency();
        }, 1000);
    }

    private void stopService() {
        // TODO: 停止TunnelService
        isRunning = false;
        updateUI();
    }

    private void updateUI() {
        if (isRunning) {
            statusText.setText(R.string.connected);
            btnToggle.setText(R.string.stop_service);
            btnToggle.setEnabled(true);
            flagIcon.setVisibility(View.VISIBLE);
            ipText.setVisibility(View.VISIBLE);
            latencyText.setVisibility(View.VISIBLE);
            
            // TODO: 显示实际IP和国旗
            ipText.setText(getString(R.string.ip_format, "检测中..."));
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
}
