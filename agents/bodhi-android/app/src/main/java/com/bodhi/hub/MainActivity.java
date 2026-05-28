package com.bodhi.hub;

import android.app.Activity;
import android.content.Intent;
import android.graphics.Color;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class MainActivity extends Activity {

    private static final String HUB_URL = "http://127.0.0.1:8080";
    private static final int MAX_WAIT_MS = 15_000;
    private static final int POLL_INTERVAL_MS = 300;

    private WebView webView;
    private LinearLayout splashLayout;
    private TextView statusText;
    private ProgressBar progressBar;
    private final ExecutorService executor = Executors.newSingleThreadExecutor();
    private final Handler mainHandler = new Handler(Looper.getMainLooper());

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // Edge-to-edge full-screen
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(
                WindowManager.LayoutParams.FLAG_FULLSCREEN,
                WindowManager.LayoutParams.FLAG_FULLSCREEN);

        buildLayout();
        startService(new Intent(this, BodhiService.class));
        waitForHub();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        executor.shutdownNow();
    }

    @Override
    public void onBackPressed() {
        if (webView.canGoBack()) {
            webView.goBack();
        } else {
            // Keep service running, just go home
            moveTaskToBack(true);
        }
    }

    // ── Layout ────────────────────────────────────────────────────────────────

    private void buildLayout() {
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setBackgroundColor(Color.parseColor("#0f172a"));

        // Splash screen shown while hub starts
        splashLayout = new LinearLayout(this);
        splashLayout.setOrientation(LinearLayout.VERTICAL);
        splashLayout.setGravity(android.view.Gravity.CENTER);
        splashLayout.setBackgroundColor(Color.parseColor("#0f172a"));
        LinearLayout.LayoutParams fill = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.MATCH_PARENT);
        splashLayout.setLayoutParams(fill);

        TextView logo = new TextView(this);
        logo.setText("🌸");
        logo.setTextSize(64);
        logo.setGravity(android.view.Gravity.CENTER);
        splashLayout.addView(logo);

        TextView title = new TextView(this);
        title.setText("Bodhi");
        title.setTextColor(Color.WHITE);
        title.setTextSize(32);
        title.setGravity(android.view.Gravity.CENTER);
        splashLayout.addView(title);

        TextView subtitle = new TextView(this);
        subtitle.setText("35 specialist agents");
        subtitle.setTextColor(Color.parseColor("#94a3b8"));
        subtitle.setTextSize(14);
        subtitle.setGravity(android.view.Gravity.CENTER);
        splashLayout.addView(subtitle);

        progressBar = new ProgressBar(this);
        progressBar.setPadding(0, 48, 0, 16);
        splashLayout.addView(progressBar);

        statusText = new TextView(this);
        statusText.setText("Starting swarm…");
        statusText.setTextColor(Color.parseColor("#64748b"));
        statusText.setTextSize(12);
        statusText.setGravity(android.view.Gravity.CENTER);
        splashLayout.addView(statusText);

        // WebView (hidden until hub is ready)
        webView = new WebView(this);
        webView.setLayoutParams(fill);
        webView.setVisibility(View.GONE);

        WebSettings ws = webView.getSettings();
        ws.setJavaScriptEnabled(true);
        ws.setDomStorageEnabled(true);
        ws.setLoadWithOverviewMode(true);
        ws.setUseWideViewPort(true);
        ws.setBuiltInZoomControls(false);
        ws.setDisplayZoomControls(false);
        ws.setSupportZoom(false);
        ws.setCacheMode(WebSettings.LOAD_DEFAULT);

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest req) {
                // Keep all navigation inside the WebView (localhost only)
                return false;
            }
        });

        root.addView(splashLayout);
        root.addView(webView);
        setContentView(root);
    }

    // ── Hub readiness polling ─────────────────────────────────────────────────

    private void waitForHub() {
        executor.submit(() -> {
            long deadline = System.currentTimeMillis() + MAX_WAIT_MS;
            int attempt = 0;
            while (System.currentTimeMillis() < deadline) {
                attempt++;
                if (isHubReady()) {
                    mainHandler.post(this::showWebView);
                    return;
                }
                String msg = "Starting agents… (" + attempt + ")";
                mainHandler.post(() -> statusText.setText(msg));
                try {
                    Thread.sleep(POLL_INTERVAL_MS);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    return;
                }
            }
            // Timed out — try to load anyway, WebView will show error if truly offline
            mainHandler.post(this::showWebView);
        });
    }

    private boolean isHubReady() {
        try {
            HttpURLConnection conn = (HttpURLConnection) new URL(HUB_URL + "/status").openConnection();
            conn.setConnectTimeout(500);
            conn.setReadTimeout(500);
            conn.setRequestMethod("GET");
            int code = conn.getResponseCode();
            conn.disconnect();
            return code == 200;
        } catch (IOException e) {
            return false;
        }
    }

    private void showWebView() {
        splashLayout.setVisibility(View.GONE);
        webView.setVisibility(View.VISIBLE);
        webView.loadUrl(HUB_URL);
    }
}
