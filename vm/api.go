package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/xunleichain/tc-wasm/mock/types"
)

type Args struct {
	Params []Param `json:"Params"`
}

type Param struct {
	Ptype string `json:"type"`
	Pval  string `json:"value"`
}

type Result struct {
	Ptype    string `json:"type"`
	Pval     string `json:"value"`
	Psucceed int    `json:"succeed"`
}

//trim the '\00' byte
func TrimBuffToString(bytes []byte) string {

	for i, b := range bytes {
		if b == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)

}

type TCMemcpy struct{}

func (t *TCMemcpy) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcMemcpy(eng, index, args)
}
func (t *TCMemcpy) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasMemcpy(eng, index, args)
}

//c: void* memcpy(void * dest, const void * src, size_t length)
func tcMemcpy(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dest := args[0]
	src := args[1]
	length := int(args[2])
	return vmem.Memcpy(dest, src, length)
}

type TCMemset struct{}

func (t *TCMemset) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcMemset(eng, index, args)
}
func (t *TCMemset) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasMemset(eng, index, args)
}

//c: void *memset(void *str, int c, size_t n)
func tcMemset(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dest := args[0]
	c := byte(args[1])
	length := int(args[2])
	return vmem.Memset(dest, c, length)
}

type TCMemmove struct{}

func (t *TCMemmove) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcMemmove(eng, index, args)
}
func (t *TCMemmove) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasMemmove(eng, index, args)
}

//c: void *memmove(void *str1, const void *str2, size_t n)
func tcMemmove(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dest := args[0]
	src := args[1]
	length := int(args[2])
	return vmem.Memmove(dest, src, length)
}

type TCMemcmp struct{}

func (t *TCMemcmp) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcMemcmp(eng, index, args)
}
func (t *TCMemcmp) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasMemcmp(eng, index, args)
}

//c: int memcmp(const void *str1, const void *str2, size_t n))
func tcMemcmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	str1 := args[0]
	str2 := args[1]
	length := int(args[2])
	ret, err := vmem.Memcmp(str1, str2, length)
	return uint64(ret), err
}

type TCStrcmp struct{}

func (t *TCStrcmp) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcStrcmp(eng, index, args)
}
func (t *TCStrcmp) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasStrcmp(eng, index, args)
}

//c: int strcmp(const char *str1, const char *str2)
func tcStrcmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	str1 := args[0]
	str2 := args[1]
	ret, err := vmem.Strcmp(str1, str2)
	return uint64(ret), err
}

type TCStrcpy struct{}

func (t *TCStrcpy) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcStrcpy(eng, index, args)
}
func (t *TCStrcpy) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasStrcpy(eng, index, args)
}

//c: char *strcpy(char *dest, const char *src)
func tcStrcpy(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dest := args[0]
	src := args[1]
	return vmem.Strcpy(dest, src)
}

type TCStrconcat struct{}

func (t *TCStrconcat) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcStrconcat(eng, index, args)
}
func (t *TCStrconcat) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasStrconcat(eng, index, args)
}

//c: char * strconcat(char *a,char *b)
func tcStrconcat(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	str1, err := vmem.GetString(args[0])
	if err != nil {
		return 0, err
	}

	str2, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	newString := TrimBuffToString(str1) + TrimBuffToString(str2)

	idx, err := vmem.SetBytes([]byte(newString))
	if err != nil {
		return 0, err
	}
	eng.Logger().Debug("WASM RUN DebugLOG:call stringconcat", "str1", string(str1), "str2", string(str2), "retstr", newString)
	return uint64(idx), nil
}

type TCAtoi struct{}

func (t *TCAtoi) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAtoi(eng, index, args)
}
func (t *TCAtoi) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAtoi(eng, index, args)
}

//c: int Atoi(char * s)
func tcAtoi(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	addr := args[0]

	pBytes, err := vmem.GetString(addr)
	if err != nil {
		return 0, errors.New("GetString err:" + err.Error())
	}
	if pBytes == nil || len(pBytes) == 0 {
		return 0, nil
	}

	str := TrimBuffToString(pBytes)
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.New("Atoi err:" + err.Error())
	}

	return uint64(int32(i)), nil
}

type TCAtof64 struct{}

