## Proxy list:
- **httpobfs:** http camouflage for tcp network. example:
```
glider -verbose -listen :1080 -forward httpobfs://xxx.xxx.xxx.xx:443/?host=google.com,vmess://:794ae901-cc7e-4ca7-a7fc-8cf68acea186@?alterID=0 -dialtimeout 10
``` 

## Usage:
Add

```
import _ "github.com/doorbash/glider-custom-proxies"
```

To 

```
github.com/nadoo/glider/feature.go
```

