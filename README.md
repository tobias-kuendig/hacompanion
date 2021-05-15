# Desktop Companion for Home Assistant

[![test](https://github.com/tobias-kuendig/hacompanion/actions/workflows/test.yml/badge.svg)](https://github.com/tobias-kuendig/hacompanion/actions/workflows/test.yml)
[![lint](https://github.com/tobias-kuendig/hacompanion/actions/workflows/lint.yml/badge.svg)](https://github.com/tobias-kuendig/hacompanion/actions/workflows/lint.yml)
[![goreleaser](https://github.com/tobias-kuendig/hacompanion/actions/workflows/release.yml/badge.svg)](https://github.com/tobias-kuendig/hacompanion/actions/workflows/release.yml)

This is an unofficial Desktop Companion app for Home Assistant.

The companion is running as a background process and sends local hardware information 
to your own Home Assistant instance.

Currently, **Linux** is the only supported operating system (tested on Ubuntu 20.04 / KDE Neon)

## Getting started

1. Download a copy of the [configuration file](companion.toml). Save it as `~/.config/hacompanion.toml`.
1. In Home Assistant, generate a token by visting your profile page, then click on `Generate Token` at the end of the page.
1. Paste the token into the `~/.config/hacompanion.toml` in the `[homeassistant]` section of the configu file.
1. Set the display name of your device (`device_name`) and the URL of your Home Assistant instance (`host`).
1. If you plan to receive notifications from Home Assistant on your Desktop, change the `notification_server` value to point to your local IP address. Optionally, set a different port that works for you.
1. Run the companion by executing `./hacompanion` (use the `-config=/path/to/config` flag to pass the path to a custom configuration file).

## Run the companion on system boot

If your system is using Systemd, you can use the following unit file to run the companion on system boot:

```ini
# sudo vi /etc/systemd/system/hacompanion.service
[Unit]
Description=Home Assistant Desktop Companion

[Service]
Type=simple
Restart=on-failure
RuntimeMaxSec=604800
ExecStart=/path/to/hacompanion -config=/home/yourname/.config/hacompanion.toml

[Install]
WantedBy=multi-user.target
```

Start the companion now by running

```bash
sudo service hacompanion start
# check status with
sudo service hacompanion status
```

## Receiving notifications

The companion can receive notifications from Home Assistant and display
them using `libnotify`. To test the integration, start the companion
and execute the following service in Home Assistant:

```yaml
service: notify.mobile_app_your_device # change this!
data: 
  title: "Message Title"
  message: "Message Body"
  data:
    expire: 4000 # display for 4 seconds
    urgency: normal
```

## Automation ideas

Feel free to share your automation ideas [in the Discussions section](https://github.com/tobias-kuendig/hacompanion/discussions) of this repo.

