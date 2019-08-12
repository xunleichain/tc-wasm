English | [简体中文](./README.zh-CN.md)

tc-wasm
=====

[![License Apache 2](https://img.shields.io/badge/license-apache%202-blue.svg?style=flat-square)](https://github.com/xunleichain/tc-wasm/blob/master/LICENSE)

Thunder chain wasm virtual machine. Easy to integrate into other  
blockchain platforms, support smart contracts written in c/c++.

**NOTE:** `tc-wasm` requires `Go >= 1.12.x`.

## Try
1. cd cmd/tcvm/; go build (Here we get the executable command `tcvm`)
2. Open the web page https://catalyst.onethingcloud.com/#/catalyst,  
where you can write a c/c++ contract and copy the compiled wasm bytecode
3. Save the bytecode to the file contract.wasm, and save the called  
method and parameters to the file contract.params
4. `tcvm` -file contract.wasm -call contract.params
5. Try modifying `cmd/tcvm/main.go`, repeat steps 1-4, observe and verify  
the results.

## Code organization
| Directory | Description |
|-----|-----|
| /cmd | A binary tool for quick and easy testing of wasm contracts |
| /vm | Virtual machine |
| /mock | The data structure that needs to be implemented when integrating this<br>virtual machine. Part of the code is derived from Ethereum |
| /mock/deps | Basic dependencies copied from the Ethereum corresponding directory |
| /mock/log | Implement the log interface |
| /mock/state | Implement global account status interface |
| /mock/types | Implement some basic types of blockchain, such as: address, hash,<br>block header, etc |
| /testdata | Test contract code and data |
