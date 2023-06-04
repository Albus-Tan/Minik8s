# Serverless

借鉴 Config Map 的思想设计

# Serverless v1

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

# Serverless v2

## API

**Function Template**

函数模板，类似于正常 ApiObject，可通过以下URL对模板进行增删改查，支持 Watch 方法

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
	Name           string `json:"name"`
	// ServiceAddress is address where service of this func template is
	// redirect http request to this address when this func is called 
	ServiceAddress string `json:"serviceAddress"`

	PreRun   string `json:"preRun"`
	Function string `json:"function"`
	Left     string `json:"left"`
	Right    string `json:"right"`
    
	// InitInstanceNum means how many pod instance will be instantly
	// created when func template is created
	InitInstanceNum *int `json:"initInstanceNum,omitempty"`
	// MaxInstanceNum means how many pod instance will be
	// created max for this func template
	MaxInstanceNum *int `json:"maxInstanceNum,omitempty"`
	// MinInstanceNum means how many pod instance will be
	// created min for this func template
	MinInstanceNum *int `json:"minInstanceNum,omitempty"`
}

type FuncStatus struct {
	// ServiceUID is the uid of service the function template related
	// Forward request to pods managed by rs
	ServiceUID types.UID `json:"serviceId,omitempty"`
	// ReplicaSetUID is the uid of rs the function template related, which
	// is responsible for managing all fun server pods life cycle, by changing
	// spec.replicas in it
	ReplicaSetUID types.UID `json:"replicaSetId,omitempty"`
	// Counter is used to count call request number of this func
	// to decide the pod should have
	Counter int `json:"counter,omitempty"`
	// TimeStamp record the time last this func template is called
	TimeStamp types.Time `json:"timeStamp,omitempty"`
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
    - 将 HTTP 请求直接转发给对应 service，service 会将请求转发给对应 label match 的 pod，label 部分由用户告知 apiserver 创建 func template 时，处理函数自动根据 func 的名字生成（同时也会生成对应 func server 的 pod 模板）
    - pod 中会在名为 name 函数逻辑结束后，调用下一个函数（同样是 `PUT /api/funcs/:name/:id` 接口，name 字段设置为下一个将被调用的函数名即可），并将用户调用的实例 id 即 `instanceId` 递归传递
    - pod 个数由所属 func template 中的 replica set 管理
      - 每当出现新的对函数的调用请求，更新对应 func template 中 status 里的 timestamp 时间戳，同时增加 counter
      - 每隔一定时间，将所有现存的 func template 中的 counter 统一减少一定数值，同时 replicaset 中的 replica num 保持与 counter 一致，从而实现函数不被调用时 scale to 中 zero
      - 设置策略限定 counter 上界，同时优化 counter 不同时的扩缩策略，实现更佳效果