func (t *TCAtof64) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAtof64(eng, index, args)
}
func (t *TCAtof64) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAtof64(eng, index, args)
}

//c: int Atof64(char * s)
func tcAtof64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	addr := args[0]

	pBytes, err := vmem.GetString(addr)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] GetString err:" + err.Error())
	}
	if pBytes == nil || len(pBytes) == 0 {
		return 0, nil
	}

	str := TrimBuffToString(pBytes)
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] Atoi err:" + err.Error())
	}
	return math.Float64bits(i), nil
}

type TCAtof32 struct{}

func (t *TCAtof32) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAtof32(eng, index, args)
}
func (t *TCAtof32) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAtof32(eng, index, args)
}

//c: int Atof32(char * s)
func tcAtof32(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	addr := args[0]

	pBytes, err := vmem.GetString(addr)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] GetString err:" + err.Error())
	}
	if pBytes == nil || len(pBytes) == 0 {
		return 0, nil
	}

	str := TrimBuffToString(pBytes)
	i, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] Atoi err:" + err.Error())
	}

	return uint64(math.Float32bits(float32(i))), nil
}

type TCAtoi64 struct{}

func (t *TCAtoi64) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAtoi64(eng, index, args)
}
func (t *TCAtoi64) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAtoi64(eng, index, args)
}

//c: long long Atoi64(char *s)
func tcAtoi64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	addr := args[0]

	pBytes, err := vmem.GetString(addr)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] GetString err:" + err.Error())
	}

	if pBytes == nil || len(pBytes) == 0 {
		return 0, nil
	}

	str := TrimBuffToString(pBytes)
	i, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return 0, errors.New("[jsonMashalParams] Atoi err:" + err.Error())
	}
	eng.Logger().Debug("WASM RUN DebugLOG:call strToInt64", "str", str, "ret", i)
	return uint64(i), nil

}

type TCItoa struct{}

func (t *TCItoa) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcItoa(eng, index, args)
}
func (t *TCItoa) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasItoa(eng, index, args)
}

//c: char * Itoa(int a)
func tcItoa(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	i := int32(args[0])
	str := strconv.Itoa(int(i))
	idx, err := vmem.SetBytes([]byte(str))
	if err != nil {
		return 0, err
	}
	return uint64(idx), nil
}

type TCI64toa struct{}

func (t *TCI64toa) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcI64toa(eng, index, args)
}
func (t *TCI64toa) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasI64toa(eng, index, args)
}

//c: char * I64toa(long long amount,int radix)
func tcI64toa(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	i := int64(args[0])
	radix := int(args[1])
	str := strconv.FormatInt(i, radix)
	idx, err := vmem.SetBytes([]byte(str))
	if err != nil {
		return 0, err
	}
	eng.Logger().Debug("WASM RUN DebugLOG:call int64ToString",
		"amount", i, "radix", radix, "ret", str)
	return uint64(idx), nil
}

type bigIntOpType uint8

const (
	bigIntOpAdd bigIntOpType = iota
	bigIntOpSub
	bigIntOpMul
	bigIntOpDiv
	bigIntOpMod
	bigIntOpCmp
)

var (
	base16Prefix1 = []byte("0x")
	base16Prefix2 = []byte("0X")
	base8Prefix   = []byte("0")
	base2Prefix1  = []byte("0b")
	base2Prefix2  = []byte("0B")
)

func parseBase(b []byte) int {
	if bytes.HasPrefix(b, base16Prefix1) || bytes.HasPrefix(b, base16Prefix2) {
		return 16
	}

	if bytes.HasPrefix(b, base8Prefix) {
		return 8
	}

	if bytes.HasPrefix(b, base2Prefix1) || bytes.HasPrefix(b, base2Prefix2) {
		return 2
	}

	return 10
}

