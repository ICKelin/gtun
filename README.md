[![Build Status](https://travis-ci.org/ICKelin/gtun.svg?branch=master)](https://travis-ci.org/ICKelin/gtun) ![goreport](https://goreportcard.com/badge/github.com/ICKelin/gtun)

[doc](./README-EN.md) | [release](https://github.com/ICKelin/gtun/releases)

## gtun是什么

- gtun是个ip代理，能够最多个IP、IP段进行代理转发
- gtun默认是个vpn隧道，能够将多个客户端连接组成虚拟局域网
- gtun支持内网穿透功能(后期将内网穿透部分独立出来成[Notr](https://www.notr.tech)项目)

## 限制
gtun自身具备一定的穿墙功能，在法律允许的范围内使用gtun.

gtun自身使用了tun/tap虚拟网卡技术，tun/tap在各个系统对实现有些许差异，因此在使用gtun时需要考虑各个系统的限制：

| 操作系统 | tun | tap | ip加速 | 虚拟局域网 | 反向代理 |
|:-------:|:----:|:---:|:----:|:--------:|:-------:|
| Linux   |  是  |  否  | 是 | 是 | 是 |
| Mac OS  |  是  |  否  | 是 | 是 | 是 |
| Windows |  否  |  是  | 是 | 否 | 否 |

## gtun模块介绍
在最新版本当中，gtun包含三个模块

### gtund
gtund代理服务器，也就是出口服务器，所有需要转发的流量都是从gtund出去，到真实的服务器。主要包括以下几个功能：

- 在客户端连接之初，发送vpn的ip地址段等信息
- 在与客户端建立连接完成之时，用于接收客户端发送的数据包，并与远程服务器建立连接，进行一个转发功能

### gtun
gtun是客户端，也就是部署在需要ip加速的设备，电脑或者网关处。gtun才用虚拟网卡的方式，所以需要 **以root权限启动，针对windows，需要安装tap驱动**


### registry
服务注册中心节点，gtund启动时，会像registry模块注册，并使用长链接保活。gtun启动时，会向registry寻求gtund节点信息。registry程序主要用于自动选gtund节点的，在整个系统当中不是必须的，gtun可以通过手动配置gtund信息的方式，绕过registry。


## 如何使用

**下载源码与依赖**
``` shell
go get github.com/songgao/water
go get github.com/ICKelin/gtun
```
**编译**

```
./makefile.sh
```
在bin目录下会生成gtun的服务端二进制文件gtund，gtund大部分情况下都是部署在云服务器上，所以之编译了Linux版本，除了gtund之外，还会生成各个系统的gtun文件以及registry程序

**gtund部署**
需要将gtund部署到云服务器当中，同时需要开启系统的ipv4转发功能

vi /etc/sysctl.conf
```
net.ipv4.ip_forward=1
```

给iptables添加规则做SNAT

```
iptables -t nat -I POSTROUTING -j MASQUERADE
```

**gtun部署**
在本地电脑或者网关运行gtun，配置参考: [gtun.conf](https://github.com/ICKelin/gtun/blob/master/etc/gtun.conf)

**registry部署**
在公网服务注册，配置参考:[registry.conf](https://github.com/ICKelin/gtun/blob/master/etc/registry.conf)

## thanks
[songgao/water](https://github.com/songgao/water)

## more
[tun/tap vpn](https://github.com/ICKelin/article/issues/9)

any [issues](https://github.com/ICKelin/gtun/issues/new) are welcome

## TODO
- 测试内网穿透
- 测试新版本