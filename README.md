# sdk-auto

## Build
### Local Build
> make build-sdk

### Docker Build
> docker build -t sdk-auto:latest .

## Run
### 参数
* trace_api - 配置OTEL后端接收地址，eg. jaegerip:4318
* scan_interval - 配置定时扫描新进程间隔（秒），将Go应用自动Instrument，暂不支持Java
* black_list - 进程黑名单列表，使用模糊匹配。Go应用匹配Comm文件内容，Java应用匹配CmdLine文件内容。不在黑名单列表的应用都会被改造

### Local Run
> ./originx-sdk-auto

### K8S Run
> kubectl apply -f sdk-auto-deploy.yml
