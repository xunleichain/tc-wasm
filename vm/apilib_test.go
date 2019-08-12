package vm

import (
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/xunleichain/tc-wasm/mock/log"
	"github.com/xunleichain/tc-wasm/mock/state"
	"github.com/xunleichain/tc-wasm/mock/types"
)

var (
	cState  *state.StateDB
	cAddr   types.Address
	ctxTime uint64
)

func init() {
	cState, _ = state.New()
	cState.AddBalance(cAddr, big.NewInt(int64(10000)))

	cAddr = types.BytesToAddress([]byte{1})
	ctxTime = 1565078742
}

func TestCallContract(t *testing.T) {
	wasmContractFile1 := "../testdata/contract.wasm"
	contractCode1, err := ioutil.ReadFile(wasmContractFile1)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	wasmContractFile2 := "../testdata/contract1.wasm"
	contractCode2, err := ioutil.ReadFile(wasmContractFile2)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	wasmContractFile3 := "../testdata/contract2.wasm"
	contractCode3, err := ioutil.ReadFile(wasmContractFile3)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	addr1 := types.BytesToAddress([]byte{114})
	cState.SetCode(addr1, contractCode1)

	addr2 := types.BytesToAddress([]byte{115})
	cState.SetCode(addr2, contractCode2)

	addr3 := types.BytesToAddress([]byte{116})
	cState.SetCode(addr3, contractCode3)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr1),
		CodeAddr:      &addr1,
		value:         big.NewInt(100),
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr1,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 1000000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr1.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	action := "none"
	params := "{\"contract1\":\"0x0000000000000000000000000000000000000073\",\"contract2\":\"0x0000000000000000000000000000000000000074\"}"
	input := make([]byte, len(action)+len(params)+5)
	copy(input[0:], wasmBytes[0:4])
	copy(input[4:], action)
	copy(input[4+len(action):], []byte{'|'})
	copy(input[5+len(action):], params)
	eng.contract.Input = input[4:]
	ret, err := eng.Run(app, input)
	t.Logf("ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestNotify(t *testing.T) {
	wasmFile := "../testdata/notify.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{113})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	logs := eng.stateDB.Logs()
	for i := 0; i < len(logs); i++ {
		t.Logf("log %d %s", i, logs[i].String())
	}
	return
}

func TestToken(t *testing.T) {
	wasmFile := "../testdata/token.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{112})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("strlen ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestMalloc(t *testing.T) {
	wasmFile := "../testdata/malloc.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{111})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("malloc ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestPrints(t *testing.T) {
	wasmFile := "../testdata/prints.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{110})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("prints ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestLog(t *testing.T) {
	wasmFile := "../testdata/log.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{109})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	gEnvTable.RegisterFunc("TC_Log0", new(TCLog0))
	gEnvTable.RegisterFunc("TC_Log1", new(TCLog1))
	gEnvTable.RegisterFunc("TC_Log2", new(TCLog2))
	gEnvTable.RegisterFunc("TC_Log3", new(TCLog3))
	gEnvTable.RegisterFunc("TC_Log4", new(TCLog4))

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("log ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	logs := eng.stateDB.Logs()
	for i := 0; i < len(logs); i++ {
		t.Logf("log %d %s", i, logs[i].String())
	}
	return
}

func TestSelfDestruct(t *testing.T) {
	wasmFile := "../testdata/selfdestruct.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{108})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	t.Logf("from account balance: %d before exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d before exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	t.Logf("from account code: 0x%x before exec contract method", cState.GetCode(addr))
	t.Logf("from account cache code: %v before exec contract method", eng.appByName(addr.String()))

	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)

	t.Logf("selfdestruct ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	t.Logf("from account balance: %d after exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d after exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	t.Logf("from account code: 0x%x after exec contract method", cState.GetCode(addr))
	t.Logf("from account cache code: %v after exec contract method", eng.appByName(addr.String()))
	return
}

func TestSelfAddress(t *testing.T) {
	wasmFile := "../testdata/selfaddress.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{107})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("getSelfAddress ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestGetBalance(t *testing.T) {
	wasmFile := "../testdata/getbalance.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{106})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("getBalance ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestEcrecover(t *testing.T) {
	wasmFile := "../testdata/ecrecover.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{105})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ecrecover ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestRipemd160(t *testing.T) {
	wasmFile := "../testdata/ripemd160.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{103})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ripemd160 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestSha256(t *testing.T) {
	wasmFile := "../testdata/sha256.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{102})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("sha256 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestKeccak256(t *testing.T) {
	wasmFile := "../testdata/keccak256.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{101})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("keccak256 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestTransfer(t *testing.T) {
	wasmFile := "../testdata/transfer.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{100})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	t.Logf("from account balance: %d before exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d before exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	contract := Contract{
		CallerAddress: cAddr,
		caller:        AccountRef(cAddr),
		self:          AccountRef(addr),
		CodeAddr:      &addr,
	}
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := NewEngine(cState, 100000, contract, log.Test(), ctx)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("transfer ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	t.Logf("from account balance: %d after exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d after exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	return
}
