package com.bodhi.hub;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.util.Log;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;

public class BodhiService extends Service {
    private static final String TAG = "BodhiService";
    private static final String CHANNEL_ID = "bodhi_channel";
    private static final int NOTIF_ID = 1;

    private Process hubProcess;

    @Override
    public void onCreate() {
        super.onCreate();
        createNotificationChannel();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        startForeground(NOTIF_ID, buildNotification());
        new Thread(this::startHub).start();
        return START_STICKY;
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        if (hubProcess != null) {
            hubProcess.destroy();
        }
    }

    private void startHub() {
        try {
            File binary = extractBinary();
            if (binary == null) {
                Log.e(TAG, "Failed to extract binary for this ABI");
                return;
            }

            File memoryDir = new File(getFilesDir(), "bodhi");
            memoryDir.mkdirs();

            ProcessBuilder pb = new ProcessBuilder(binary.getAbsolutePath());
            pb.directory(memoryDir);
            pb.environment().put("PORT", "8080");
            pb.environment().put("MEMORY_FILE", new File(memoryDir, ".bodhi-memory.json").getAbsolutePath());
            pb.redirectErrorStream(true);

            Log.i(TAG, "Starting Bodhi Hub at " + binary.getAbsolutePath());
            hubProcess = pb.start();

            // Drain stdout/stderr so the process doesn't block
            new Thread(() -> {
                try {
                    byte[] buf = new byte[1024];
                    while (hubProcess.getInputStream().read(buf) != -1) {
                        // Log.d(TAG, new String(buf).trim());
                    }
                } catch (IOException ignored) {}
            }).start();

            hubProcess.waitFor();
            Log.i(TAG, "Bodhi Hub process exited");
        } catch (Exception e) {
            Log.e(TAG, "Error starting Bodhi Hub", e);
        }
    }

    private File extractBinary() {
        String abi = Build.SUPPORTED_ABIS[0];
        String assetDir;
        if (abi.startsWith("arm64")) {
            assetDir = "arm64-v8a";
        } else if (abi.startsWith("armeabi")) {
            assetDir = "armeabi-v7a";
        } else if (abi.startsWith("x86_64")) {
            assetDir = "x86_64";
        } else {
            Log.e(TAG, "Unsupported ABI: " + abi);
            return null;
        }

        String assetPath = assetDir + "/bodhi-hub";
        File dest = new File(getFilesDir(), "bodhi-hub");

        // Re-extract if missing or app was updated (check version)
        File versionFile = new File(getFilesDir(), "bodhi-hub.version");
        String currentVersion = BuildConfig.VERSION_NAME;
        try {
            if (versionFile.exists()) {
                byte[] vbuf = new byte[32];
                int len;
                try (InputStream vis = new java.io.FileInputStream(versionFile)) {
                    len = vis.read(vbuf);
                }
                if (dest.exists() && new String(vbuf, 0, len).trim().equals(currentVersion)) {
                    return dest; // already extracted for this version
                }
            }
        } catch (IOException ignored) {}

        // Extract from assets
        try (InputStream in = getAssets().open(assetPath);
             FileOutputStream out = new FileOutputStream(dest)) {
            byte[] buf = new byte[65536];
            int n;
            while ((n = in.read(buf)) != -1) {
                out.write(buf, 0, n);
            }
        } catch (IOException e) {
            Log.e(TAG, "Failed to extract binary: " + assetPath, e);
            return null;
        }

        dest.setExecutable(true, true);

        // Save version stamp
        try (FileOutputStream vout = new FileOutputStream(versionFile)) {
            vout.write(currentVersion.getBytes());
        } catch (IOException ignored) {}

        Log.i(TAG, "Extracted " + assetPath + " → " + dest.getAbsolutePath());
        return dest;
    }

    private Notification buildNotification() {
        Intent tapIntent = new Intent(this, MainActivity.class);
        PendingIntent pi = PendingIntent.getActivity(this, 0, tapIntent,
                PendingIntent.FLAG_IMMUTABLE | PendingIntent.FLAG_UPDATE_CURRENT);

        Notification.Builder builder = new Notification.Builder(this, CHANNEL_ID)
                .setContentTitle("Bodhi Hub")
                .setContentText("35 agents running · tap to open")
                .setSmallIcon(android.R.drawable.ic_dialog_info)
                .setContentIntent(pi)
                .setOngoing(true);

        return builder.build();
    }

    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(
                    CHANNEL_ID, "Bodhi Hub Service",
                    NotificationManager.IMPORTANCE_LOW);
            channel.setDescription("Keeps the Bodhi agent swarm running in the background");
            NotificationManager nm = getSystemService(NotificationManager.class);
            nm.createNotificationChannel(channel);
        }
    }
}
