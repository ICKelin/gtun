## gtun
<a href="">
<img src="https://img.shields.io/badge/-Go-000?&logo=go">
</a>
<a href="https://goreportcard.com/report/github.com/ICKelin/gtun" rel="nofollow">
<img src="https://goreportcard.com/badge/github.com/ICKelin/gtun" alt="go report">
</a>

<a href="https://travis-ci.org/ICKelin/gtun" rel="nofollow">
<img src="https://travis-ci.org/ICKelin/gtun.svg?branch=master" alt="Build Status">
</a>
<a href="https://github.com/ICKelin/gtun/blob/master/LICENSE">
<img src="https://img.shields.io/github/license/mashape/apistatus.svg" alt="license">
</a>

gtun是一款开源的ip代理加速软件，通过`tproxy`技术实现流量劫持，`quic`和`kcp`等协议优化广域网传输，gtun提供一个基础通道，所有加入`ipset`的ip，出口，入口流量都会被gtun进行拦截并代理到指定出口。

gtun支持多线路配置，可以同时对美国，日本，欧洲目的网络进行加速访问。

你可以使用gtun来作为一个ip加速器，加速访问位于海外的跳板机，云服务器，网站。

gtun对标阿里云的全球应用加速，ucloud的pathX等产品的功能，对此有任何问题需要咨询，可以[联系作者](#关于作者)进行交流

## 目录
- [介绍](#gtun)
- [技术原理](#技术原理)
- [功能特性](#功能特性)
- [应用场景](#应用场景)
   - [IP加速]()
   - [域名加速]()
   - [全球应用加速]()
- [安装部署](#安装部署)
- [有问题怎么办](#有问题怎么办)
- [关于作者](#关于作者)

## 功能特性

- 纯应用层实现，不存在overlay网络，支持tcp和udp协议以及运行在其上的所有七层协议
- 支持ip加速，配合dnsmasq等软件可支持域名加速场景
- 引入`kcp`，`quic`等协议优化跨境传输

[返回目录](#目录)

## 技术原理

![](doc/assets/gtun.jpg)
<center><p>整体架构</p></center>
gtun是一款ip正向代理软件，包含代理客户端gtun和服务端gtund，如上图所示，gtun作为所有流量的入口，也即是正向代理的客户端，gtund作为所有流量的出口，也即是正向代理的服务端。

gtun最主要的功能是流量代理，gtun经过三个版本的演变，最初基于tun网卡的vpn技术，然后优化到dnat技术，再到目前的tproxy技术，现已逐步趋于稳定。

gtun本身只提供流量代理通道，至于哪些流量需要被劫持，这个是由使用者定义的，使用者最终只需要将被代理的IP加入到`ipset`当中，那么该ipset的ip就会被代理

为了实现更加快速的代理，gtun考虑集成`kcp`或者`quic`等基于UDP实现的可靠性传输协议，以避免长链路tcp丢包严重触发拥塞控制机制，降低传输效率。

[返回目录](#目录)

## 应用场景

## 安装部署

## 有问题怎么办

- [wiki](https://github.com/ICKelin/gtun/wiki)
- [查看文档]()
- [提交issue](https://github.com/ICKelin/gtun/issues)
- [查看源码](https://github.com/ICKelin/gtun)
- [联系作者交流解决](#关于作者)

[返回目录](#目录)

## 关于作者
一个爱好编程的人，网名叫ICKelin。对于以下任何问题，包括

- 项目实现细节
- 项目使用问题
- 项目建议，代码问题
- 案例分享
- 技术交流

可加微信: zyj995139094