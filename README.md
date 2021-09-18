# Next Terminal

[![Docker NT image](https://github.com/ysicing/next-terminal/actions/workflows/docker.yml/badge.svg?branch=ysicing)](https://github.com/ysicing/next-terminal/actions/workflows/docker.yml)

## 版本变更

[Changelog](./CHANGELOG.md)

## 在线体验

> 每隔60分钟重启更新镜像并&清理数据

https://next-terminal-ysicing.cloud.okteto.net

admin/admin

## 协议与条款

- 本项目不提供任何担保，亦不承担任何责任。
- 遵循GPLv2协议

## 本地开发

```bash
git clone https://github.com/ysicing/next-terminal.git
cd next-terminal
# 配置好文件
make static
air
```

## 快速安装

推荐k8s中部署, 参考[k8s](./hack/k8s/nt.yaml)

默认账号密码为 admin/admin