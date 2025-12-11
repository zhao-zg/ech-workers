package com.ech.workers;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.SharedPreferences;
import android.os.Bundle;
import android.text.InputType;
import android.view.View;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Spinner;
import android.widget.Toast;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.Set;

public class SettingsActivity extends Activity {
    private SharedPreferences prefs;
    private Spinner profileSpinner;
    private Spinner routingModeSpinner;
    private EditText serverAddrEdit;
    private EditText serverIpEdit;
    private EditText fallbackHostsEdit;
    private EditText tokenEdit;
    private EditText dnsServerEdit;
    private EditText echDomainEdit;
    
    private ArrayList<String> profiles;
    private ArrayAdapter<String> profileAdapter;
    private String currentProfile;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_settings);
        
        prefs = getSharedPreferences("profiles", MODE_PRIVATE);
        
        initViews();
        loadProfiles();
        setupListeners();
    }

    private void initViews() {
        profileSpinner = findViewById(R.id.config_profile_spinner);
        routingModeSpinner = findViewById(R.id.routing_mode);
        serverAddrEdit = findViewById(R.id.server_addr);
        serverIpEdit = findViewById(R.id.server_ip);
        fallbackHostsEdit = findViewById(R.id.fallback_hosts);
        tokenEdit = findViewById(R.id.token);
        dnsServerEdit = findViewById(R.id.dns_server);
        echDomainEdit = findViewById(R.id.ech_domain);
        
        // 设置路由模式Spinner
        ArrayAdapter<CharSequence> routingAdapter = ArrayAdapter.createFromResource(this,
                R.array.routing_modes, android.R.layout.simple_spinner_item);
        routingAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        routingModeSpinner.setAdapter(routingAdapter);
    }

    private void loadProfiles() {
        Set<String> profileSet = prefs.getStringSet("profile_list", new HashSet<>());
        profiles = new ArrayList<>(profileSet);
        
        if (profiles.isEmpty()) {
            profiles.add(getString(R.string.default_profile_name));
            saveProfileList();
        }
        
        profileAdapter = new ArrayAdapter<>(this, android.R.layout.simple_spinner_item, profiles);
        profileAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        profileSpinner.setAdapter(profileAdapter);
        
        String lastProfile = prefs.getString("last_profile", profiles.get(0));
        int position = profiles.indexOf(lastProfile);
        if (position >= 0) {
            profileSpinner.setSelection(position);
        }
    }

    private void setupListeners() {
        profileSpinner.setOnItemSelectedListener(new AdapterView.OnItemSelectedListener() {
            @Override
            public void onItemSelected(AdapterView<?> parent, View view, int position, long id) {
                currentProfile = profiles.get(position);
                loadProfileConfig(currentProfile);
            }

            @Override
            public void onNothingSelected(AdapterView<?> parent) {}
        });

        findViewById(R.id.btn_back).setOnClickListener(v -> finish());
        
        findViewById(R.id.btn_save).setOnClickListener(v -> saveCurrentProfile());
        
        findViewById(R.id.btn_add_profile).setOnClickListener(v -> showAddProfileDialog());
        
        findViewById(R.id.btn_delete_profile).setOnClickListener(v -> deleteCurrentProfile());
    }

    private void loadProfileConfig(String profileName) {
        String prefix = "profile_" + profileName + "_";
        serverAddrEdit.setText(prefs.getString(prefix + "server_addr", ""));
        serverIpEdit.setText(prefs.getString(prefix + "server_ip", "mfa.gov.ua"));
        fallbackHostsEdit.setText(prefs.getString(prefix + "fallback_hosts", ""));
        tokenEdit.setText(prefs.getString(prefix + "token", ""));
        dnsServerEdit.setText(prefs.getString(prefix + "dns_server", "dns.alidns.com/dns-query"));
        echDomainEdit.setText(prefs.getString(prefix + "ech_domain", "cloudflare-ech.com"));
        
        String routingMode = prefs.getString(prefix + "routing_mode", "bypass_cn");
        int routingPosition = getRoutingModePosition(routingMode);
        routingModeSpinner.setSelection(routingPosition);
    }

    private void saveCurrentProfile() {
        if (currentProfile == null) return;
        
        String prefix = "profile_" + currentProfile + "_";
        SharedPreferences.Editor editor = prefs.edit();
        
        editor.putString(prefix + "server_addr", serverAddrEdit.getText().toString());
        editor.putString(prefix + "server_ip", serverIpEdit.getText().toString());
        editor.putString(prefix + "fallback_hosts", fallbackHostsEdit.getText().toString());
        editor.putString(prefix + "token", tokenEdit.getText().toString());
        editor.putString(prefix + "dns_server", dnsServerEdit.getText().toString());
        editor.putString(prefix + "ech_domain", echDomainEdit.getText().toString());
        
        String routingMode = getRoutingModeValue(routingModeSpinner.getSelectedItemPosition());
        editor.putString(prefix + "routing_mode", routingMode);
        
        editor.putString("last_profile", currentProfile);
        editor.apply();
        
        Toast.makeText(this, R.string.toast_saved, Toast.LENGTH_SHORT).show();
    }

    private void showAddProfileDialog() {
        AlertDialog.Builder builder = new AlertDialog.Builder(this);
        builder.setTitle(R.string.dialog_title_add);
        
        final EditText input = new EditText(this);
        input.setInputType(InputType.TYPE_CLASS_TEXT);
        input.setHint(R.string.dialog_hint_name);
        builder.setView(input);
        
        builder.setPositiveButton(R.string.ok, (dialog, which) -> {
            String name = input.getText().toString().trim();
            if (name.isEmpty()) {
                Toast.makeText(this, R.string.toast_name_empty, Toast.LENGTH_SHORT).show();
                return;
            }
            if (profiles.contains(name)) {
                Toast.makeText(this, R.string.toast_profile_exists, Toast.LENGTH_SHORT).show();
                return;
            }
            
            profiles.add(name);
            saveProfileList();
            profileAdapter.notifyDataSetChanged();
            profileSpinner.setSelection(profiles.size() - 1);
        });
        
        builder.setNegativeButton(R.string.cancel, null);
        builder.show();
    }

    private void deleteCurrentProfile() {
        if (profiles.size() <= 1) {
            Toast.makeText(this, R.string.toast_cannot_delete_last, Toast.LENGTH_SHORT).show();
            return;
        }
        
        new AlertDialog.Builder(this)
                .setTitle(R.string.dialog_title_delete)
                .setMessage(currentProfile)
                .setPositiveButton(R.string.ok, (dialog, which) -> {
                    profiles.remove(currentProfile);
                    saveProfileList();
                    
                    // 删除配置数据
                    String prefix = "profile_" + currentProfile + "_";
                    SharedPreferences.Editor editor = prefs.edit();
                    editor.remove(prefix + "server_addr");
                    editor.remove(prefix + "server_ip");
                    editor.remove(prefix + "fallback_hosts");
                    editor.remove(prefix + "token");
                    editor.remove(prefix + "dns_server");
                    editor.remove(prefix + "ech_domain");
                    editor.remove(prefix + "routing_mode");
                    editor.apply();
                    
                    profileAdapter.notifyDataSetChanged();
                    profileSpinner.setSelection(0);
                })
                .setNegativeButton(R.string.cancel, null)
                .show();
    }

    private void saveProfileList() {
        Set<String> profileSet = new HashSet<>(profiles);
        prefs.edit().putStringSet("profile_list", profileSet).apply();
    }

    private int getRoutingModePosition(String mode) {
        switch (mode) {
            case "global": return 0;
            case "bypass_cn": return 1;
            case "none": return 2;
            default: return 1; // bypass_cn
        }
    }

    private String getRoutingModeValue(int position) {
        switch (position) {
            case 0: return "global";
            case 1: return "bypass_cn";
            case 2: return "none";
            default: return "bypass_cn";
        }
    }
}
