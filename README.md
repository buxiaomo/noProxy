# noProxy

多平台容器镜像代理服务,支持 Docker Hub, GitHub, Google, k8s, Quay, Microsoft 等镜像仓库.

## 快速搭建自己的加速器

### http

```shell
docker run -d --name noproxy \
-p 80:8080 \
--restart always \
ghcr.io/buxiaomo/noproxy:latest
```

### https

> 自动申请证书

```
docker run -d --name noproxy \
-e DOMAIN_NAME=www.example.com \
-p 443:443 \
--restart always \
ghcr.io/buxiaomo/noproxy:latest
```
## 公益仓库

> 测试域名: `duqqpwojzauy.cloud.sealos.io`
> 
> <b>下载速度取决于服务器带宽

### github文件代理下载

> 格式: `https://duqqpwojzauy.cloud.sealos.io/d/<target download url>`

```shell
wget https://duqqpwojzauy.cloud.sealos.io/d/https://github.com/mikefarah/yq/releases/download/v4.45.1/yq_linux_amd64
wget https://duqqpwojzauy.cloud.sealos.io/d/https://dl.k8s.io/v1.32.2/bin/linux/amd64/kube-apiserver
```

### docker镜像代理下载

> 格式: `duqqpwojzauy.cloud.sealos.io/<target pull url>`

```shell
docker pull duqqpwojzauy.cloud.sealos.io/docker.io/library/nginx:1.27
docker pull duqqpwojzauy.cloud.sealos.io/docker.elastic.co/elasticsearch/elasticsearch:7.17.9
```

## 域名白名单

- download.docker.com
- github.com
- get.docker.com
- docker.io
- k8s.gcr.io
- docker.elastic.co
- gcr.io
- ghcr.io
- nvcr.io
- quay.io
- dl.k8s.io
- cdn.dl.k8s.io

## 支持这个项目
### 用爱发电

我们提供的服务是免费的，但是为了维护这个项目，我们也需要花费一些精力和服务器带宽和存储费用。如果您觉得这个项目对你有帮助，欢迎您通过以下方式支持我们：

* Star 并分享 [noProxy](https://github.com/buxiaomo/noProxy.git) 🚀
* 通过以下二维码 一次性捐款，打赏作者一杯茶。🍵 非常感谢！ ❤️

| 微信 | 支付宝 |
|:--------:|:-------:|
| <img src="images/wxpay.png" width="200" /> | <img src="images/alipay.png" width="200" /> |

#### 提示

如有赞助行为，请务必添加备注，以便统一感谢！