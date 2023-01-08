## Proxy list:
- **httpobfs:** http camouflage for tcp network. example:
```sh
glider -verbose -listen :1080 -forward httpobfs://xxx.xxx.xxx.xx:443/?host=google.com,vmess://:794ae901-cc7e-4ca7-a7fc-8cf68acea186@?alterID=0 -dialtimeout 10
``` 

- **doh:** dns over https. example:
```sh
glider -verbose -listen :1080 -forward http://127.0.0.1:10809,doh://1.1.1.1
glider -verbose -listen udp://0.0.0.0:53 -forward http://127.0.0.1:10809,doh://1.1.1.1,udp://8.8.8.8:53
```

## Install:
Add

```golang
import _ "github.com/doorbash/glider-custom-proxies"
```

To 

```
github.com/nadoo/glider/feature.go
```

