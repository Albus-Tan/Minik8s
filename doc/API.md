# API

> API 标准参照 https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/

- API 接口规定为 [BindHandlers](../pkg/apiserver/httpserver.go) 函数中的注释及对应注册的 [api url](../pkg/api/url.go)
- API object 定义在 `/pkg/api` 路径下
