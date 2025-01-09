# Configuring AT&T Router Log Export for Minerva

This guide explains how to configure an AT&T router to export logs to the Raspberry Pi running Minerva.

---

## Step 1: Access Router Settings

1. Open a web browser and navigate to the router’s admin page. Typically:

   ```
   http://192.168.1.254
   ```

2. Log in with the router credentials. The default credentials are usually printed on the router or in its documentation.

---

## Step 2: Enable Syslog Export

1. Navigate to the `Diagnostics` or `Advanced Settings` section of the router admin panel.
2. Look for a `Syslog` or `Log Export` option.
3. Configure the following settings:
   - **Syslog Server IP**: Enter the IP address of your Raspberry Pi. Example:

     ```
     192.168.1.84
     ```

   - **Syslog Port**: Use the default syslog port:

     ```
     514
     ```

4. Save the settings.

---

## Step 3: Verify Log Export

1. SSH into the Raspberry Pi:

   ```bash
   ssh secpi
   ```

2. Check if logs are being received in the syslog file:

   ```bash
   tail -f /var/log/syslog
   ```

   You should see entries from the router. For example:

   ```
   Jan 8 00:01:08 dsldevice.attlocal.net L4 FIREWALL[7567]: action=DROP ...
   ```

---

## Troubleshooting

### Logs Not Appearing

1. Verify the Raspberry Pi’s IP address matches the `Syslog Server IP` configured on the router.
2. Check if the syslog service is running on the Raspberry Pi:

   ```bash
   systemctl status rsyslog
   ```

3. Ensure no firewall on the Raspberry Pi blocks port `514`.

   ```bash
   sudo ufw status
   ```

   If necessary, allow syslog traffic:

   ```bash
   sudo ufw allow 514/udp
   ```

---

By following this guide, your AT&T router should successfully export logs to the Raspberry Pi for processing by Minerva.