// helper function for BigInt operation
func tcBigIntOp(eng *Engine, args []uint64, op bigIntOpType) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	a, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrMemoryGet
	}
	b, err := vmem.GetString(args[1])
	if err != nil {
		return 0, ErrMemoryGet
	}

	aBigInt, e1 := new(big.Int).SetString(string(a), 0)
	bBigInt, e2 := new(big.Int).SetString(string(b), 0)
	if !e1 {
		aBigInt = big.NewInt(0)
	}
	if !e2 {
		bBigInt = big.NewInt(0)
	}

	switch op {
	case bigIntOpAdd:
		aBigInt = aBigInt.Add(aBigInt, bBigInt)
	case bigIntOpSub:
		aBigInt = aBigInt.Sub(aBigInt, bBigInt)
	case bigIntOpMul:
		aBigInt = aBigInt.Mul(aBigInt, bBigInt)
	case bigIntOpDiv:
		aBigInt = aBigInt.Div(aBigInt, bBigInt)
	case bigIntOpMod:
		aBigInt = aBigInt.Mod(aBigInt, bBigInt)
	case bigIntOpCmp:
		cmpret := aBigInt.Cmp(bBigInt)

		return uint64(cmpret), nil
	}

	sumPointer, err := vmem.SetBytes([]byte(aBigInt.String()))
	if err != nil {
		return 0, ErrMemorySet
	}

	return uint64(sumPointer), nil
}

type TCBigIntAdd struct{}

func (t *TCBigIntAdd) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntAdd(eng, index, args)
}
func (t *TCBigIntAdd) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntAdd(eng, index, args)
}

// c: char *TC_BigIntAdd(char *a, char *b)
func tcBigIntAdd(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpAdd)
}

type TCBigIntSub struct{}

func (t *TCBigIntSub) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntSub(eng, index, args)
}
func (t *TCBigIntSub) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntSub(eng, index, args)
}

// c: char *TC_BigIntSub(char *a, char *b)
func tcBigIntSub(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpSub)
}

type TCBigIntMul struct{}

func (t *TCBigIntMul) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntMul(eng, index, args)
}
func (t *TCBigIntMul) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntMul(eng, index, args)
}

// c: char *TC_BigIntMul(char *a, char *b)
func tcBigIntMul(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpMul)
}

type TCBigIntDiv struct{}

func (t *TCBigIntDiv) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntDiv(eng, index, args)
}
func (t *TCBigIntDiv) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntDiv(eng, index, args)
}

// c: char *TC_BigIntDiv(char *a, char *b)
func tcBigIntDiv(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpDiv)
}

type TCBigIntMod struct{}

func (t *TCBigIntMod) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntMod(eng, index, args)
}
func (t *TCBigIntMod) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntMod(eng, index, args)
}

// c: char *TC_BigIntMod(char *a, char *b)
func tcBigIntMod(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpMod)
}

type TCBigIntCmp struct{}

func (t *TCBigIntCmp) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntCmp(eng, index, args)
}
func (t *TCBigIntCmp) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntCmp(eng, index, args)
}

// c: int TC_BigIntCmp(char *a, char *b)
func tcBigIntCmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	return tcBigIntOp(eng, args, bigIntOpCmp)
}

type TCBigIntToInt64 struct{}

func (t *TCBigIntToInt64) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcBigIntToInt64(eng, index, args)
}
func (t *TCBigIntToInt64) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasBigIntToInt64(eng, index, args)
}

// c: int64_t TC_BigIntToInt64(char *a)
func tcBigIntToInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	a, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrMemoryGet
	}

	aBigInt, _ := new(big.Int).SetString(string(a), 0)
	if aBigInt == nil {
		return 0, fmt.Errorf("invalid args")
	}

	return uint64(aBigInt.Int64()), nil
}

type TCGetMsgData struct{}

func (t *TCGetMsgData) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetMsgData(eng, index, args)
}
func (t *TCGetMsgData) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetMsgData(eng, index, args)
}

// char *TC_get_msg_data()
func tcGetMsgData(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	data := eng.Contract.Input

	dataPtr, err := vmem.SetBytes(data)
	if err != nil {
		return 0, ErrMemorySet
	}
	return uint64(dataPtr), nil
}

type TCGetMsgGas struct{}

func (t *TCGetMsgGas) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetMsgGas(eng, index, args)
}
func (t *TCGetMsgGas) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetMsgGas(eng, index, args)
}

// long long TC_get_msg_gas()
func tcGetMsgGas(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	gas := eng.Gas()

	return gas, nil
}

type TCGetMsgSender struct{}

func (t *TCGetMsgSender) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetMsgSender(eng, index, args)
}
func (t *TCGetMsgSender) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetMsgSender(eng, index, args)
}

