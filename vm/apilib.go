package vm

import (
	"crypto/sha256"
	"fmt"
	"github.com/xunleichain/tc-wasm/mock/types"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// const THUNDERCHAINID = 30261
var THUNDERCHAINID = SignParam.Int64()

type TCKeccak256 struct{}

func (t *TCKeccak256) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcKeccak256(eng, index, args)
}
func (t *TCKeccak256) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasKeccak256(eng, index, args)
}

//char *TC_Keccak256(char* data)
func tcKeccak256(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	data, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	d := sha3.NewLegacyKeccak256()
	d.Write(data)
	hash := d.Sum(nil)
	eng.Logger().Debug("tcKeccak256 0x", "hash", hash)
	return vmem.SetBytes([]byte(fmt.Sprintf("0x%x", hash)))
}

type TCSha256 struct{}

func (t *TCSha256) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcSha256(eng, index, args)
}
func (t *TCSha256) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasSha256(eng, index, args)
}

//char *TC_Sha256(char* data)
func tcSha256(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	data, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	hash := sha256.Sum256(data)
	eng.Logger().Debug("tcSha256 0x", "hash", hash)
	return vmem.SetBytes([]byte(fmt.Sprintf("0x%x", hash)))
}

type TCRipemd160 struct{}

func (t *TCRipemd160) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRipemd160(eng, index, args)
}
func (t *TCRipemd160) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRipemd160(eng, index, args)
}

//char *TC_Ripemd160(char* data)
func tcRipemd160(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	data, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	ripemd := ripemd160.New()
	ripemd.Write(data)
	hash := ripemd.Sum(nil)
	eng.Logger().Debug("tcRipemd160 0x", "hash", hash)
	return vmem.SetBytes([]byte(fmt.Sprintf("0x%x", hash)))
}

type TCGetSelfAddress struct{}

func (t *TCGetSelfAddress) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetSelfAddress(eng, index, args)
}
func (t *TCGetSelfAddress) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetSelfAddress(eng, index, args)
}

//char *TC_GetSelfAddress()
func tcGetSelfAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	eng.Logger().Debug("tcGetSelfAddress", "t", eng.Contract.Self.Address().String())
	return vmem.SetBytes([]byte(eng.Contract.Self.Address().String()))
}

type TCPrints struct{}

func (t *TCPrints) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcPrints(eng, index, args)
}
func (t *TCPrints) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasPrints(eng, index, args)
}

//void TC_Prints(const char * cstr)
func tcPrints(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)
	eng.Logger().Info("WASM RUN LOG: " + string(data))
	return 0, nil
}

type TCPrintsl struct{}

func (t *TCPrintsl) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcPrintsl(eng, index, args)
}
func (t *TCPrintsl) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasPrintsl(eng, index, args)
}

//void TC_Printsl(const char * cstr, uint32_t len)
func tcPrintsl(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	length := int(args[1])
	if length > len(dataTmp) {
		length = len(dataTmp)
	}
	data := make([]byte, length)
	copy(data, dataTmp[:length])
	eng.Logger().Info("WASM RUN LOG: " + string(data))
	return 0, nil
}

type TCMalloc struct{}

func (t *TCMalloc) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcMalloc(eng, index, args)
}
func (t *TCMalloc) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasMalloc(eng, index, args)
}

//void *TC_Malloc(uint size)
func tcMalloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	ptr, err := vmem.Malloc(int(args[0]))
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

type TCFree struct{}

func (t *TCFree) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcFree(eng, index, args)
}
func (t *TCFree) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasFree(eng, index, args)
}

//void TC_Free(void* ptr)
func tcFree(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	vmem.Free(args[0])
	return 0, nil
}

type TCCalloc struct{}

func (t *TCCalloc) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcCalloc(eng, index, args)
}
func (t *TCCalloc) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasCalloc(eng, index, args)
}

//void *TC_Calloc(size_t, size_t)
func tcCalloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	itemCnt := int(args[0])
	itemSize := int(args[1])
	size := itemCnt * itemSize
	ptr, err := vmem.Malloc(size)
	if err != nil {
		return 0, err
	}
	vmem.Memset(ptr, 0, size)
	return ptr, nil
}

type TCRealloc struct{}

func (t *TCRealloc) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcRealloc(eng, index, args)
}
func (t *TCRealloc) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasRealloc(eng, index, args)
}

//void *TC_Realloc(void *, size_t)
func tcRealloc(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	ptr, err := vmem.Realloc(args[0], int(args[1]))
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

type TCStrlen struct{}

func (t *TCStrlen) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcStrlen(eng, index, args)
}
func (t *TCStrlen) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasStrlen(eng, index, args)
}

//int TC_Strlen(const char *str);
func tcStrlen(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	size, err := vmem.Strlen(args[0])
	return uint64(size), err
}

type TCIsHexAddress struct{}

func (t *TCIsHexAddress) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcIsHexAddress(eng, index, args)
}
func (t *TCIsHexAddress) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasIsHexAddress(eng, index, args)
}

//bool TC_IsHexAddress(const char *str)
func tcIsHexAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	ret := uint64(0)
	if types.IsHexAddress(string(dataTmp)) {
		ret = uint64(1)
	}
	return ret, err
}
