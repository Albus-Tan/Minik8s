# Serverless

## API

**Function Template**

函数模板，类似于正常 ApiObject，可通过以下URL对模板进行增删改查，不支持 Watch 方法

```sh
/api/funcs/template
/api/funcs/template/:name	# name 为函数模板中 Spec 中的 Name
```

对应 Func 类型 ApiObject

```go
type Func struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec            FuncSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status          FuncStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type FuncSpec struct {
	// Name is unique for Func, should be same as Name in ObjectMeta field
	Name string `json:"name"`

	PreRun   string `json:"preRun"`
	Function string `json:"function"`
	Left     string `json:"left"`
	Right    string `json:"right"`
}

// SHOULD NOT used in func template
type FuncStatus struct {
	InstanceId types.UID `json:"instanceId,omitempty"`
}
```

**Function Instance and Call**

真正进行函数调用

**用户接口**

- `POST /api/funcs/:name` （body 部分进行参数传递）
  - 依据名为 name 的函数模板创建并运行实例，返回实例 id 即 `instanceId`
    - 会生成本次调用的实例 id 即 `instanceId` 并返回
    - 调用内部接口 `PUT /api/funcs/:name/:id`
- `GET /api/funcs/:id`
  - 依据实例 id 即 `instanceId` 查看所调用函数返回的结果

**内部接口**

- `PUT /api/funcs/:name/:id`（body 部分进行参数传递）
  - 用户调用的实例 id 即 `instanceId` 参数用于识别是用户哪次实际调用中的调用流，也便于存储最终结果
  - 依据 name 字段调用对应函数
    - 如果 name 字段为 RETURN，则说明此次调用负责存储函数返回值
    - 创建相应 pod
    - pod 中会在名为 name 函数逻辑结束后，调用下一个函数（同样是 `PUT /api/funcs/:name/:id` 接口，name 字段设置为下一个将被调用的函数名即可），并将用户调用的实例 id 即 `instanceId` 递归传递；同时会自我调用当前 pod 的 delete 方法，自行析构