// char *TC_get_msg_sender()
func tcGetMsgSender(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	sender := eng.Contract.CallerAddress
	senderStr := sender.String()

	dataPtr, err := vmem.SetBytes([]byte(senderStr))
	if err != nil {
		return 0, ErrMemorySet
	}
	return uint64(dataPtr), nil
}

type TCGetMsgSign struct{}

func (t *TCGetMsgSign) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetMsgSign(eng, index, args)
}
func (t *TCGetMsgSign) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetMsgSign(eng, index, args)
}

// char *TC_get_msg_sig()
// return calldata funcname
func tcGetMsgSign(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	input := eng.Contract.Input
	arr := bytes.Split(input, []byte("|"))
	lenAction := len(arr[0])
	action := string(input[:lenAction])

	dataPtr, err := vmem.SetBytes([]byte(action))
	if err != nil {
		return 0, ErrMemorySet
	}
	return uint64(dataPtr), nil
}

type TCAssert struct{}

func (t *TCAssert) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAssert(eng, index, args)
}
func (t *TCAssert) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAssert(eng, index, args)
}

// void TC_assert(bool condition)
func tcAssert(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidApiArgs
	}
	condition := int(args[0])
	if condition == 0 {
		eng.Logger().Debug("WASM RUN LOG:call TC_assert")
		return 0, ErrExecutionReverted
	}
	return 0, nil
}

type TCExit struct{}

func (t *TCExit) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcExit(eng, index, args)
}
func (t *TCExit) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasExit(eng, index, args)
}

// void exit(int)
func tcExit(eng *Engine, index int64, args []uint64) (uint64, error) {
	//Returned is a number, not a string address
	if len(args) != 1 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	app.VmProcess.Terminate()
	eng.Logger().Debug("WASM RUN LOG:exit")
	return args[0], nil
}

type TCAbort struct{}

func (t *TCAbort) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcAbort(eng, index, args)
}
func (t *TCAbort) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasAbort(eng, index, args)
}

// void abort()
func tcAbort(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	// app, _ := eng.RunningAppFrame()
	// app.VmProcess.Terminate()
	eng.Logger().Debug("WASM RUN LOG:call c abort")
	return 0, ErrContractAbort
}

type TCRequire struct{}

func (t *TCRequire) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRequire(eng, index, args)
}
func (t *TCRequire) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRequire(eng, index, args)
}

// void TC_require(bool condition)
func tcRequire(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidApiArgs
	}
	condition := int(args[0])
	if condition == 0 {
		// app, _ := eng.RunningAppFrame()
		// app.VmProcess.Terminate()
		eng.Logger().Debug("WASM RUN LOG:call TC_require")

		return 0, ErrExecutionReverted
	}
	return 0, nil
}

type TCGasLeft struct{}

func (t *TCGasLeft) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGasLeft(eng, index, args)
}
func (t *TCGasLeft) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGasLeft(eng, index, args)
}

// long long TC_gasleft()
func tcGasLeft(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	gas := eng.Gas()

	return gas, nil
}

type TCRequireWithMsg struct{}

func (t *TCRequireWithMsg) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRequireWithMsg(eng, index, args)
}
func (t *TCRequireWithMsg) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRequireWithMsg(eng, index, args)
}

// void TC_requireWithMsg(bool condition, char *msg)
func tcRequireWithMsg(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 2 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	condition := int(args[0])
	a, err := vmem.GetString(args[1])
	if err != nil {
		return 0, ErrMemoryGet
	}
	msg := string(a)

	if condition == 0 {
		//TODO: write log
		// eng.State.AddLog()
		eng.Logger().Info("WASM RUN LOG:call TC_requireWithMsg", "msg", msg)
		return 0, ErrExecutionReverted
	}
	return 0, nil
}

type TCRevert struct{}

func (t *TCRevert) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRevert(eng, index, args)
}
func (t *TCRevert) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRevert(eng, index, args)
}

// void TC_revert()
func tcRevert(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 0 {
		return 0, ErrInvalidApiArgs
	}
	// TODO:write log
	eng.Logger().Debug("WASM RUN LOG:call TC_revert")

	return 0, ErrExecutionReverted
}

type TCRevertWithMsg struct{}

