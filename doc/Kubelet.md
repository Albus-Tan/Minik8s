# Kubelet

## Cadvisor

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


