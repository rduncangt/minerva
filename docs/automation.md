# Automation and Scheduling

This document outlines the automation process used to ingest logs into Minerva, including how to compile the Go application, create a wrapper script, and schedule log ingestion with launchd on macOS.

## Overview

Minerva is configured to retrieve log data from a remote Raspberry Pi using SSH, process these logs, and insert relevant entries into the PostgreSQL database. Automation is achieved by:

- Compiling the Go application into a standalone binary.
- Creating a wrapper script that fetches logs via SSH.
- Scheduling the wrapper script to run via launchd on macOS (and optionally via systemd on Linux).

## Compiling the Go Application

1. Navigate to the application directory:

   ```bash
   cd cmd/minerva
   ```

2. Build the binary:

    ```bash
    go build -o minerva
    ```

3. Move the binary to a system-wide directory:

    ```bash
    sudo mv minerva /usr/local/bin/minerva
    ```

## Wrapper Script

During development, logs were piped directly into the Go application using:

```bash
ssh secpi "cat /var/log/syslog" | go run cmd/minerva/main.go
```

For automation, create a wrapper script (`/usr/local/bin/minerva-run.sh`) to replicate this behavior:

```bash
#!/bin/bash
# Connect to the 'secpi' machine, output syslog data, and pipe it into the Minerva binary.
ssh secpi "cat /var/log/syslog" | /usr/local/bin/minerva
```

Make the script executable:

```bash
sudo chmod +x /usr/local/bin/minerva-run.sh
```

## Launchd Setup on macOS

Create a launchd plist file at `~/Library/LaunchAgents/com.minerva.plist` to schedule the job. For example, to run the job twice per day (at 7:00 AM and 7:00 PM):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>com.minerva</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/minerva-run.sh</string>
    </array>
    <key>StartCalendarInterval</key>
    <array>
      <dict>
        <key>Hour</key>
        <integer>7</integer>
        <key>Minute</key>
        <integer>0</integer>
      </dict>
      <dict>
        <key>Hour</key>
        <integer>19</integer>
        <key>Minute</key>
        <integer>0</integer>
      </dict>
    </array>
    <key>StandardOutPath</key>
    <string>/tmp/minerva.out</string>
    <key>StandardErrorPath</key>
    <string>/tmp/minerva.err</string>
  </dict>
</plist>
```

## Loading the Launch Agent

After saving the plist file, load it with:

```bash
launchctl load ~/Library/LaunchAgents/com.minerva.plist
```

If you encounter errors, use:

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.minerva.plist
```

Verify that the job is loaded:

```bash
launchctl list | grep com.minerva
```

## Troubleshooting

- Plist Syntax Issues:
    Validate with:

```bash
plutil -lint ~/Library/LaunchAgents/com.minerva.plist
```

- Permissions:
    Ensure that the plist and the wrapper script have the correct permissions.
- Output Logs:
    Check `/tmp/minerva.out` and `/tmp/minerva.err` for runtime messages.

## Summary

This automation setup allows Minerva to automatically fetch and process logs from the remote Raspberry Pi at scheduled intervals. Adjust the scheduling in the plist file as needed for your operational requirements.
