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

gtun支持多线路配置，可以同时对美国，日本，欧洲目的网络进行加速访问。您可以结合dnsmasq来使用，将需要配置加速的域名解析结果加入ipset，从而实现域名加速。

[![](https://res.cloudinary.com/marcomontalbano/image/upload/v1686622903/video_to_markdown/images/youtube--pxv02e5EXPE-c05b58ac6eb4c4700831b2b3070cd403.jpg)](https://www.youtube.com/watch?v=pxv02e5EXPE "")

**使用场景**

- SaaS软件加速，加速访问Salesforce，offce365等产品
- 云服务器加速，加速访问海外服务器，跳板机，提升操作流畅度
- 直播加速，tiktok海外直播加速，抖音直播加速
- 游戏加速，结合专线网络和路由盒子实现游戏加速盒

gtun是一个完整的加速器，**目前只支持linux**

同时我们也基于gtun开发了收费版本，对标阿里云的全球应用加速，ucloud的pathX等产品的功能，只是会更加灵活，支持私有化部署，独立部署，可以下沉到办公室，如果您感兴趣，可以访问[我们的网站](https://www.beyondnetwork.net)进行免费免费体验。

关于项目有任何问题需要咨询，可以[联系作者](#关于作者)进行交流

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
- 支持多链路容灾和竞速
- 支持动态和静态内容访问加速
- 引入`kcp`，`quic`等协议优化跨境传输（quic进行中）

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

首先生成配置文件，可以下载 [gtund.yaml](https://github.com/ICKelin/gtun/blob/master/etc/gtund.yaml) 进行修改

```yaml
trace: ":3003"
server:
  - listen: ":3002"
    authKey: "rewrite with your auth key"
    scheme: "kcp"

log:
  days: 5
  level: "debug"
  path: "gtund.log"
```

大部分情况下，如果您的端口未被占用，不需要修改任何配置
`./gtund -c gtund.yaml`文件即可。

### 安装运行gtun
gtun可以运行在内网，也可以运行在公有云，在本场景当中，gtun会被部署在内网。

首先生成配置文件，可以下载 [gtun.yaml](https://github.com/ICKelin/gtun/blob/master/etc/gtun.yaml) 进行修改

```yaml
settings:
  US:
    # 代理ip文件，可以是本地文件，也可以是网络文件，一行是一个IP或者cidr
    proxy_file: "https://www.ipdeny.com/ipblocks/data/countries/us.zone"
    route:
      # 拨测地址，需要修改US_SERVER_IP和US_SERVER_TRACE_PORT，对应gtund的公网IP和端口
      - trace_addr: ${US_SERVER_IP}:${US_SERVER_TRACE_PORT}
        scheme: "kcp"
        # 服务端地址，修改为对应gtund的IP和端口
        addr: ${US_SERVER_IP}:${US_SERVER_PORT}
        auth_key: "rewrite with your auth key"
    proxy:
      # 代理插件配置
      "tproxy_tcp": |
        {
          "read_timeout": 30,
          "write_timeout": 30,
          "listen_addr": ":8524",
          "rate_limit": 50,
          "region": "US"
        }
      "tproxy_udp": |
        {
          "read_timeout": 30,
          "write_timeout": 30,
          "session_timeout": 30,
          "listen_addr": ":8524",
          "rate_limit": 50,
          "region": "US"
        }
log:
  days: 5
  level: Debug
  path: gtun.log

http_server:
  listen_addr: ":9001"
```

配置完成之后可以启动gtun程序，运行`./gtun -c gtun.yaml`即可启动。

gtund启动时，会自动设置 iptables规则和路由表，并将需要加速的ip加入ipset当中，如果ip量比较大，启动时间会稍微长一些。

[返回目录](#目录)

### 配置加速ip
目前支持两种方式配置IP：

- 基于接口的方式，我们提供HTTP接口进行动态增删IP，目前正在开发页面配置动态管理加速的IP，应用，域名，敬请期待。
- 使用命令手动配置，手动将ip加入到ipset当中

接下来以命令配置的方式进行配置，以`1.1.1.1`为例，只需要将`1.1.1.1`加入其中ipset当中`ipset add GTUN-US 1.1.1.1`即可。
```
root@raspberrypi:/home/pi# nslookup www.google.com 1.1.1.1
Server:		1.1.1.1
Address:	1.1.1.1#53

Non-authoritative answer:
Name:	www.google.com
Address: 142.250.73.228
```

[返回目录](#目录)

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

[返回目录](#目录)

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