# Next Terminal

![Docker image](https://github.com/ysicing/next-terminal/workflows/Docker%20image/badge.svg?branch=ysicing)

## 快速了解

Next Terminal是使用Golang和React开发的一款HTML5的远程桌面网关，具有小巧、易安装、易使用、资源占用小的特点，支持RDP、SSH、VNC和Telnet协议的连接和管理。

Next Terminal基于 [Apache Guacamole](https://guacamole.apache.org/) 开发，使用到了guacd服务。

目前支持的功能有：

- 授权凭证管理
- 资产管理（支持RDP、SSH、VNC、TELNET协议）
- 指令管理
- 批量执行命令
- 在线会话管理（监控、强制断开）
- 离线会话管理（查看录屏）
- 双因素认证
- 资产标签
- 资产授权
- 多用户&用户分组
- 计划任务

## 在线体验

> 每隔60分钟重启更新镜像并&清理数据

https://next-terminal-ysicing.cloud.okteto.net

admin/admin

## 协议与条款

- 本项目不提供任何担保，亦不承担任何责任。
- 遵循GPLv2协议

## 快速安装

推荐k8s中部署, 参考[k8s](./hack/k8s/nt.yaml)

默认账号密码为 admin/admin