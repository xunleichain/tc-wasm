package vm

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
)

type gasFunc func(eng *Engine, index int64, args []uint64) (uint64, error)

// toWordSize returns the ceiled word size required for memory expansion.
func ToWordSize(size uint64) uint64 {
	if size > math.MaxUint64-31 {
		return math.MaxUint64/32 + 1
	}
	return (size + 31) / 32
}

// safeAdd returns the result and whether overflow occurred.
func SafeAdd(x, y uint64) (uint64, bool) {
	return x + y, y > math.MaxUint64-x
}

// safeMul returns multiplication result and whether overflow occurred.
func SafeMul(x, y uint64) (uint64, bool) {
	if x == 0 || y == 0 {
		return 0, false
	}
	return x * y, y > math.MaxUint64/x
}

func gasKeccak256(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas := GasExtStep + HashSetGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), Sha3WordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasSha256(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas := GasExtStep + HashSetGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), Sha256PerWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasRipemd160(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas := GasExtStep + AddrSetGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), Ripemd160PerWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasEcrecover(eng *Engine, index int64, args []uint64) (uint64, error) {
	return EcrecoverGas, nil
}

func GasGetBalance(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasTableEIP158.Balance, nil
}

func GasTransfer(eng *Engine, index int64, args []uint64) (uint64, error) {
	return CallValueTransferGas, nil
}

func GasTransferToken(eng *Engine, index int64, args []uint64) (uint64, error) {
	return CallValueTransferGas, nil
}

func gasGetSelfAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func GasSelfDestruct(eng *Engine, index int64, args []uint64) (uint64, error) {
	return 0, nil
}

func MakeGasLog(n uint64) gasFunc {
	return func(eng *Engine, index int64, args []uint64) (uint64, error) {
		runningFrame, _ := eng.RunningAppFrame()
		vmem := runningFrame.VM.VMemory()
		dataLen, err := vmem.Strlen(args[0])
		if err != nil {
			return 0, err
		}
		gas := GasFastStep
		gas, overflow := SafeAdd(gas, n*LogTopicGas)
		if overflow {
			return 0, ErrGasOverflow
		}
		memorySizeGas, overflow := SafeMul(uint64(dataLen), LogDataGas)
		if overflow {
			return 0, ErrGasOverflow
		}
		if gas, overflow = SafeAdd(gas, memorySizeGas); overflow {
			return 0, ErrGasOverflow
		}
		return gas, nil
	}
}

