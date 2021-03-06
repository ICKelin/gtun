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
- [功能特性](#功能特性)
- [技术原理](#技术原理)
- [安装部署](#安装部署)
  - [前期准备](#前期准备)
  - [安装运行gtund](#安装运行gtund)
  - [安装运行gtun](#安装运行gtun)
  - [配置加速ip](#配置加速ip)
  - [加速效果测试](#加速效果)
- [应用场景](#应用场景)
- [有问题怎么办](#有问题怎么办)
- [关于作者](#关于作者)

## 功能特性

- 纯应用层实现，不存在overlay网络，支持tcp和udp协议以及运行在其上的所有七层协议
- 支持ip加速，配合dnsmasq等软件可支持域名加速场景
- 支持动态和静态内容访问加速
- 引入`kcp`，`quic`等协议优化跨境传输

[返回目录](#目录)

## 技术原理

![](doc/assets/gtun.jpg)
<center><p>整体架构</p></center>
gtun是一款ip正向代理软件，包含代理客户端gtun和服务端gtund，如上图所示，gtun作为所有流量的入口，也即是正向代理的客户端，gtund作为所有流量的出口，也即是正向代理的服务端，gtun的客户端比较重，服务端程序则非常的轻量级。

gtun最主要的功能是流量代理，gtun经过三个版本的演变，最初基于tun网卡的vpn技术，然后优化到dnat技术，再到目前的tproxy技术，现已逐步趋于稳定。

gtun本身只提供流量代理通道，至于哪些流量需要被劫持，这个是由使用者定义的，使用者最终只需要将被代理的IP加入到`ipset`当中，那么该ipset的ip就会被代理

为了实现更加快速的代理，gtun考虑集成`kcp`或者`quic`等基于UDP实现的可靠性传输协议，以避免长链路tcp丢包严重触发拥塞控制机制，降低传输效率。

[返回目录](#目录)

## 安装部署
在这一节当中结合实际应用场景说明如何安装和部署gtun和gtund程序，在本应用场景当中，通过配置IP代理加速，加快访问speedtest的测速文件

### 前期准备

- 一台公有云服务器，用于部署服务端程序gtund，区域越靠近被加速区域（源站）越好，并且确认gtund监听的端口被打开
- 另外一台可以是公有云服务器，也可以是内网机器，用于部署客户端程序gtun，目前gtun只支持linux系统。

### 安装运行gtund
gtund需要运行在公有云上，相对比较简单，原则上越靠近源站越好。

首先生成配置文件，可以下载[gtund.yaml](https://github.com/ICKelin/gtun/blob/tproxy/etc/gtund.yaml)进行修改

```yaml
server:
  listen: ":9098"
  authKey: "rewrite with your auth key"

log:
  days: 5
  level: "info"
  path: "gtund.log"
```

大部分情况下您只需要修改`authKey`字段即可，配置文件生成之后，通过运行
`./gtund -c gtund.yaml`文件即可。

### 安装运行gtun
gtun可以运行在内网，也可以运行在公有云，在本场景当中，gtun会被部署在内网。

首先生成配置文件，可以下载[gtun.yaml](https://github.com/ICKelin/gtun/blob/tproxy/etc/gtun.yaml)进行修改

```yaml
forwards:
  CN: 
    server: "10.60.6.95:8524"
    authKey: "rewrite with your auth key"
    tcp:
      listen: ":8524"
    udp:
      listen: ":8525"

log:
  days: 5
  level: "info"
  path: "gtun.log"
```

- `forwards`配置了转发路线相关配置，key为转发的区域
- `server`字段配置了gtund所在机器的ip和端口
- `authKey`字段配置了gtun和gtund双方认证的key，在gtund的配置文件当中指定
- `tcp/udp`为tcp/udp代理相关配置，此处配置了tcp/udp代理监听的端口，所有需要代理加速的流量都会被重定向到该端口。

配置完成之后可以启动gtun程序，运行`./gtun -c gtun.yaml`即可启动。
### 配置加速ip
在上述过程中，启动了gtun和gtund程序，但是并未添加任何需要加速的信息，那么gtun如何进行加速呢？需要额外手动配置加速ip，并将该ip的tcp流量全部转发至`127.0.0.1:8524`端口，udp流量全部转发至`127.0.0.1:8525`端口。

这个过程是通过ipset和路由来配置的。以`1.1.1.1`为例

第一步，创建ipset，并将`1.1.1.1`加入其中

`ipset create GTUN-US hash:net`

`ipset add GTUN-US 1.1.1.1`

第二步，创建iptables规则，匹配目的ip为`GTUN-US`这个ipset内部的ip，然后做`tproxy`操作，将流量重定向到本地`8524`和`8525`端口

```
iptables -t mangle -I PREROUTING -p tcp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -I PREROUTING -p udp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -I OUTPUT -m set --match-set GTUN-US dst -j MARK --set-mark 1
```

第三步，添加路由表，确保数据包不被路由选择子系统丢弃

```
ip rule add fwmark 1 lookup 100
ip ro add local default dev lo table 100
```

至此所有配置都已经完成，后续需要新增代理ip，只使用以下命令将ip加入`GTUN-US`这个ipset当中即可，现在可以先尝试测试`1.1.1.1`这个ip的代理
```
root@raspberrypi:/home/pi# nslookup www.google.com 1.1.1.1
Server:		1.1.1.1
Address:	1.1.1.1#53

Non-authoritative answer:
Name:	www.google.com
Address: 142.250.73.228
```

### 加速效果

有了上述的基础，现在可以进行下载速度测试对比，以`http://speedtest.atlanta.linode.com/100MB-atlanta.bin`这个文件作为测试，

首先是通过gtun代理加速之后的测试，需要将`speedtest.atlanta.linode.com`加入到GTUN-US当中

`ipset add GTUN-US speedtest.atlanta.linode.com`

```shell
root@raspberrypi:/home/pi# wget http://speedtest.atlanta.linode.com/100MB-atlanta.bin -v
--2021-05-18 22:00:23--  http://speedtest.atlanta.linode.com/100MB-atlanta.bin
正在解析主机 speedtest.atlanta.linode.com (speedtest.atlanta.linode.com)... 50.116.39.117, 2600:3c02::f03c:91ff:feae:641
正在连接 speedtest.atlanta.linode.com (speedtest.atlanta.linode.com)|50.116.39.117|:80... 已连接。
已发出 HTTP 请求，正在等待回应... 200 OK
长度：104857600 (100M) [application/octet-stream]
正在保存至: “100MB-atlanta.bin”

100MB-atlanta.bin                   100%[==================================================================>] 100.00M  2.39MB/s    in 57s

2021-05-18 22:01:21 (1.77 MB/s) - 已保存 “100MB-atlanta.bin” [104857600/104857600])
```

然后通过正常网络测试，将`speedtest.atlanta.linode.com`从GTUN-US当中移除即可

`ipset del GTUN-US speedtest.atlanta.linode.com`

```
root@raspberrypi:/home/pi# wget http://speedtest.atlanta.linode.com/100MB-atlanta.bin -v
--2021-05-18 22:04:44--  http://speedtest.atlanta.linode.com/100MB-atlanta.bin
正在解析主机 speedtest.atlanta.linode.com (speedtest.atlanta.linode.com)... 50.116.39.117, 2600:3c02::f03c:91ff:feae:641
正在连接 speedtest.atlanta.linode.com (speedtest.atlanta.linode.com)|50.116.39.117|:80... 已连接。
已发出 HTTP 请求，正在等待回应... 200 OK
长度：104857600 (100M) [application/octet-stream]
正在保存至: “100MB-atlanta.bin.1”

100MB-atlanta.bin.1                   0%[                                                                   ]   1012K  9.50KB/s    eta 98m 58s
```

可以看见，通过gtun加速之后，速度可以达到2.39MB/s，而未通过gtun加速的正常下载速度则为15KB/s左右的速度，两者差了一个数量级。

## 应用场景

- IP加速，可用于ip，子网加速
- 域名，站点加速，需要使用dnsmasq或者nginx/openresty等组件实现
- k8s集群网络代理，ip加速的一个子集，可代理访问k8s的service，pod网段
- 全球应用加速

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