func (t *TCRevertWithMsg) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRevertWithMsg(eng, index, args)
}
func (t *TCRevertWithMsg) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRevertWithMsg(eng, index, args)
}

// TODO void TC_revertWithMsg(char *msg)
func tcRevertWithMsg(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidApiArgs
	}
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	a, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrMemoryGet
	}
	msg := string(a)

	// TODO:write log
	eng.Logger().Info("WASM RUN LOG:call TC_revertWithMsg", "msg", msg)
	return 0, ErrExecutionReverted
}

type TCPayable struct{}

func (t *TCPayable) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcPayable(eng, index, args)
}
func (t *TCPayable) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasPayable(eng, index, args)
}

// void TC_Payable(bool condition)
func tcPayable(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidApiArgs
	}
	condition := int(args[0])
	if condition == 0 {
		v := eng.Contract.value
		if v != nil && v.Cmp(big.NewInt(0)) > 0 {
			eng.Logger().Debug("WASM RUN LOG:call TC_Payable, payable false,msg.value > 0")
			return 0, ErrContractNotPayable
		}
	}
	return 0, nil
}

// ---------------------------------------------------------
// go json api (optional)

type TCJSONParse struct{}

func (t *TCJSONParse) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONParse(eng, index, args)
}
func (t *TCJSONParse) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONParse(eng, index, args)
}

// c: void *TC_JsonParse(char *data)
func tcJSONParse(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	data, err := vmem.GetString(args[0])
	if err != nil {
		return 0, err
	}

	obj := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &obj); err != nil {
		eng.logger.Error(" TC_JSONParse", "data", string(data), "err", err)
		return 0, err
	}

	root := len(eng.jsonCache)
	eng.jsonCache = append(eng.jsonCache, obj)
	return uint64(root), nil
}

type TCJSONGetInt struct{}

func (t *TCJSONGetInt) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetInt(eng, index, args)
}
func (t *TCJSONGetInt) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetInt(eng, index, args)
}

// c: int TC_JsonGetInt(void *root, char *key)
func tcJSONGetInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}

	v = bytes.Trim(v, "\"")
	i, err := strconv.ParseInt(string(v), 0, 32)
	if err != nil {
		return 0, err
	}

	return uint64(i), nil
}

type TCJSONGetInt64 struct{}

func (t *TCJSONGetInt64) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetInt64(eng, index, args)
}
func (t *TCJSONGetInt64) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetInt64(eng, index, args)
}

// c: long long TC_JsonGetInt64(void *root, char *key)
func tcJSONGetInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}

	v = bytes.Trim(v, "\"")
	i, err := strconv.ParseInt(string(v), 0, 64)
	if err != nil {
		return 0, err
	}

	return uint64(i), nil
}

type TCJSONGetString struct{}

func (t *TCJSONGetString) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetString(eng, index, args)
}
func (t *TCJSONGetString) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetString(eng, index, args)
}

// c: char * TC_JsonGetString(void *root, char *key)
func tcJSONGetString(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}
	lenV := len(v)
	v = v[1 : lenV-1]

	pointer, err := vmem.SetBytes([]byte(v))
	eng.Logger().Debug("WASM RUN LOG:call tcJSONGetString,str", "v", v)
	if err != nil {
		return 0, err
	}
	return uint64(pointer), nil
}

type TCJSONGetAddress struct{}

func (t *TCJSONGetAddress) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetAddress(eng, index, args)
}
func (t *TCJSONGetAddress) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetAddress(eng, index, args)
}

// c: char * TC_JsonGetAddress(void *root, char *key)
func tcJSONGetAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}
	lenV := len(v)
	v = v[1 : lenV-1]

	if !types.IsHexAddress(string(v)) {
		return 0, fmt.Errorf("key(%s) not address", v)
	}

	pointer, err := vmem.SetBytes(bytes.ToLower([]byte(v)))
	eng.Logger().Debug("WASM RUN LOG:call tcJSONGetAddress,address ", "v", string(v[:]))
	if err != nil {
		return 0, err
	}
	return uint64(pointer), nil
}

type TCJSONGetBigInt struct{}

func (t *TCJSONGetBigInt) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetBigInt(eng, index, args)
}
func (t *TCJSONGetBigInt) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetBigInt(eng, index, args)
}

