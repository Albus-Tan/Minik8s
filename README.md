# Minik8s

Minik8s 是一个类似于 [Kubernetes](https://kubernetes.io/) 的迷你容器编排工具，能够在多机上对满足 CRI 接口的容器进行管理，支持容器生命周期管理、动态伸缩、自动扩容等基本功能，并且基于 minik8s 实现了 Serverless 平台集成。

## 目录

[TOC]

## 快速开始

### 安装依赖

- [flannel](###Flannel): 多机容器统一网络抽象
- docker: 容器管理
- [etcd](###etcd): 存储
- ipvsadm: service 底层 nat 实现
- iproute2: 虚拟 ip 创建
- [cadvisor](###Cadvisor): 获取容器运行状态

### 配置网络环境

#### 启用转发

```bash
# 在所有机器上运行
sysctl -w sysctl net.ipv4.ip_forward=1
```

#### 配置虚拟网卡

```bash
# 在所有机器上运行
ip l a minik8s-proxy0 dev dummy
ip l s minik8s-proxy0 up
````

### 配置 flannel

由于 flannel 和 minik8s 均需要使用 etcd 作为存储组件，建议使用两个 etcd 实例分别存储。

我们以三个节点为例，node1(192.168.1.1/24), node2(192.168.1.2/24), node3(192.168.1.3/24)，这里将在 node1 上为 flannel 配置存储所用的 etcd。

**node1:**

```bash
# 由于 etcd 和 flannel 实际作为服务启动，可能需要使用类似 tmux/screen 的程序托管
# 启动 etcd
etcd --listen-client-urls="http://192.168.1.1:2379" --advertise-client-urls="http://192.168.1.1:2379"
```

**node1, node2, node3:**

```bash
# 启动 flannel
flannel --etcd-endpoints=http://192.168.1.13:2379 --iface=192.168.1.13 --ip-masq=true --etcd-prefix=/coreos.com/network
```
同时我们需要让 docker 使用 flannel 的网络环境
```bash
# 修改 systemd docker 参数
# 修改后如下
# cat /lib/systemd/system/docker.service 
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service containerd.service
Wants=network-online.target
Requires=docker.socket containerd.service

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
EnvironmentFile=/run/docker_opts.env
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock $DOCKER_OPTS
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always

# Note that StartLimit* options were moved from "Service" to "Unit" in systemd 229.
# Both the old, and new location are accepted by systemd 229 and up, so using the old location
# to make them work for either version of systemd.
StartLimitBurst=3

# Note that StartLimitInterval was renamed to StartLimitIntervalSec in systemd 230.
# Both the old, and new name are accepted by systemd 230 and up, so using the old name to make
# this option work for either version of systemd.
StartLimitInterval=60s

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Comment TasksMax if your systemd version does not support it.
# Only systemd 226 and above support this option.
TasksMax=infinity

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process
OOMScoreAdjust=-500

[Install]
WantedBy=multi-user.target

```

```bash
# 创建 docker ops 文件
# 根据 flannel 创建的 网络配置文件修改 docker opts 文件
source /run/flannel/subnet.env
echo "DOCKER_OPTS=\" --bip=$FLANNEL_SUBNET --ip-masq=$FLANNEL_IPMASQ --mtu=$FLANNEL_MTU\"" > /run/docker_opts.env
```

### 启动 Minik8s

node2 将作为集群的 Master 节点。

假设项目被 clone 到了 `root/minik8s` 下，cadvisor 二进制文件在 `root/minik8s` 下

**node2:**

```bash
export API_SERVER=192.168.1.2
export PORT=8080
export NODE_CONFIG=/root/minik8s/examples/dash/node/worker2.yaml
make clean && make
# start cadvisor to use hpa
./cadvisor  -port=8090 &>>/var/log/cadvisor.log
NODE_CONFIG=/root/minik8s/examples/dash/node/master.yaml; ./build/master
./build/kube-proxy
./build/kubelet
```

**node1:**

```bash
export API_SERVER=192.168.1.2
export PORT=8080
export NODE_CONFIG=/root/minik8s/examples/dash/node/worker1.yaml
# start cadvisor to use hpa
./cadvisor  -port=8090 &>>/var/log/cadvisor.log
./build/kube-proxy
./build/kubelet
```

**node3:**

```bash
export API_SERVER=192.168.1.2
export PORT=8080
export NODE_CONFIG=/root/minik8s/examples/dash/node/worker3.yaml
# start cadvisor to use hpa
./cadvisor  -port=8090 &>>/var/log/cadvisor.log
./build/kube-proxy
./build/kubelet
```

### Minik8s Kubectl

Kubectl 是用于运行 Minik8s 集群命令的管理工具。本部分概述涵盖了 kubectl 语法，对命令操作的描述，并列举的常见例子。Kubectl 工具应当在控制面所在物理节点上被使用。

**安装和编译**

编译整个项目时会同时编译 Kubectl 命令行工具：

```bash
make clean && make
```

或单独编译：

```sh
go build minik8s/cmd/kubectl
```

**基本语法**

进入到项目目录 `/minik8s` 下，打开 bash 命令行：

如果通过 make 脚本编译：

```sh
./build/kubectl [command] [TYPE] [NAME] [flags]
```

单独编译：

```sh
./kubectl [command] [TYPE] [NAME] [flags]
```

- `command`：指定要在一个或多个资源执行的操作，例如操作 `create`，`get`，`describe`，`delete`
- `TYPE`：指定资源类型[Resource types](##Resources 资源概览)。区分大小写，也可以指定单数，复数或缩写的形式
  - 例如，以下命令将输出相同的结果：`$ kubectl get pods d022d439-fc71-4bd7-820e-f1cf21f9567a`，`$ kubectl get pod d022d439-fc71-4bd7-820e-f1cf21f9567a`
- `NAME`：指定 Resource 的唯一标识名称。对于 Func 类型的资源，即函数的 Name，对于其余资源 `NAME` 均指代创建该资源后所返回的 `UID`。如果省略 Name，则显示所有该类型资源的信息，例如 `$ kubectl get pods`
- `flags`：指定可选 flags

**Operations**

下表包括了所有 kubectl 操作简短描述和通用语法：

| Operation | Syntax                                             | Description    |
|-----------|----------------------------------------------------|----------------|
| apply     | kubectl apply [TYPE] -f FILENAME [flags]           | 从文件创建资源        |
| create    | kubectl create [TYPE] -f FILENAME [flags]          | 从文件创建资源        |
| delete    | kubectl del [TYPE] [NAME] [flags]                  | 删除资源           |
| describe  | kubectl describe [TYPE] ([NAME]) [flags]           | 显示一个或所有资源的详细状态 |
| get       | kubectl get [TYPE] ([NAME]) [flags]                | 列出一个或所有资源的简略状态 |
| update    | kubectl update [TYPE] ([NAME]) -f FILENAME [flags] | 从文件更改资源        |
| clear     | kubectl clear                                      | 清空现有所有资源       |
| help      | kubectl --help/-h                                  | 帮助信息           |

`FILENAME`：其中文件支持 `yaml` 和 `json` 格式的配置文件

**示例**

```bash
# 依据 network-test.yaml 配置文件创建 Pod
kubectl create pod -f examples/dash/pod/network-test.yaml
# 获取现存所有 Pod 简略信息
kubectl get pod
# 获取 UID 为 d022d439-fc71-4bd7-820e-f1cf21f9567a 的 Pod 简略信息
kubectl get pod d022d439-fc71-4bd7-820e-f1cf21f9567a
# 获取现存所有 Pod 详细信息
kubectl describe pod
# 删除 UID 为 d022d439-fc71-4bd7-820e-f1cf21f9567a 的 Pod
kubectl del pod d022d439-fc71-4bd7-820e-f1cf21f9567a
```

## 总体架构

Minik8s 的总体架构整体上参考了课上所提供的 minik8s best practice 的架构，主要分为控制面 Master 和工作节点 Worker 两部分。

### 架构图

![minik8s-framework.drawio](README.assets/minik8s-framework.drawio.svg)

### 组件

**核心组件**

- 控制面 Master
  - ApiServer：负责与各组件交互，将 API 对象持久化进入 etcd
  - Scheduler：负责新创建的 Pod 的调度
  - ControllerManager：负责管理各个 Controller
    - ReplicaSetController：负责实现并管理 ReplicaSet
    - HorizontalController：负责实现并管理 HPA
    - DnsController：负责实现并管理 Dns
    - ServerlessController：负责实现 Serverless 的函数调用及实例管理等
    - PodController：管理 Pod 生命周期，负责实现 Pod RestartPolicy
  - GpuServer：管理 Gpu Job
- 工作节点 Worker
  - Kubelet：在每个节点上控制管理 Pod 生命周期
  - Kubeproxy：配置节点网络，实现统一网络抽象
- 其他
  - Kubectl：命令行工具，用于与控制面交互
  - ApiClient：能够与 ApiServer 交互通信的 Client

**组件概览**

- **控制面 Master**
  - ApiServer：暴露 API，负责处理用户/各组件的 HTTP 请求，并将 API 对象持久化进入 etcd
    - HttpServer：负责接收用户/组件的 HTTP 请求
      - Handlers：负责调用 EtcdClient 处理用户/其他组件对 API 对象的增删改查操作
      - ServerlessFuncHandler：负责转发用户 Call 函数的请求到具体运行的函数实例，并将函数执行结果反馈给用户；同时负责实现 Serverless 的函数递归调用
    - EtcdClient：与 etcd 直接交互的 client，负责增删改查，并提供 watch 监听机制转发
  - Scheduler：负责新创建的 Pod 的调度
  - ControllerManager：负责管理各个 Controller，监控整个集群的状态，确保集群中的真实状态和期望状态一致
    - Controller 基本组件
      - Informer：Controller 本地的数据缓存，将 Object 的数据缓存在本地，只监听更新并及时同步到本地cache中
        - Reflector：负责监听更新 Object
        - ThreadSafeStore：负责 Object 的存储，线程安全
      - Workqueue：包含 Object 变化的事件，Controller 可以通过启动工作线程（可以并行处理）在 workqueue 中获取需要处理的对象并操作
    - PodController：管理 Pod 生命周期，负责实现 Pod RestartPolicy
    - ReplicaSetController：管理实现 ReplicaSet 功能，保证期望数量，符合 selector 条件的 Pod 实例在正常运行
    - HorizontalController：管理实现 HPA 自动扩缩容功能，通过获取各节点上资源占用进行有关决策
      - MetricsClient：聚合各节点上的资源占用（如依据一类 Pod 进行聚合）
        - CadvisorClient：与节点上的 cadvisor 进行交互，获取资源占用信息
    - DnsController：管理实现 Dns 与 Http 请求转发功能
    - ServerlessController：负责实现 Serverless 函数的实例生命周期管理等
  - GpuServer：负责提交 Gpu 任务至云平台，根据配置进行脚本生成，以及下载反馈结果
    - JobClient：负责通过 ssh 与云平台进行交互
  - HeartbeatWatcher：负责监听工作节点的 heartbeat，并对状态异常的工作节点进行处理
- **工作节点 Worker**
  - HeartbeatSender：负责发送当前工作节点的 heartbeat 给 Master 控制面，告知自身状态正常
  - Kubelet：每个从节点 Node 的管理者，与主节点交互，控制管理 Pod 生命周期
    - PodManager
    - CriClient
  - Kubeproxy：管理节点网络，专门负责容器网络的部分，以及 Node 间连接，实现并管理 Service
    - ServiceManager：负责实现并管理 Service
    - IpvsClient
- 其他
  - Kubectl：命令行工具，用于与控制面交互
  - NodeManager：负责 master 及 worker 节点 node 抽象的初始化及删除等工作
  - ApiClient：能够与 ApiServer 交互通信的 Client
    - RESTClient：能够与 ApiServer 交互通信的 REST Client
    - ListerWatcher：专门负责 List 与 Watch 两个操作
  - Logger：日志管理与打印

### 软件栈

- 控制面 Master
  - ApiServer
    - uuid: https://github.com/google/uuid
    - gin: https://github.com/gin-gonic/gin
    - etcd: https://github.com/etcd-io/etcd
  - ControllerManager
    - HorizontalController
      - cadvisor: https://github.com/google/cadvisor
    - DnsController
    - ServerlessController
  - GpuServer
    - goph: https://github.com/melbahja/goph
    - sftp: https://github.com/pkg/sftp
- 工作节点 Worker
  - Kubelet
    - docker: https://github.com/moby/moby
  - Kubeproxy
    - ipvs: https://github.com/moby/ipvs
    - net: https://golang.org/x/net
- 其他
  - Kubectl
    - cobra: https://github.com/spf13/cobra
    - viper: https://github.com/spf13/viper
    - yaml: https://gopkg.in/yaml.v3

## 项目信息

gitee目录地址：https://gitee.com/albus-tan/minik8s

主要编程语言：`go 1.18`，`python`，`shell`

### 开发规范

#### 分支介绍

采用 [Vincent Driessen](https://nvie.com/posts/a-successful-git-branching-model/) 提出的 git branch model 进行分支管理，主要分支包括：

- `master` 分支：提供给用户使用的正式版本和稳定版本，所有版本发布和 Tag 操作都在这里进行。不允许开发者日常 push，允许从 `develop` 合并
-  `develop` 分支：日常开发的汇总分支。开发者可以检出 `feat` 和 `fix` 分支，开发完成后提出 pull request，经过 peer review 后被合并回 `develop`。不允许开发者日常直接 push，只允许完成功能开发或 bug 修复后通过 pull request 进行合并
- `feat` 分支：从 `develop` 分支检出，用于新功能开发。开发完毕，经过测试后通过 pull request 合并到 `develop` 分支，允许开发者日常 push
  - 命名为 `feat/component/detail`，如 `feat/apiserver/handlers`，表示对于 ApiServer 组件的 handlers 功能的开发分支
- `fix` 分支：从 `develop` 分支检出，用于 bug 修复（feat 过程中的 bug 直接就地解决）；修复完毕，经过测试后合并到 `develop` 分支，允许开发者日常 push
  - 命名为 `fix/component/detail`，如 `fix/etcd/endpoint_config` ，表示对于 etcd 开发时的 endpoint 配置的修复

**分支概览**

<img src="README.assets/image-20230526174458952.png" alt="image-20230526174458952" style="zoom:33%;" />

#### Commit Message 规范

```
<type>: <body>
```

type 有下面几类

- `feat` 新功能
- `fix` 修补bug（在 `<body>` 里面加对应的 Issue ID）
- `test` 测试相关
- `doc` 注释/文档变化
- `refactor` 重构（没有新增功能或修复 BUG）

##### 规范自动检查

提交后会通过 `.githooks/commit-msg` 的脚本自动检查规范

#### 项目新功能开发流程

每当需要开发新功能时，小组成员会在开会讨论后，由负责功能开发的成员新建 `feat` 分支，命名为 `feat/component/detail`，并在其上进行开发。在开发完毕后负责开发的成员会提出 pull request，待至少另一名小组成员完成 peer review 审查通过后，方可 merge 进 `develop` 分支。

**部分开发分支流**

![image-20230526181123018](README.assets/image-20230526181123018.png)

### CI/CD介绍

项目使用了 gitlab ci/cd 接口，将该仓库推送至您的 gitlab 仓库中，即可使用 gitlab 提供的自动化构建与测试（[配置文件](./.gitlab-ci.yml)）。

CI/CD 中配置了 vet check 和单元测试，同时检查 build 情况，并提供自动构建脚本。

### 软件测试方法介绍

#### 自动化测试

自动化测试通过撰写**测试脚本**后 **`go test`** 进行测试。主要包括针对小型组件（如 `client`）和函数（如 `ParseQuantity`）的**单元测试**，旨在测试组件是否依照需求工作，同时及时发现代码错误，协助开发。

此部分遵循 go 语言测试的基本原则，要求以 `*_test.go` 新建文件，并在文件中以 `TestXxx` 命名函数。然后再通过 `go test [flags] [packages]` 执行函数。

- 注意部分 `*_test.go` 文件中测试依赖于函数顺序，因此在 `go test` 时不能开并行测试，也不能调换测试函数在文件中出现的先后顺序。

#### 手动测试

手动测试通过操控 `postman` 或 `kubectl` 命令行工具，以及 `example` 文件夹下撰写的 `yaml/json` 案例进行测试。主要包括针对复杂逻辑，组件间交互进行的**集成测试**和**系统测试**。

采用手动测试的核心原因是本组人手不足，为每一种功能都撰写自动化测试脚本需要较大的额外工作量。但是在开发过程中以及验收前我们对所有需求都进行了详细充分的手动测试，包括一系列可能的边界情况，能够保证所开发项目代码的质量。

## Resources 资源概览

下表列出了所有支持的资源类型 Kind 及其缩写。其中 Resource Kind 为对应资源 yaml 文件中的 Kind 字段，Resource type 以及 Abbreviated alias 对应 Kubectl `TYPE` 字段：

| Resource Kind           | Resource type | Abbreviated alias |
| ----------------------- | ------------- | ----------------- |
| Node                    | nodes         | node              |
| Pod                     | pods          | pod               |
| ReplicaSet              | replicasets   | replicaset, rs    |
| Service                 | services      | service, svc      |
| HorizontalPodAutoscaler | hpas          | hpa               |
| Func                    | funcs         | func, f           |
| Job                     | jobs          | job, j            |
| DNS                     | dns           | -                 |

## 功能介绍

### Pod 抽象及容器生命周期管理

支持 Pod 抽象，能够根据用户指令对 Pod 的生命周期进行管理，包括控制 Pod 的启动和终止。如果 Pod 中的容器发生崩溃或自行终止，Minik8s 会将 Pod 重启。用户可以通过 `kubectl get pod`，`kubectl describe pod` 等指令获得 Pod 的运行状态。

支持创建包含多容器 Pod，可指定容器的镜像名称与版本、容器镜像所执行的命令、对容器资源用量的限制、容器所暴露的端口等。同一个 Pod 的多个容器间可以利用 localhost 进行相互通信。支持通过 volume 实现同一个 Pod 内的多个容器文件共享。

Pod 抽象可以通过类型为 Pod 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
    tier: frontend
  name: succeed-failure
  namespace: default
spec:
  containers:
    - image: lwsg/notice-server
      imagePullPolicy: PullIfNotPresent
      name: notice-server
      ports:
        - containerPort: 80
          protocol: TCP
      env:
        - name: _NOTICE
          value: 1
    - image: "ubuntu:bionic"
      imagePullPolicy: PullIfNotPresent
      name: timer
      command:
        - sleep
      args:
        - 30s
      resources:
        limits:
          cpu: 100m
          memory: 200M
  restartPolicy: RestartOnFailure
```

通过 volume 实现同一个 Pod 内的多个容器文件共享配置案例如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
    tier: frontend
  name: arg-volume-test
  namespace: default
spec:
  containers:
    - image: lwsg/debug-server
      imagePullPolicy: PullIfNotPresent
      name: debug-server-write
      volumeMounts:
        - name: share
          mountPath: "/share"
    - image: lwsg/debug-server
      imagePullPolicy: PullIfNotPresent
      name: debug-server-read
      volumeMounts:
        - name: share
          mountPath: "/share"
  restartPolicy: Never
```

#### Pod 间通信

通过 CNI 支持 Pod 间通信。在 Pod 启动时能为 Pod 分配独立的内网 IP，Pod 可以使用分配的 IP 与同节点或不同节点上的 Pod 通信。

### 多机 Minik8s

#### Node 抽象

Node 抽象通过 name 进行区分，需要保证不同物理实体机对应的 Node config 文件的 Name 字段在集群中全局唯一。Node 初始化时会检查当前有无 Node 与其重名，如果有，判断 config 文件是否与已有 Node 信息不同：

- 如果一致，则复用当前 Node，不再创建新 Node
- 如果不一致，报错给用户并退出；用户需要修改 config 文件的 Name 字段，或通过 put 方式修改原有 Node 的配置文件相关内容，以实现配置的修改

Node 抽象可以通过类型为 Node 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: v1
kind: Node
metadata:
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: node1
    kubernetes.io/os: linux
  name: node1
spec:
  podCIDR: 10.244.1.0/24
  podCIDRs:
  - 10.244.1.0/24
```

在节点上启动 Kubelet 即可自动注册 Worker 节点到集群中。启动 Master 节点主程序即可自动注册 Master 节点。也可以选择先手动注册节点（此时节点状态会显示为 `Pending`），之后当对应节点 Kubelet 启动后节点状态将自动更新为 `Running`。

##### Heartbeat 心跳机制

通过 worker 节点不断向 master 控制面发送 heartbeat，来告知控制面 worker 节点当前状态。如果一段时间 master 控制面没有接收到某一 worker 节点的心跳，就认为该节点异常，会将其删除。

**实现简述**

所有 worker 节点启动后，会由 Heartbeat Sender 持续向 Master 节点的 Heartbeat Watcher 发送心跳，一旦  Heartbeat Watcher 一段时间没有接收到 worker 节点发来的心跳，就认为对应 worker 节点挂掉，并将其信息在 etcd 内删除。

#### Scheduler：Pod 调度

调度器监听 Pod 创建事件（Create），之后通过具体的调度策略为 Pod 绑定将要调度到的物理节点 Node，通过 Put 更新 Pod Spec 中的 Node name 字段为所绑定的物理节点名称实现调度。之后对应物理节点上的 Kubelet 会监听到 Pod 修改事件（Modify），发现是新调度至自己节点的 Pod，就会实际创建并运行 Pod。

##### 调度策略

调度器共支持三种调度策略，分别为 `NodeAffinity`，`PodAntiAffinity` 与 `Round Robin`

-  `NodeAffinity`：Pod 可以直接指定希望在哪个 Node 上运行（通过在 yaml 配置文件中指定 Node name）
-  `PodAntiAffinity`：Pod 可以指定和拥有某种 label 的 Pod 不运行在相同的 Node 上；调度时会尽可能满足 Pod 的 AntiAffinity 需求，当然如果当前所有 Node 都不能满足（比如所有 Node 上都跑了所指定的不能与其一同运行的 Pod），则此配置不生效
-  `Round Robin`：新来的 Pod 依次轮流调度到各个 Node 上；期间通过 `NodeAffinity` 调度的 Pod 不会影响 RR 队列，通过 `PodAntiAffinity` 调度的 Pod 会将被调度到的节点置于 RR 队列的末尾

**Pod 反亲和性配置案例**

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
    tier: frontend
    scheduleAntiAffinity: large
  name: myapp-schedule-large
  namespace: default
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchLabels:
            scheduleAntiAffinity: tiny
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx
    ports:
    - containerPort: 80
      protocol: TCP
    resources: {}
  restartPolicy: Always
```

通过其中 `affinity` 字段 `podAntiAffinity` 下的 `requiredDuringSchedulingIgnoredDuringExecution` 配置反亲和性。具体而言，配置中 `labelSelector` 表明不希望调度到有对应 `matchLabels` 标签的 Pod 的 Node 节点。

##### 调度逻辑

1. 在调度时，会首先判断新创建的 Pod 有无指定 `NodeAffinity`（通过 Pod 的 Spec 中 Node name 字段），如果有则直接调度至对应 node，无则判断有无指定 `PodAntiAffinity`
2. 如果指定了 `PodAntiAffinity`，会尝试采用此策略进行调度；否则直接采用默认的 `Round Robin` 策略调度
   -  `PodAntiAffinity` 中会通过新创建 Pod 的 label selector 判断各个 Node 上现有的 Pod 的 label 是否与其相符，来决定新 Pod 不能调度到哪些 Node 上；如果所有 Node 都被排除，会无视反亲和性配置，采用  `Round Robin` 进行调度
   -  如果通过 `PodAntiAffinity` 调度成功，会将 RR 队列中对应的 Node 移到末尾
3. `Round Robin` 策略通过维护一个 Node 队列实现，每次调度时取队首 Node ，之后将对应 Node 放置队尾，实现 RR 目的
4. 调度时只会调度到状态正常（Running）的 Node 上

### Service

支持 Service 抽象，支持多个 Pod 的通信。用户能够通过所指定的虚拟 IP 访问 Service，由 Minik8s 将请求转发至对应的具体 Pod（可以看作一组 Pod 的前端代理）。Service 通过 selector 筛选包含对应 label 的 Pod，并将发往 Service 的流量通过 Round Robin 的策略负载均衡到这些 Pod 上。在符合 selector 筛选条件的 Pod 更新时（如 Pod 加入和 Pod 被删除），Service 会动态更新（如将被删除的 Pod 移出管理和将新启动的 Pod 纳入管理）。

Service 的抽象会隐藏 Pod的 具体运行位置，即 Pod 无论运行在哪个物理节点，都可以通过 Service 提供的 IP 访问到。

该配置可以通过类型为 Service 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: notice
  name: notices
  namespace: default
spec:
  ports:
    - name: hello
      port: 80
      targetPort: 80
  selector:
    app: notice
  clusterIP: 10.6.0.1
  type: ClusterIP
```

**主要功能**

允许用户定义虚拟 ip，将对 pod 的访问封装成对 ServiceIP 的访问。用户可以定义任意 IP 地址，将其作为 Service 的访问地址。同时用户可以通过 selector 配置 Service 对应的 Pod。默认的调度策略为 round robin。

**实现简述**

minik8s 中使用 ipvs 作为底层 NAT 实现，ipvs 作用在 INPUT 与 POSTROUTING 链上，因此需要使 serice 对应的 ip 可以被路由到本地，否则包会进入 FORWARD 链，INPUT 链上的规则不会生效。

具体地，对于每一个 service，其 虚拟的 ip 地址均会被绑定到虚拟网卡 `minik8s-proxy0` 上，并通过 ipvs 添加相应的规则。

### ReplicaSet：Pod 数量控制

支持 ReplicaSet 抽象。ReplicaSet 对 Pod 指定一定数目的期望数量（`replicas`），并且监控这些 Pod 的状态。当 Pod 异常（发生 crash 或者被 kill 掉）时，会自动根据 Pod Spec 模板启动新 Pod（或接管已有的 `label` 符合对应 `selector` 条件的 Running Pod），使得 ReplicaSet 管理的 Pod 数量（同时 `label` 符合对应 `selector` 条件）恢复到 `replicas` 指定的数目。ReplicaSet 的 Pod 实例支持跨多机部署。

该配置可以通过类型为 ReplicaSet 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  labels:
    app: myapp
    tier: frontend
  name: myapp-replicas
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      tier: frontend
  template:
    metadata:
      labels:
        app: myapp
        tier: frontend
    spec:
      containers:
        - image: nginx
          imagePullPolicy: Always
          name: nginx
          ports:
            - containerPort: 80
              protocol: TCP
          resources: {}
      restartPolicy: Always
```

**主要功能**

此功能主要由 ReplicaSet Controller 负责，其维护与 `selector` `matchLabels` 标签匹配的 `replicas` 数量的 Pod，多删少增。

当创建 ReplicaSet 时，如果已经有 Pod，并且其 `label` 匹配 ReplicaSet 的 `selector`，ReplicaSet 会直接接管这些 Pod；此后没有这样满足要求的 Pod 才会根据其中的模板 `template` 字段创建新的 Pod。

对于原本受到 ReplicaSet 管理的 Pod 的 `label` 发生更新时，会重新检查是否符合 ReplicaSet 的 `selector` 匹配，否的话会新接管/创建 Pod。

**字段意义**

其中各个字段意义如下：

- `replicas`：期望的副本数量。ReplicaSet 会维护自己管理的 Pod 数量与此一致
- `selector`：用户选择标签匹配的 Pod 的选择器。Pod label 标签键和值必须与此匹配才能被这个 ReplicaSet 所控制。此处的标签必须与 Pod 模板（`template`）的标签相匹配（也即依据 Pod 模板创建的 Pod 必须能够被该 selector 选择，以被 ReplicaSet 管理）
- `template`：Pod 模板，描述在检测到当前实际管理的 Pod 数量不足 `replicas` 时，将创建的 Pod 对象

### Auto scaling：动态伸缩

支持 HPA（`HorizontalPodAutoscaler`）抽象，可以根据其管理的 ReplicaSet 所管理的所有 Pod 中任务的实时负载，对 ReplicaSet `replicas` 数量进行动态扩容和缩容，使 ReplicaSet 所管理的所有 Pod 占用的资源量满足给出的限制。Pod 中任务的实时负载通过每个物理机节点上的 cadvisor 进行实时监控和数据收集（目前支持 cpu 和内存占用指标）。HPA 的 Pod 实例支持跨多机部署。

用户可以在配置文件自定义所要监控的资源指标及相应的扩缩容标准，包括 CPU 使用率和内存使用率。用户也可以在配置文件中自定义扩缩容策略，以限制扩缩容速度和方式。

该配置可以通过类型为 HorizontalPodAutoscaler 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  name: hpa-practice-cpu-policy-scale-up
spec:
  minReplicas: 3
  maxReplicas: 6
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 20
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 20
  scaleTargetRef:
    apiVersion: apps/v1
    kind: ReplicaSet
    name: myapp-replicas
  behavior:
    scaleUp:
      selectPolicy: Max
      stabilizationWindowSeconds: 0
      policies:
      - type: Pods
        value: 1
        periodSeconds: 15
```

**主要功能**

- 扩缩容：以扩容为例，HPA 的目标 ReplicaSet 管理的 Pod 负载增加时，如果达到扩容策略 metrics 规定的值，HPA 会增加 Pod 数量（通过修改所管理的 ReplicaSet Spec 的 `replicas` 字段实现），最大会增加到 `maxReplicas`。缩容同理。
- 与 ReplicaSet 一样，扩缩容所新创建的 Pod 将分布在不同节点中。
- 扩缩容策略：用户可自定义扩缩容策略，包括对扩缩容的速度限制，时间间隔限制等

**字段意义**

其中各个字段意义如下：

- `minReplicas`：HPA 自动扩缩时最少能缩小到多少 `replicas`
- `maxReplicas`：HPA 自动扩缩时最多能扩增到多少 `replicas`
- `metrics`：扩缩容决策所依据的资源指标；定义了在当前资源指标的量化标准下，应该怎么 scale
  - `type`：资源指标类型，目前仅支持 `Resource`
  - `resource`：Resource 资源指标信息
    - `name`：资源名称，目前支持 `cpu` 和 `memory`
    - `target`：资源的目标值
      - `type`：目前支持 `AverageValue` 和 `Utilization`
      - `averageValue`：当指标的平均值或资源的平均利用率超过这个的时候，会进行 scale（对于所有相关 Pod 计算该指标平均值）
        - 平均量（AverageValue）= 总量 / 当前实例数
      - `averageUtilization`：当整体的资源利用率超过这个百分比的时候，会进行 scale，百分比表示（对于所有相关 Pod 计算该指标平均值）
        - 利用率（Utilization）= 平均量 / Request
- `scaleTargetRef`：HPA 控制的对象，当前仅支持 ReplicaSet
- `behavior`：扩缩容策略，其中 `scaleUp` 与 `scaleDown` 分别配置扩容与缩容策略
  - `stabilizationWindowSeconds`：从上一次 auto scale 事件开始必须经过 `stabilizationWindowSeconds` 的秒数，才可以进行下一次的自动 scale
  - `selectPolicy`：对所配置策略组 `Policies` 中各个 policy 结果如何综合（每个 policy 规范 scale 所扩/缩的 Pod 数量应当不大于多少）
    - `Max`：从 Policies 的所有 Policy 中选出扩/缩的 Pod 数量最多的
    - `Min`：从 Policies 的所有 Policy 中选出扩/缩的 Pod 数量最少的
    - `Disabled`：禁止这一维度的 scale（也即不允许自动扩容 `ScaleUp` 或自动缩容 `ScaleDown` ）
  - `policies`：具体策略（数组，可配置多个）
    - `type`：
      - `Pods`：表示 scale 所对 Pod 数量做的变化 delta 需要小于等于 `Value` 的数值（限定变化的绝对数量）
      - `Percent`：此时 `Value` 对应 0 至 100，表示百分之几；表示 scale 所对 Pod 数量做的变化 delta 需要小于等于当前现有 Pod 数量的百分之多少（如 Value 为 100，则 scale 的增/减数量至多为当前 Pod 数量个 Pod，也即至多倍增/全删）
    - `PeriodSeconds`：从上一次 auto scale 事件开始必须经过 `PeriodSeconds` 的秒数，此 Policy 才可以生效

**扩缩容默认策略**

如果配置文件中未定义策略，则默认如下：

- 扩容：当资源指标满足需要扩容的条件时，在以下两个原则中取高值扩容，允许至多扩容至 `maxReplicas` 数量
  - 每 15 秒至多增加 1 个 Pod
  - 每 60 秒至多倍增 Pod 数量至当前数量的两倍
  - `stabilizationWindowSeconds` 为 0
- 缩容：当资源指标满足需要缩容的条件时，允许至多缩小至 `minReplicas` 数量
  - 每 15 秒至多减少 100 个 Pod
  - `stabilizationWindowSeconds` 为 300

**实现原理**

- 实际资源使用情况信息的收集与监控：每个物理节点上部署了 cadvisor，用于监控当前节点上实时的 CPU 和内存资源占用信息（包括物理机的总资源信息和每个容器的占用信息）。同时控制面中的 HPAController 会通过 cadvisor client 与其交互，每次需要相应信息时就发起请求，收集最近一段时间内的资源占用 status 情况（包括若干个时间点的采样指标）
- 使用信息整合：控制面中的 HPAController 下的 Mertic Client 会将这些各容器资源占用信息按照 Pod 进行整合，从而得知 Pod 实际资源使用
- 扩缩容决策：HPAController 通过 Pod 实际资源使用得到 ReplicaSet 的实际资源使用，并依据对应 HPA 中的资源占用要求进行决策，决定是否扩缩容
- 扩缩容执行：如果决定扩缩容，则按照相应的扩缩容策略执行，也即通过修改所管理的 ReplicaSet Spec 的 `replicas` 字段实现

### DNS 与转发

***注意：minik8s 的 dns 与 kubernetes 有较大区别***

DNS 支持用户通过 yaml 配置文件自定义域名，将 http 请求的路径与集群中其他的 http 服务绑定。
DNS 允许用户通过域名与路径的组合将集群中多个 http 服务聚合到同一域名下。

该配置可以通过类型为 DNS 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: v1
kind: DNS
name: dns-test
spec:
  serviceAddress: 10.8.0.1
  hostname: hello.world.minik8s
  mappings:
    - address: http://10.6.1.1:80
      path: "/world"
    - address: http://10.6.1.2:80
      path: "/new/world"
```

其中 `/world` 路径对应的 Service 的 yaml 配置文件示例如下：

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dns-test
  name: dns-world
  namespace: default
spec:
  ports:
    - name: notice
      port: 80
      targetPort: 80
  selector:
    app: world
  clusterIP: 10.6.1.1
  type: ClusterIP
```

**字段意义**

其中各个字段意义如下：

- `serviceAddress`: 域名所绑定的虚拟 IP
- `hostname`：域名的主路径
- `mappings`：子路径映射（path 列表）
  - `address`：对应的 Service 名称和端口
  - `path`：具体的路径地址

上述配置文件案例中，用户先创建了一个 Service（IP 为 `10.6.1.1`），可以使用 `ServiceIP:Port` 来访问该Service。
而配置DNS和转发后，用户和 Pod 内就可以通过 `hello.world.minik8s:80/path` 来访问该 Service，起到和 `ServiceIP:Port` 相同的效果。

**实现原理**

域名绑定虚拟 IP 部分的映射通过 coreDNS 实现。虚拟 IP 将绑定 Nginx Pod，通过 Nginx 实现子路径转发。

### 容错

Minik8s 的控制面有容错功能（包括 ApiServer，Controller，Scheduler，节点的 Kubelet，Kubeproxy 等）。控制面组件支持 crash 后重启，重启过程中及重启后原本的 Pod 和 Service 都能够正常运行，不会被控制面 crash 所影响。

同时，支持通过[心跳机制](#####Heartbeat 心跳机制)自动检测并删除长时间失联/挂掉的节点，节点恢复后可重新加入集群。

### GPU 应用支持

支持用户编写 CUDA 程序的 GPU 应用，并帮助用户将 CUDA 程序提交至[交我算平台](https://docs.hpc.sjtu.edu.cn/index.html)编译和运行。

用户只需要编写 CUDA 程序，并通过 yaml 配置文件提交对应 Job，Minik8s 会通过内置的 server 自动生成 slurm 脚本，并将程序上传至交我算平台编译运行。在任务执行完后 Minik8s 会自动下载结果到用户端，同时可配置通过交大邮箱通知用户。

该配置可以通过类型为 Job 的 yaml 配置文件来指定，示例如下：

```yaml
apiVersion: v1
kind: Job
metadata:
  name: matrix-sum
  namespace: default
spec:
  cuFilePath: D:\SJTU\Minik8s\minik8s\pkg\gpu\cuda\sum_matrix\sum_matrix.cu
  resultFileName: sum_matrix
  resultFilePath: D:\SJTU\Minik8s\minik8s\pkg\gpu\cuda\sum_matrix
  args:
    numTasksPerNode: 1
    cpusPerTask: 2
    mail:
      type: all
      userName: albus_tan
```

**主要功能**

- 用户可以通过编写 Job 类型的 yaml 配置文件提交运行编写的 CUDA 程序
- 根据 yaml 配置文件自动生成 slurm 脚本，并将用户编写的 CUDA 程序上传至交我算平台编译运行
- 可配置每节点核数，任务能使用的 CPU，GPU 数量等
- 提交成功后，可以通过 get job 方法得到当前 Job 的实时执行状态（Pending，Running，Failed，Completed）
- 支持任务开始时/完成后通过交大邮箱通知用户
- 任务完成后，自动将结果下载至用户端指定的目录中

**字段意义**

- `cuFilePath`：用户想要提交的 CUDA 程序的路径
- `resultFileName`：执行结果的文件名
- `resultFilePath`：执行结果下载至本地的路径
- `args`：任务可配置参数
  - `numTasksPerNode`：每节点核数
  - `cpusPerTask`：使用 CPU 数量
  - `gpuResources`：使用 GPU 数量
  - `mail`：任务状态改变时通过交大邮箱通知用户
    - `type`：支持 begin（任务开始时通知），end（任务结束时通知），fail（任务失败时通知），all（任务状态变化时通知）
    - `userName`：用户的交大邮箱用户名，如此处填写 `albus_tan`，则会将通知邮件发送至 `albus_tan@sjtu.edu.cn`

**实现原理**

- 任务提交：server 监听 Job 创建事件，之后通过 ssh client，连接登录交我算 π 2.0 集群，使用 CUDA 编译 .cu 文件，并提交 dgx2 队列作业（GPU 任务队列）
- 任务状态获取：server 后台线程每隔一段时间通过 squeue 命令和 sacct 命令查看作业的状态，并修改对应 Job 的 status 字段
- 自动结果下载：server 后台线程监听 Job 修改，当发现 Job status 对应字段显示 Job 完成，通过 sftp 从交我算平台下载对应 Job 结果文件夹中的执行结果到本地目标路径

**矩阵乘法和加法程序**

![](README.assets/cuda.png)

`blockIdx` 代表一个 block 的坐标，例如，左上角的块的坐标为 `blockIdx(0,0)`；`blockDim` 代表一个 `block` 的尺寸，一个 `block` 是二维的，`blockDim.x` 代表宽度，`blockDim.y` 代表高度。`threadIdx` 代表一个 block 内线程的坐标，与 `blockIdx` 类似。将 CUDA 网格（grid）中的每个块都对应于矩阵中的一个区域，也即可以将块（block）中的一个单元映射到矩阵中的一个元素：

```c++
int i = blockIdx.x * blockDim.x + threadIdx.x;
int j = blockIdx.y * blockDim.y + threadIdx.y;
```

在 GPU 上执行的函数称为 CUDA 核函数，核函数会被 GPU 上多个线程并行执行，用 `__global__` 声明，在调用时需要用 `<<\>>` 来指定 kernel 要执行的线程数量和维度结构。

矩阵加法 CUDA 核函数：

```c++
__global__ void matrix_add(int **A, int **B, int **C) {
    int i = blockIdx.x * blockDim.x + threadIdx.x;
    int j = blockIdx.y * blockDim.y + threadIdx.y;
    C[i][j] = A[i][j] + B[i][j];
}
```

矩阵乘法 CUDA 核函数：

```c++
__global__ void matrix_multiply(int **A, int **B, int **C) {
    int i = blockIdx.x * blockDim.x + threadIdx.x;
    int j = blockIdx.y * blockDim.y + threadIdx.y;
    int value = 0;
    for (int k = 0; k < N; k++) {
        value += A[i][k] * B[k][j];
    }
    C[i][j] = value;
}
```

以矩阵乘法为例，调用方法如下（结果为 M*M 的矩阵），每一个线程运算结果矩阵的一个元素：

```c++
dim3 threadPerBlock(5, 5);
dim3 numBlocks(M / threadPerBlock.x, M / threadPerBlock.y);
matrix_multiply <<<numBlocks, threadPerBlock>>> (dev_A, dev_B, dev_C);
```

### Serverless

Serverless 平台能够提供按函数粒度运行程序，并且支持自动扩容和 scale-to-0，能够支持函数链的构建，并且支持函数链间通信。目前支持 Python 语言的函数。

分为 v1 和 v2 两个版本，用户可根据函数应用场景选择适合的版本：

- `Serverless v1`：适用于 Unlikely Path（被调用频率很低的函数，如错误处理函数等），每次函数被调用时会创建 Pod 运行相应函数实例，函数调用完成后 Pod 会立即析构，释放资源
- `Serverless v2`：适用于普通函数，在函数被首次调用时会创建实例（冷启动），之后函数再次被调用时实例会被复用（热启动），不会再重新创建新实例（调用返回速率会是冷启动的 10 倍左右）。如果某个函数调用频率很高，会为其创建多个实例，系统会自动将用户的调用请求分发至各个实例上。当某个函数一段时间不被调用，其实例数量会被逐渐减少，直至 `scale-to-0`。

两个版本的函数都支持各种 Workflow，包括条件判断，循环等。两个版本的函数可以互相调用，相互兼容，用户只需在配置文件中定义希望使用的版本即可。

**主要功能**

- 用户可以定义函数内容（函数模板），并上传至系统，此后可以对模板进行修改和删除
- 上传函数内容（函数模板）后，用户可以通过 http 请求对函数进行调用，传入参数，并得到返回结果
  - 如果长时间函数尚未执行完，会先返回本次调用的 `id` 号，用户可以在一段时间后通过 `id` 号进行调用结果查询；如果函数在默认超时时间内执行完成，会直接将结果反馈给用户
- 用户上传的每一个函数会在独立的 Pod 内运行，以确保隔离性
- Workflow：用户可以定义多个函数之间的 Workflow 调用关系，支持条件判断，循环等（具体而言，用户只需要指定真值判断条件，并指定条件判断为真应当执行哪个函数，为假应当执行哪个函数即可）
- 自动扩容：对于某一个函数（函数模板），当首次被调用时（函数实例不存在）会自动生成新的实例；当对于这个函数的请求并发数增多时，此函数会被自动扩容成多个实例，并且用户的调用请求可以发送至这些实例中的任意一个进行处理
  - 对于首次调用，如果希望在模板定义后立即调用，期望定义模板时系统就开始准备冷启动，使得初次调用获得更快响应，可以指定初始化实例数量；该参数默认为 0
  - 可以在配置中指定函数的最大最小实例数量
- scale-to-0：当一段时间没有新的请求到来时，对应函数实例会逐步减少，直至完全清零（或至指定的最小实例数）

该配置可以通过配置文件来指定，示例如下：

```bash
# is_hello.env
API_SERVER=192.168.1.10
PORT=8080
VERSION=v2
NAME=is_hello
MAIN=examples/dash/func/is_hello.py
PRE_RUN=examples/dash/func/nop.sh
LEFT_BRANCH=append_world
RIGHT_BRANCH=append_world
ADDR=10.7.0.3
```

```python
# is_hello.py
def run(arg):
    return arg

def check(arg):
    return arg == "hello"
```

**字段意义**

`.py` 文件中定义了函数的具体内容，用户需要在 `.py` 文件中定义两个函数：

- `run(arg)`：主函数体，相当于 `main`，其中应当定义函数的主要执行逻辑和内容；可以传入参数 `arg`（string 类型，如果有多个参数/其它类型的参数，需要用户自行实现编码解码）；可以有返回值（string 类型，如果有多个返回值/其它类型的返回值，需要用户自行实现编码解码）。
- `check(arg)`：在 `run` 执行完之后被自动调用，此函数返回值为 True/False，依据 `check` 返回值决定当前函数调用结束之后，会继续调用哪个函数。如果 `check` 返回值为 True，则会执行 `.env` 文件中 `LEFT_BRANCH` 函数名对应的函数，否则执行 `RIGHT_BRANCH` 函数名对应的函数。此函数参数 `arg` 为 `run` 函数返回值，用户书写 Workflow 逻辑时可以依据 `run` 函数的执行结果来决定接下来的执行流走向。

`.env` 文件中定义了函数 Workflow 及相关配置信息等：

- `API_SERVER`：对应集群的 ApiServer 的 IP 地址
- `PORT`：对应集群的 ApiServer 的端口
- `VERSION`：函数希望使用的 Serverless 版本
- `NAME`：函数名（需确保全局唯一性，函数的唯一标识）
- `MAIN`：包含函数内容的文件路径（目前仅支持 `.py` 文件）
- `PRE_RUN`：在当前函数执行前会执行的 shell 脚本
- `LEFT_BRANCH`：在当前函数执行完成后，如果 `check` 返回值为 True，则会执行 `LEFT_BRANCH` 函数名对应的函数
- `RIGHT_BRANCH`：在当前函数执行完成后，如果 `check` 返回值为 False，则会执行 `RIGHT_BRANCH` 函数名对应的函数
- `ADDR`：指定函数对应的 Service IP 地址，也即可以通过访问该地址，通过 Service 的转发机制，转发 call 请求到具体的函数实例 Pod

**使用**

用户能够通过 http 请求定义/修改/删除函数模板，或进行函数调用。为简化用户操作，可以通过调用 `./script` 目录下的脚本进行函数模板的上传，修改，函数调用，异步获取结果等操作（对于函数模板的定义/修改/删除也可以通过 kubectl 完成）：

- [`uploader.sh`](./script/uploader.sh)：上传新的函数模板
- [`updater.sh`](./script/updater.sh)：修改已定义的函数模板
- [`call.sh`](./script/call.sh)：调用函数（需要已经上传过对应的函数模板）
- [`get-result.sh`](./script/get-result.sh)：异步获取函数调用的结果

**使用案例**

```bash
# 上传所有函数模板（包括入口以及 Workflow 中可能被调用到的所有函数）
$ script/uploader.sh examples/dash/func/is_hello.env
$ script/uploader.sh examples/dash/func/append_world.env
$ script/uploader.sh examples/dash/func/append_branch.env
# 查看已上传的函数模板
$ build/kubectl get func
# 调用函数（参数为函数名以及传入的函数参数）
$ bash script/call.sh is_hello hi
# 更新函数模板
$ bash script/updater.sh examples/dash/func/is_hello_modify.env
# 删除函数模板（参数为函数名）
$ build/kubectl del func append_branch
```

**实现简述**

详细实现介绍部分参见 [Serverless](./doc/Serverless.md)，两个版本的实现不同主要集中在内部接口部分

- 函数模板 Function Template：函数模板，对应 Func 类型 ApiObject，可通过以下 URL 对模板进行增删改查

  ```bash
  /api/funcs/template
  /api/funcs/template/:name	# name 为函数模板中 Spec 中的 Name
  ```

- 函数实例管理和调用 Function Instance and Call：真正进行函数调用的接口，分为对用户暴露的接口和内部实现使用的接口

  - 用户接口
    - `POST /api/funcs/:name` （body 部分进行参数传递）
      - 依据名为 name 的函数模板创建并运行实例，返回实例 id 即 `instanceId`
        - 会生成本次调用的实例 id 即 `instanceId`，并等待函数调用返回写入 `etcd`；如果长时间还未检查到结果，会先返回 `instanceId`，用户可以在一段时间之后调用 `GET /api/funcs/:id` 查看函数执行结果；如果在超时时间内检查到结果，直接返回函数执行结果给用户
        - 调用内部接口 `PUT /api/funcs/:name/:id`
    - `GET /api/funcs/:id`
      - 依据实例 id 即 `instanceId` 查看所调用函数返回的结果
  - 内部接口
    - `PUT /api/funcs/:name/:id`（body 部分进行参数传递）
      - 用户调用的实例 id 即 `instanceId` 参数用于识别是用户哪次实际调用中的调用流，也便于存储最终结果
      - 依据 name 字段调用对应函数，如果 name 字段为 RETURN，则说明此次调用负责存储函数返回值

#### Serverless v1 实现

在内部接口 `PUT /api/funcs/:name/:id`（body 部分进行参数传递）被调用时，会直接创建相应 Pod，Pod 中会在名为 name 函数逻辑结束后，自动调用下一个将要执行的函数（同样是 `PUT /api/funcs/:name/:id` 接口，name 字段设置为下一个将被调用的函数名即可），并将用户调用的实例 id 即 `instanceId` 递归传递；同时会自我调用当前 Pod 的 delete 方法，自行析构。

#### Serverless v2 实现

在用户创建/修改函数模板时，ServerlessController 会监听对应事件，并为该函数模板创建对应 ReplicaSet（用于管理当前函数模板的所有 Pod 实例），并创建其对应的 Service（用于提供访问该函数模板所有 Pod 的统一入口）；同时 ServerlessController 也负责根据函数近期调用频次周期性改变 ReplicaSet 的 `replicas` 字段数目，以实现 scale 的功能（通过最近调用的时间戳和，以及被调用次数等，根据算法进行计算实现）。

在内部接口 `PUT /api/funcs/:name/:id`（body 部分进行参数传递）被调用时，会将调用的 HTTP 请求直接转发给对应 Service，Service 会将请求转发给对应 label match 的 Pod（如果对应 Pod 无响应就转发至下一个，直至超时），label 部分由用户告知 apiserver 创建 func template 时，ServerlessController 自动根据 func 的名字生成（同时也会生成对应 func server 的 Pod 模板），Pod 中会在名为 name 函数逻辑结束后，调用下一个函数（同样是 `PUT /api/funcs/:name/:id` 接口，name 字段设置为下一个将被调用的函数名即可），并将用户调用的实例 id 即 `instanceId` 递归传递。

Pod 个数由所属 func template 中的 ReplicaSet 管理。每当出现新的对函数的调用请求，更新对应 func template 中 status 里的 timestamp 时间戳，同时增加 counter。每隔一定时间，将所有现存的 func template 中的 counter 统一减少一定数值，同时 replicaset 中的 `replicas` 数量保持与 counter 一致，从而实现函数不被调用时 scale-to-zero。设置策略限定 counter 上界，同时优化 counter 不同时的扩缩策略，实现更佳效果。

## 实现简述

### 相关实现文档

实现部分的文档可以参考以下链接：

- [API & API 对象](./doc/API.md)
- [ApiServer, ApiClient 及 ListWatch](./doc/ApiServer.md)
- [Scheduler](./doc/Scheduler.md)
- [Controller（Informer, ReplicaSetController 及 HorizontalController）](./doc/Controller.md)
- [Gpu Server](./doc/GPU.md)
- [Kubelet](./doc/Kubelet.md)
- Kubeproxy
- [Serverless](./doc/Serverless.md)
- [Node Manager](./doc/Node.md)
- [CI/CD](./doc/CI CD.md)
- [CNI](./doc/CNI.md)
- [Test](./doc/Test.md)

### 实现亮点

#### ApiServer

##### Concurrency Control and Consistency 乐观并发控制

API 对象资源更新时支持 ResourceVersion 检查，避免了多个组件/用户同时更新同一个对象资源时可能的并发问题，如两个组件同时基于同一版本的对象更新了某对象的 A 与 B 字段，之后 Put 时后更新的就会将先更新的字段覆盖，从而导致部分字段的更新遗漏。

> Ref：https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

参照 Kubernetes 的实现，所有资源都有一个 `resourceVersion` 字段，作为其元数据的一部分。这个 ResourceVersion 是一个字符串，它标识了一个对象的内部版本，可以被客户端用来确定对象何时改变。当一条记录要被更新时，它的版本会与一个预先保存的值进行核对，如果不匹配，更新就会以 StatusConflict（HTTP状态码409）失败。

资源版本目前是由 [etcd 的 mod_revision](https://etcd.io/docs/latest/learning/api/#key-value-pair) 支持的。然而，需要注意的是，应用程序不应该依赖所维护的版本系统的实现细节。我们可能会在未来改变资源版本的实现，比如把它改成一个时间戳或每个对象的计数器。

由于 mod revision 在对 etcd 的 Put 操作后才能获取到，同时需要将这个 revision 写入 API 对象自身的 ResourceVersion 字段，实现时需要注意同步维护全局 revision 号（通过 `ResourceVersionManager` 维护），并且保证获取下一次 mod revision，写入 API 对象自身的 ResourceVersion 字段，存储 API 对象这一系列 Get version，Set version，Store Object 操作同一时刻只能有一个在发生。实现时添加锁 `VLock` 来保障这一点。

#### ListWatch

通过 `client.Interface` 创建，封装接口为 `ListWatcher`，专门用来调用对应资源的 `GetAll` 与 `WatchAll` 方法。

##### Watch 监听机制

可以对某个/种 API 对象进行监听，当其发生创建，修改或者删除时，获得通知。

- 在 watch 请求后建立 http 长连接，通过 `http.Flusher` 将 event 实时刷新给请求者，而不用断开重连。
- 内部通过 `etcd` 的 `Watch` 机制实现。对某个 `key` 进行监听，每当对应 `value` 发生修改，就会通过 channel 进行通知。 `etcd` 的 `Watch` 在键值对被删除时响应 value 为 `“”`，考虑到许多组件需要获取被删除前的内容，因此使用 `clientv3.WithPrevKV()` 添加这个字段。

##### StreamWatcher

`ListWatcher` 接口的实现，需要对资源进行 List 和 Watch 操作的组件可以通过 `ListWatcher` 轻松的实现监听：

- `Decoder`：负责将 `Watch` 监听到的 `ApiServer` 发来的事件类型转换为 `watch.Event` 类型
  - 此处 `ApiServer` 发来的事件类型为 `Etcd` 内置事件类型，这么做的好处在于解耦，修改实现只需要实现对应 `Decoder interface` 接口即可
- `Reporter`：错误处理，将报告错误的事件转换为标准的 `watch.Event` 类型，同时也将过程中产生的错误 转换为标准的 `watch.Event` 类型
- `chan Event`：通过此通道将监听到的事件发送出去，使用 `StreamWatcher` 的组件可以通过 `watch.Interface` 中的 `ResultChan()` 拿到这个通道，并获取事件；通道使用完成后/结束时需要通过 `Stop()` 方法关闭通道

#### Controller

参照 Kubernetes，实现了 Informer 和 WorkQueue 两个基本组件，方便所有 Controller 的实现。

##### Informer 本地缓存

相当于每个 Node 上对于不同 ApiObject 的本地缓存，避免频繁向 ApiServer 发送网络请求，可以很大程度上提升性能。每一种 ApiObject 资源对应一个 Informer（由 `objType` 指定）。启动时其中的 `Reflector` 通过 `List` 向 ApiServer 拿取所有 ApiObject 信息，并且存储在 `ThreadSafeStore` 中。之后其 `Reflector` 通过 `Watch` 监听所有对应 ApiObject 的变化事件，并存储在 `ThreadSafeStore` 中，同时调用注册进来的 `ResourceEventHandler` 进行相应处理。

- `Reflector`：启动时先通过 `List` 向 ApiServer 拿取所有 ApiObject 信息，并且存储在 `ThreadSafeStore` 中，之后通过 `Watch` 监听所有对应 ApiObject 的变化事件，并通知  `Informer` （通过将事件放入 `WorkQueue`）；其中 `List` 与 `Watch` 都由 `listwatch.ListerWatcher` 组件完成
- `ThreadSafeStore`：与其 `Reflector` 共享同一个存储，存储对应 ApiObject 对象的本地缓存
- `ResourceEventHandler`：注册对于各种 `Watch` 事件的响应，使用 `Informer` 的组件可以通过 `AddEventHandler ` 添加对应处理函数
- `WorkQueue`：每次 `Reflector` 监听到新事件，就放进此队列，等待 `Informer` 在 `run` 中进行处理，并调用相应注册进来的 `EventHandler` 函数

##### WorkQueue 工作队列

工作队列可以允许 controller 中多个 worker 同时消费对象相关事件，实现处理并行化，提升性能。

- 线程安全的队列，通过读写锁允许多个线程同时处理而不出现并发问题
- 在 `Dequeue` 时如果队列为空，会通过 Conditional Variable 等待 `Enqueue` 操作唤醒，再尝试进行 `Dequeue`

## 组员分工和贡献度

- 谈子铭（44%）
  - ApiServer + Etcd
  - 制定 API 对象字段
  - ApiClient 及 ListWatcher
  - Controller 基本组件
  - ReplicaSet抽象和其基本功能
  - 动态伸缩 HPA 功能
  - 多机上实现容器编排的功能（Node 抽象与 Scheduler）
  - 完成 GPU 部分
  - DNS 及 Serverless Controller
  - CNI，CI/CD等的尝试，kubectl 重构等其它
- 王家睿（44%）
  - 实现 Pod 抽象，对容器生命周期管理
  - 制定 API 对象字段
  - 实现 kubelet 节点管理功能
  - 实现 CNI，支持 Pod 间通信
  - 实现 Service 抽象
  - 实现 DNS 抽象，实现转发功能
  - 实现 Serverless V1 和 Serverless V2 功能
  - gitlab CI/CD
  - HPA 场景构建及测试，kubectl 重构等其它
- 陆胤松（12%）
  - 实现部分 kubectl 命令行工具



## 安装教程

- Go 开发环境及 GoLand 项目配置 https://blog.csdn.net/m0_56510407/article/details/123544438

### etcd

控制面 Master 节点上需要安装 etcd

> 最新版本见 **[etcd](https://github.com/etcd-io/etcd)**，下载安装方式见 [install](https://etcd.io/docs/v3.5/install/)

### Cadvisor

每个 Worker 节点上需要安装部署 cadvisor，并在启动前启动，方可正常使用 HPA 功能

**使用二进制部署**

```sh
# 下载二进制
https://github.com/google/cadvisor/releases/latest
# 本地运行
./cadvisor  -port=8090 &>>/var/log/cadvisor.log
# 查看进程信息
ps -aux | grep cadvisor
# 查看端口占用
netstat -anp | grep 8090
```

**使用docker部署**

```bash
docker run \
--volume=/:/rootfs:ro \
--volume=/var/run:/var/run:rw \
--volume=/sys:/sys:ro \
--volume=/var/lib/docker/:/var/lib/docker:ro \
--volume=/dev/disk/:/dev/disk:ro \
--publish=8090:8090 \
--detach=true \
--name=cadvisor \
google/cadvisor:latest
```

**端口转发**

这样在本机上就可以看到远端机器上的 cadvisor

```
ssh -N minik8s-dev -L 8090:localhost:8090
```

### Flannel

每个节点都需要通过 flannel 进行网络配置。

Flannel配置第3层IPv4  overlay网络。它会创建一个大型内部网络，跨越集群中每个节点。在此overlay网络中，每个节点都有一个子网，用于在内部分配IP地址。在配置pod时，每个节点上的Docker桥接口都会为每个新容器分配一个地址。同一主机中的Pod可以使用Docker桥接进行通信，而不同主机上的pod会使用flanneld将其流量封装在UDP数据包中，以便路由到适当的目标。

**参考 [Running flannel](https://github.com/flannel-io/flannel/blob/master/Documentation/running.md) Running manually 章节**

若 wget 失败可以考虑手动上传文件到服务器

```bash
sudo apt install etcd
wget https://github.com/flannel-io/flannel/releases/latest/download/flanneld-amd64 && chmod +x flanneld-amd64
sudo ./flanneld-amd64
```

```bash
docker run --rm --net=host quay.io/coreos/etcd
```

```bash
docker run --rm -e ETCDCTL_API=3 --net=host quay.io/coreos/etcd etcdctl put /coreos.com/network/config '{ "Network": "10.5.0.0/16", "Backend": {"Type": "vxlan"}}'
```

**查看端口占用**

```bash
netstat -nap | grep 2380
```

**测试**

```bash
docker run -it  busybox sh
# 查看容器IP
$ cat /etc/hosts
# ping
ping -c3  10.0.5.2
```

