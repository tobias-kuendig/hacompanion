# hadaemon
Daemon that sends local hardware information to Home Assistant 

## Notifications

```yaml
service: notify.mobile_app_your_device
data: 
  title: "Message Title"
  message: "Message Body"
  data:
    expire: 4000 # display for 4 seconds
    urgency: normal
```