func gasPrints(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas, overflow := SafeMul(ToWordSize(uint64(dataLen)), PrintWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasPrintsl(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	length := int(args[1])
	if length > dataLen {
		length = dataLen
	}
	gas, overflow := SafeMul(ToWordSize(uint64(length)), PrintWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasMalloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasFree(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasCalloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(args[0]*args[1]), MemWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasRealloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(args[1]), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasStrlen(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasIsHexAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func GasIssue(eng *Engine, index int64, args []uint64) (uint64, error) {
	return IssueGas, nil
}

func GasTokenBalance(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasTableEIP158.Balance, nil
}

func GasTokenAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func gasMemcpy(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[2]), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasMemset(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[2]), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasMemmove(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[2]), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasMemcmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[2]), MemWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasStrcmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	data1Len, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	data2Len, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	dataLen := data1Len
	if dataLen > data2Len {
		dataLen = data2Len
	}
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasStrcpy(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasStrconcat(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	data1Len, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	data2Len, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	dataLen := data1Len + data2Len
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasAtoi(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasAtof64(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasAtof32(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasAtoi64(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasItoa(eng *Engine, index int64, args []uint64) (uint64, error) {
	strLen := len(strconv.Itoa(int(args[0])))
	gas := GasExtStep * 2
	wordGas, overflow := SafeMul(ToWordSize(uint64(strLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasI64toa(eng *Engine, index int64, args []uint64) (uint64, error) {
	i := int64(args[0])
	radix := int(args[1])
	strLen := len(strconv.FormatInt(i, radix))
	gas := GasExtStep * 2
	wordGas, overflow := SafeMul(ToWordSize(uint64(strLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasNotify(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	eventIDLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	dataLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	gas := LogTopicGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(eventIDLen)), Sha3WordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	memorySizeGas, overflow := SafeMul(uint64(dataLen), LogDataGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, memorySizeGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasCheckSign(eng *Engine, index int64, args []uint64) (uint64, error) {
	return EcrecoverGas, nil
}

func GasStorageGet(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasTableEIP158.SLoad, nil
}
func GasStoragePureGet(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasTableEIP158.SLoad, nil
}
func GasContractStorageGet(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasTableEIP158.SLoad, nil
}

func GasStorageSetBytes(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}

	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)+args[2]), SstoreSetGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasStorageSet(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}

	dataLen2, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}

	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)+uint64(dataLen2)), SstoreSetGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasStoragePureSetBytes(eng *Engine, index int64, args []uint64) (uint64, error) {
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[1]+args[3]), SstoreSetGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasStoragePureSetString(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataLen, err := vmem.Strlen(args[2])
	if err != nil {
		return 0, err
	}
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(args[1]+uint64(dataLen)), SstoreSetGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func GasStorageDel(eng *Engine, index int64, args []uint64) (uint64, error) {
	return 0, nil
}

func bigIntOpRetLen(eng *Engine, args []uint64, op bigIntOpType) (int, error) {
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
	default:
		return 0, nil
	}
	return len(aBigInt.String()), nil
}

func gasBigIntAdd(eng *Engine, index int64, args []uint64) (uint64, error) {
	retLen, err := bigIntOpRetLen(eng, args, bigIntOpAdd)
	if err != nil {
		return 0, err
	}
	gas := GasExtStep * 3
	wordGas, overflow := SafeMul(ToWordSize(uint64(retLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasBigIntSub(eng *Engine, index int64, args []uint64) (uint64, error) {
	retLen, err := bigIntOpRetLen(eng, args, bigIntOpSub)
	if err != nil {
		return 0, err
	}
	gas := GasExtStep * 3
	wordGas, overflow := SafeMul(ToWordSize(uint64(retLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasBigIntMul(eng *Engine, index int64, args []uint64) (uint64, error) {
	retLen, err := bigIntOpRetLen(eng, args, bigIntOpMul)
	if err != nil {
		return 0, err
	}
	gas := GasExtStep * 3
	wordGas, overflow := SafeMul(ToWordSize(uint64(retLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasBigIntDiv(eng *Engine, index int64, args []uint64) (uint64, error) {
	retLen, err := bigIntOpRetLen(eng, args, bigIntOpDiv)
	if err != nil {
		return 0, err
	}
	gas := GasExtStep * 3
	wordGas, overflow := SafeMul(ToWordSize(uint64(retLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasBigIntMod(eng *Engine, index int64, args []uint64) (uint64, error) {
	retLen, err := bigIntOpRetLen(eng, args, bigIntOpMod)
	if err != nil {
		return 0, err
	}
	gas := GasExtStep * 3
	wordGas, overflow := SafeMul(ToWordSize(uint64(retLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasBigIntCmp(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep * 3, nil
}

func gasBigIntToInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep * 3, nil
}

func GasBlockHash(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + HashSetGas, nil
}

func GasGetCoinbase(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func GasGetGasLimit(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func GasGetNumber(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func GasGetTimestamp(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasGetMsgData(eng *Engine, index int64, args []uint64) (uint64, error) {
	dataLen := len(eng.Contract.Input)
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasGetMsgGas(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasGetMsgSender(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func gasGetMsgSign(eng *Engine, index int64, args []uint64) (uint64, error) {
	input := eng.Contract.Input
	arr := bytes.Split(input, []byte("|"))
	actionLen := len(arr[0])
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(actionLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasAssert(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasExit(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasAbort(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasRequire(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasGasLeft(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func GasNow(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func GasGetTxGasPrice(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func GasGetTxOrigin(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func gasRequireWithMsg(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dataLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), PrintWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasRevert(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasRevertWithMsg(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas := GasQuickStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), PrintWordGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasPayable(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasQuickStep, nil
}

func gasJSONParse(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	dataLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	gas := JsonGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONGetInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasJSONGetInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasJSONGetString(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	obj := eng.jsonCache[root]
	v := obj[string(key)]
	dataLen := len(v)
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONGetAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep + AddrSetGas, nil
}

func gasJSONGetBigInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	obj := eng.jsonCache[root]
	v := obj[string(key)]
	dataLen := len(v)
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONGetFloat(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasJSONGetDouble(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasJSONGetObject(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	root := int(args[0])
	key, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}
	obj := eng.jsonCache[root]
	v := obj[string(key)]
	dataLen := len(v)
	gas := JsonGas
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONNewObject(eng *Engine, index int64, args []uint64) (uint64, error) {
	return GasExtStep, nil
}

func gasJSONPutInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutInt64(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutString(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	valLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen + valLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	valLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen + valLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutBigInt(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	valLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen + valLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutFloat(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutDouble(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[0])
	if err != nil {
		return 0, err
	}
	dataLen := keyLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasJSONPutObject(eng *Engine, index int64, args []uint64) (uint64, []byte, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	keyLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, nil, err
	}
	child := int(args[2])
	childObj := eng.jsonCache[child]
	childJSON, err := json.Marshal(childObj)
	if err != nil {
		return 0, nil, err
	}
	childLen := len(childJSON)

	dataLen := keyLen + childLen
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, nil, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, nil, ErrGasOverflow
	}
	return gas, childJSON, nil
}

func gasJSONToString(eng *Engine, index int64, args []uint64) (uint64, []byte, error) {
	root := int(args[0])
	obj := eng.jsonCache[root]
	data, err := json.Marshal(obj)
	if err != nil {
		return 0, nil, err
	}
	dataLen := len(data)
	gas := GasExtStep
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), MemoryGas)
	if overflow {
		return 0, nil, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, nil, ErrGasOverflow
	}
	return gas, data, nil
}

func gasCallContract(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	actionLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	paramLen := 0
	if len(args) == 3 {
		paramLen, err = vmem.Strlen(args[2])
		if err != nil {
			return 0, err
		}
	}
	dataLen := actionLen + paramLen
	gas := GasTableEIP158.Calls + GasExtStep*2
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}

func gasDelegateCallContract(eng *Engine, index int64, args []uint64) (uint64, error) {
	app, _ := eng.RunningAppFrame()
	vmem := app.VM.VMemory()
	actionLen, err := vmem.Strlen(args[1])
	if err != nil {
		return 0, err
	}
	paramLen := 0
	if len(args) == 3 {
		paramLen, err = vmem.Strlen(args[2])
		if err != nil {
			return 0, err
		}
	}
	dataLen := actionLen + paramLen
	gas := GasTableEIP158.Calls + GasExtStep*2
	wordGas, overflow := SafeMul(ToWordSize(uint64(dataLen)), CopyGas)
	if overflow {
		return 0, ErrGasOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, ErrGasOverflow
	}
	return gas, nil
}
