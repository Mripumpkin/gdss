# 📂 GDSS FileServer

GDSS FileServer 是一个基于对等网络 (P2P) 的文件存储和分发系统，支持文件的加密存储、网络传输以及节点间协同同步。

## ✨ 功能特性

* 基于 P2P 的文件分发网络
* 支持本地磁盘加密存储
* 支持加密文件的上传 (Store) 和下载 (Get)
* 自动广播文件变更事件
* 简单易用的插件式 Transport 模块
* 支持自定义路径转换函数 (PathTransformFunc)

## 📁 项目结构

```
gdss/
├── server        ：文件服务核心实现
├── store/         ：文件加密存储模块
├── p2p/           ：网络通信模块（如 Transport、Peer）
├── gcrypto/       ：ID 生成与加密/解密逻辑
├── log/           ：日志封装
```

## 🚀 快速开始

```bash
make run
```

## 🧐 设计说明

* **加密存储**：写入时调用 `gcrypto.CopyEncrypt` 进行数据加密，读取时自动解密
* **P2P 广播**：通过广播机制向网络中其他节点通知文件变更
* **Gob 协议**：消息通过 Go 的 `gob` 序列化机制在网络上传输
* **网络读取流控制**：使用 `IncomingStream` 标识数据流传输起点

## 🥮 测试建议

* 启动多个节点，指定不同的 ID 和 StorageRoot
* 模拟节点掉线、恢复、重新同步
* 用不同的文件、并发上传下载进行压力测试

## ⚠️ 注意事项

* `EncKey` 长度需满足加密算法要求（例如 AES-256 要求 32 字节）
* 请确保你的 Transport 实现支持 `Send`、`RemoteAddr`、`CloseStream` 等接口
* 当前网络通信简单实现，未做完整的错误处理和重试机制
