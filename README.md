# Desktop Companion for Home Assistant

[![Test and build](https://github.com/tobias-kuendig/hacompanion/actions/workflows/build.yml/badge.svg)](https://github.com/tobias-kuendig/hacompanion/actions/workflows/build.yml)

This is an unofficial Desktop Companion App for Home Assistant written in Go.

The companion is running as a background process and sends local hardware information 
to your Home Assistant instance.

Currently, **Linux** is the only supported operating system (tested on Ubuntu 20.04 / KDE Neon)

## Supported sensors

* CPU Temperature
* CPU Usage
* Load average
* Memory Usage
* Uptime
* Power stats
* Online check
* Audio volume
* Webcam Process Count
* Custom scripts

## Getting started

1. Download a copy of the [configuration file](hacompanion.toml). Save it to `~/.config/hacompanion.toml`.
1. In Home Assistant, generate a token by visting [your profile page](https://www.home-assistant.io/docs/authentication/#your-account-profile), then click on `Generate Token` at the end of the page.
1. Update your `~/.config/hacompanion.toml` file's `[homeassistant]` section with the generated `token`.
1. Set the display name of your device (`device_name`) and the URL of your Home Assistant instance (`host`).
1. If you plan to receive notifications from Home Assistant on your Desktop, change the `push_url` setting under `[notifications]` to point to your local IP address. 
1. Configure all sensors in the configuration file as you see fit.
1. Run the companion by executing `./hacompanion` (use the `-config=/path/to/config` flag to pass the path to a custom configuration file, `~/.config/hacompanion.toml` is used by default).

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

Start the companion by running:

```bash
sudo service hacompanion start
# check status with
# sudo service hacompanion status
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

