[![Build Status](https://travis-ci.org/ICKelin/gtun.svg?branch=master)](https://travis-ci.org/ICKelin/gtun) ![goreport](https://goreportcard.com/badge/github.com/ICKelin/gtun)

[doc](./README-EN.md) | [release](https://github.com/ICKelin/gtun/releases)

## gtun是什么

- gtun默认是个vpn隧道，能够将多个客户端连接组成虚拟局域网
- gtun是个ip代理，可以通过手动配置路由，对多个IP、IP段进行代理转发
- ~~gtun支持内网穿透功能~~ 最新版本的gtun不再支持内网穿透功能，如果需要使用内网穿透，可以移步至[opennotr](https://github.com/ICKelin/opennotr)

## 相关视频、教程

- [使用gtun跨可用区访问k8s集群]()


## 限制
gtun自身具备一定的穿墙功能，在法律允许的范围内使用gtun.

~~gtun自身使用了tun/tap虚拟网卡技术，tun/tap在各个系统对实现有些许差异，因此在使用gtun时需要考虑各个系统的限制~~

**最新版本的gtun只支持linux**


## gtun能解决什么问题

- IP地址加速，gtun本身包含代理功能，但是不包含加速功能，可以结合kcptun等工具实现代理加速
- 访问内网，gtun的客户端和服务端在一个虚拟局域网内，在服务端可以通过访问客户端的虚拟IP来访问客户端所在的机器（请注意其中的安全风险）
- VPC/k8s集群访问，gtun本身代理IP地址块的功能可用于打通单机器到VPC/k8s集群的访问。

## gtun如何使用

- 您需要有一台公有云机器，用于部署服务端程序（gtund），配置需要1C1GB即可，操作系统需要linux，并且确保安装iproute2软件
- 您需要一台linux的设备，可以是您正在用的电脑，树莓派等终端，配置需要1C1GB即可，操作系统需要linux，并且确保安装iproute2软件

**第一步：部署gtund**
在公有云上部署。

1. 下载gtund及配置文件模版

```
wget https://github.com/ICKelin/gtun/releases/download/v2.0.2/gtund_linux_amd64
wget https://github.com/ICKelin/gtun/releases/download/v2.0.2/gtund.toml

```

2. 根据需要修改配置文件，通常而言无需修改即可使用。

```gtund.toml
# 实例名称，暂时无效
name="us-node-1-1"

[server]
# 监听地址，根据需要修改，并且打开公有云安全组对应的端口
listen=":9623"
# 鉴权key，根据需要修改，处理客户端握手时根据此key判断是否为无效连接
auth_key="gtun-cs-token"

[dhcp]
# ip地址块
cidr="100.64.240.1/24"
# 本机在虚拟局域网当中的ip
gateway="100.64.240.1"

[log]
level="debug"
path="log.log"
days=3

```
3. 运行程序

```
chmod +x gtund_linux_amd64
sudo nohup ./gtund_linux_amd64 -c gtund.toml &
```

运行成功之后，执行ifconfig命令查看是否有tunX网卡存在并且配置了IP地址。

**第二步，部署gtun**
在任意linux部署均可。

1. 下载gtun及配置文件

```
wget https://github.com/ICKelin/gtun/releases/download/v2.0.2/gtun_linux_amd64
wget https://github.com/ICKelin/gtun/releases/download/v2.0.2/gtun.toml
```

2. 修改gtun.toml文件，将IP修改为gtund所在公有云的公网IP。

```
[client]
# gtund地址，修改为gtund部署所在机器的公网IP及端口
server = "192.168.31.65:9399"
# 鉴权key，与gtund.toml中的auth_key保持一致。
auth="gtun-cs-token"

[log]
level="debug"
path="log.log"
days=3
```

3. 运行gtun

```
chmod +x gtun_linux_amd64
sudo nohup ./gtun_linux_amd64 -c gtun.toml &
```
运行成功之后，执行ifconfig命令查看是否有tunX网卡存在并且配置了IP地址。

完成以上步骤之后，基本完成了整个程序的部署，以下是可选内容，如果您需要IP加速，可以参考。

**第三步(可选)：添加IP代理**

需要在gtun所在机器上进行操作

假设第二步当中gtun的网卡是tun0，现在需要配置8.8.8.8/16这个IP地址的代理功能，那么需要通过路由来操作

`ip ro add 8.8.8.8/16 dev tun0`

依次类推，如果需要加入更多的ip代理，将上面命令中的8.8.8.8/16修改即可。

## 更多
如果对网络感兴趣，可以查看我的一些[文章列表](https://github.com/ICKelin/article)，或者关注我的个人公众号.

![ICKelin](qrcode.jpg)
