# CI/CD

## docker部署jenkins

```bash
docker pull jenkins/jenkins:lts-jdk11
docker run --name c-jenkins -p 8008:8080 -p 50000:50000 --restart=always -u root -v /var/run/docker.sock:/var/run/docker.sock -v jenkins-data:/var/jenkins_home -d jenkinsci/blueocean
```

浏览器访问 http://localhost:8008/

## 初始化jenkins

浏览器访问jenkins后，需要输入初始密码，获取密码如下：

```bash
# 进入容器
docker exec -it c-jenkins bash
# 查看初始密码
cat /var/jenkins_home/secrets/initialAdminPassword
# 选择安装推荐的插件，有些插件安装失败可以先跳过，要用的时候再装
# 新建用户，也可以不新建，使用admin账户
```

## jenkins+Gitee 自动化构建

具体流程见 Ref

Reference: https://zhuanlan.zhihu.com/p/90612874

- 注意 WebHooks 中需要使用公网 IP，百度中搜索 IP 即可知道本机公网 IP

## 部署 gitlab-runner

```bash
# Replace ${arch} with any of the supported architectures, e.g. amd64, arm, arm64
# A full list of architectures can be found here https://gitlab-runner-downloads.s3.amazonaws.com/latest/index.html
curl -LJO "https://gitlab-runner-downloads.s3.amazonaws.com/latest/rpm/gitlab-runner_${arch}.rpm"

rpm -i gitlab-runner_${arch}.rpm
sudo gitlab-runner start
```



# Ref

1. [安装Jenkins](https://www.jenkins.io/zh/doc/book/installing/)
2. [学习使用gitee+jenkins实现cicd](https://blog.csdn.net/hyx1229/article/details/127213111)
3. [Jenkins 插件](https://gitee.com/help/articles/4193)
4. [Jenkins部署Golang](https://blog.csdn.net/weixin_46837396/article/details/119247154)
5. [安装Gitlab Runner](https://docs.gitlab.com/runner/install/index.html)