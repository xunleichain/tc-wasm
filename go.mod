module tc-wasm

go 1.12

replace (
	github.com/go-interpreter/wagon => github.com/xunleichain/wagon v0.5.1
	github.com/xunleichain/tc-wasm => ./
)

require (
	github.com/go-interpreter/wagon v0.0.0
	github.com/xunleichain/tc-wasm v0.0.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
)