// c: char * TC_JsonGetBigInt(void *root, char *key)
func tcJSONGetBigInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}
	lenV := len(v)
	v = v[1 : lenV-1]

	aBigInt, e1 := new(big.Int).SetString(string(v), 0)
	if !e1 {
		return 0, fmt.Errorf("key(%s),value(%s) not BigInt type", string(key), string(v))
	}

	pointer, err := vmem.SetBytes([]byte(aBigInt.String()))
	eng.Logger().Debug("WASM RUN LOG:call tcJSONGetBigInt,value", "bigInt", aBigInt.String())
	if err != nil {
		return 0, err
	}
	return uint64(pointer), nil
}

type TCJSONGetFloat struct{}

func (t *TCJSONGetFloat) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetFloat(eng, index, args)
}
func (t *TCJSONGetFloat) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetFloat(eng, index, args)
}

// c: float TC_JsonGetFloat(void *root, char *key)
func tcJSONGetFloat(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}

	v = bytes.Trim(v, "\"")
	f, err := strconv.ParseFloat(string(v), 32)
	if err != nil {
		return 0, err
	}

	return math.Float64bits(f), nil
}

type TCJSONGetDouble struct{}

func (t *TCJSONGetDouble) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetDouble(eng, index, args)
}
func (t *TCJSONGetDouble) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetDouble(eng, index, args)
}

// c: double TC_JsonGetDouble(void *root, char *key)
func tcJSONGetDouble(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]
	if len(v) == 0 {
		return 0, fmt.Errorf("key(%s) not exist", string(key))
	}

	v = bytes.Trim(v, "\"")
	f, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return 0, err
	}

	return math.Float64bits(f), nil
}

type TCJSONGetObject struct{}

func (t *TCJSONGetObject) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONGetObject(eng, index, args)
}
func (t *TCJSONGetObject) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONGetObject(eng, index, args)
}

// c: void* TC_JsonGetObject(void *root, char *key)
func tcJSONGetObject(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	v := obj[string(key)]

	childObj := make(map[string]json.RawMessage)
	if err = json.Unmarshal(v, &childObj); err != nil {
		return 0, err
	}

	childIndex := len(eng.jsonCache)
	eng.jsonCache = append(eng.jsonCache, childObj)
	return uint64(childIndex), nil
}

type TCJSONNewObject struct{}

func (t *TCJSONNewObject) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONNewObject(eng, index, args)
}
func (t *TCJSONNewObject) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONNewObject(eng, index, args)
}

// c: void* TC_JsonNewObject()
func tcJSONNewObject(eng *Engine, index int64, args []uint64) (uint64, error) {
	obj := make(map[string]json.RawMessage)
	i := len(eng.jsonCache)
	eng.jsonCache = append(eng.jsonCache, obj)
	return uint64(i), nil
}

type TCJSONPutInt struct{}

func (t *TCJSONPutInt) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutInt(eng, index, args)
}
func (t *TCJSONPutInt) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutInt(eng, index, args)
}

// c: void TC_JsonPutInt(void *root, char *key, int value)
func tcJSONPutInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	obj[string(key)], _ = json.Marshal(int(args[2]))
	return 0, nil
}

type TCJSONPutInt64 struct{}

func (t *TCJSONPutInt64) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutInt64(eng, index, args)
}
func (t *TCJSONPutInt64) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutInt64(eng, index, args)
}

// c: void TC_JsonPutInt64(void *root, char *key, long long value)
func tcJSONPutInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	obj[string(key)], err = json.Marshal(int64(args[2]))
	if err != nil {
		eng.logger.Error(" TC_JSONPutInt64", "key", string(key), "val", int64(args[2]), "err", err)
		return 0, err
	}
	return 0, nil
}

type TCJSONPutString struct{}

func (t *TCJSONPutString) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutString(eng, index, args)
}
func (t *TCJSONPutString) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutString(eng, index, args)
}

// c: void TC_JsonPutString(void *root, char *key, char *value)
func tcJSONPutString(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	val, err := vmem.GetString(args[2])
	if err != nil {
		return 0, err
	}

	obj := eng.jsonCache[root]
	obj[string(key)], err = json.Marshal(string(val))
	if err != nil {
		eng.logger.Error(" TC_JSONPutString", "key", string(key), "val", string(val), "err", err)
		return 0, err
	}
	return 0, nil
}

