[English](./README.md) | 简体中文

tc-wasm
=====

[![License Apache 2](https://img.shields.io/badge/license-apache%202-blue.svg?style=flat-square)](https://github.com/xunleichain/tc-wasm/blob/master/LICENSE)

迅雷链wasm虚拟机. 方便集成到其它区块链平台, 支持c/c++编写的智能合约.

**NOTE:** `tc-wasm` requires `Go >= 1.12.x`.

## 尝试
1. 进入目录cmd/tcvm/, 执行`go build` (这里得到可执行命令tcvm)
2. 打开网页https://catalyst.onethingcloud.com/#/catalyst, 在上面可以  
编写c/c++合约并拷贝编译后的wasm字节码
3. 将字节码保存到文件contract.wasm, 将调用的方法和参数存入文件contract.params
4. 执行tcvm -file contract.wasm -call contract.params
5. `尝试修改cmd/tcvm/main.go, 重复上面步骤1-4, 观察并验证修改结果`

## 源码组织
| 目录 | 说明 |
|-----|-----|
| /cmd | 一个方便快速测试wasm合约的二进制工具 |
| /vm | 虚拟机代码 |
| /mock | 集成该虚拟机时需要实现的数据结构, 部分代码基于以太坊作的修改 |
| /mock/deps | 基本依赖, 从以太坊对应目录拷贝过来 |
| /mock/log | 实现虚拟机使用的日志接口 |
| /mock/state | 实现虚拟机使用全局账户状态接口 |
| /mock/types | 实现区块链的一些基本类型, 如: 地址、哈希、区块头 等 |
| /testdata | 测试代码及数据 |
