# kratos-openapiv3-api

## Quick Start

在项目中引入openapiv3

```go
import	openapiHandler github.com/web-xiaxia/kratos-openapiv3-api

h := openapiHandler.NewHandler()
//将/q/路由放在最前匹配
httpSrv.HandlePrefix("/q/", h)
```

```shell
curl /q/services

curl /q/service/{serviceId}

curl /q/service/group/{servicePrefix}

```