type TCJSONPutAddress struct{}

func (t *TCJSONPutAddress) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutAddress(eng, index, args)
}
func (t *TCJSONPutAddress) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutAddress(eng, index, args)
}

// c: void TC_JsonPutAddress(void *root, char *key, char *value)
func tcJSONPutAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	val, err := vmem.GetString(args[2])
	if err != nil {
		return 0, err
	}
	if !types.IsHexAddress(string(val)) {
		return 0, fmt.Errorf("key(%s),value(%s) not address", string(key), string(val))
	}
	obj := eng.jsonCache[root]
	obj[string(key)], err = json.Marshal(strings.ToLower(string(val)))
	if err != nil {
		eng.logger.Error(" TC_JSONPutAddress", "key", string(key), "val", string(val), "err", err)
		return 0, err
	}
	return 0, nil
}

type TCJSONPutBigInt struct{}

func (t *TCJSONPutBigInt) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutBigInt(eng, index, args)
}
func (t *TCJSONPutBigInt) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutBigInt(eng, index, args)
}

// c: void TC_JsonPutAddress(void *root, char *key, char *value)
func tcJSONPutBigInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	val, err := vmem.GetString(args[2])
	if err != nil {
		return 0, err
	}
	_, e1 := new(big.Int).SetString(string(val), 0)
	if !e1 {
		return 0, fmt.Errorf("key(%s),value(%s) not BigInt type", string(key), string(val))
	}
	obj := eng.jsonCache[root]
	obj[string(key)], err = json.Marshal(string(val))
	if err != nil {
		eng.logger.Error(" TC_JSONPutBigInt", "key", string(key), "val", string(val), "err", err)
		return 0, err
	}
	return 0, nil
}

type TCJSONPutFloat struct{}

func (t *TCJSONPutFloat) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutFloat(eng, index, args)
}
func (t *TCJSONPutFloat) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutFloat(eng, index, args)
}

// c: void TC_JsonPutFloat(void *root, char *key, float value)
func tcJSONPutFloat(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	val := math.Float32frombits(uint32(args[2]))

	obj := eng.jsonCache[root]
	obj[string(key)], _ = json.Marshal(val)
	return 0, nil
}

type TCJSONPutDouble struct{}

func (t *TCJSONPutDouble) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutDouble(eng, index, args)
}
func (t *TCJSONPutDouble) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasJSONPutDouble(eng, index, args)
}

// c: void TC_JsonPutDouble(void *root, char *key, double value)
func tcJSONPutDouble(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	val := math.Float64frombits(args[2])

	obj := eng.jsonCache[root]
	obj[string(key)], _ = json.Marshal(val)
	return 0, nil
}

type TCJSONPutObject struct{}

func (t *TCJSONPutObject) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONPutObject(eng, index, args)
}
func (t *TCJSONPutObject) Gas(index int64, ops interface{}, args []uint64) (gas uint64, err error) {
	eng := ops.(*Engine)
	gas, data, err := gasJSONPutObject(eng, index, args)
	if err != nil {
		return gas, err
	}

	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	root := int(args[0])
	key, _ := vmem.GetString(args[1])
	rootObj := eng.jsonCache[root]
	rootObj[string(key)] = json.RawMessage(data)
	return gas, nil
}

// c: void TC_JsonPutObject(void *root, char *key, void *child)
func tcJSONPutObject(eng *Engine, index int64, args []uint64) (uint64, error) {
	// @Note: do nothing
	return 0, nil
}

type TCJSONToString struct{}

func (t *TCJSONToString) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcJSONToString(t, eng, index, args)
}
func (t *TCJSONToString) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gas, data, err := gasJSONToString(eng, index, args)
	if err != nil {
		return gas, err
	}

	app, _ := eng.RunningAppFrame()
	app.result = data
	return gas, nil
}

// c: char *TC_JsonToString(void *root)
func tcJSONToString(t *TCJSONToString, eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()

	data := app.result.([]byte)
	// eng.logger.Debug("TC_JsonToString", "data", string(data))
	pointer, err := vmem.SetBytes([]byte(data))
	if err != nil {
		return 0, err
	}
	return uint64(pointer), nil
}
