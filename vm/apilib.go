package vm

import (
	"crypto/sha256"
	"fmt"
	"math/big"

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

type TCEcrecover struct{}

func (t *TCEcrecover) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcEcrecover(eng, index, args)
}
func (t *TCEcrecover) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasEcrecover(eng, index, args)
}

//char *TC_Ecrecover(char* hash, char* v, char* r, char* s)
func tcEcrecover(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	hashTmp, herr := vmem.GetString(args[0])
	vTmp, verr := vmem.GetString(args[1])
	rTmp, rerr := vmem.GetString(args[2])
	sTmp, serr := vmem.GetString(args[3])
	if herr != nil || verr != nil || rerr != nil || serr != nil {
		return 0, ErrInvalidApiArgs
	}
	hash := types.HexToHash(string(hashTmp))
	v, vok := new(big.Int).SetString(string(vTmp), 0)
	r, rok := new(big.Int).SetString(string(rTmp), 0)
	s, sok := new(big.Int).SetString(string(sTmp), 0)
	if !vok || !rok || !sok {
		return 0, ErrInvalidApiArgs
	}
	sign := make([]byte, 65)
	copy(sign[:32], r.Bytes())
	copy(sign[32:64], s.Bytes())
	chainIdMul := new(big.Int).SetInt64(THUNDERCHAINID * 2)
	sign[64] = byte(new(big.Int).Sub(v, chainIdMul).Uint64() - 35)
	// tighter sig s values input homestead only apply to tx sigs
	if !types.ValidateSignatureValues(sign[64], r, s, false) {
		return 0, ErrInvalidApiArgs
	}
	// v needs to be at the end for libsecp256k1
	pubKey, err := types.Ecrecover(hash.Bytes(), sign)
	// make sure the public key is a valid one
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	ret := fmt.Sprintf("0x%x", types.Keccak256(pubKey[1:])[12:])
	eng.Logger().Debug("tcEcrecover", "ret", ret)
	return vmem.SetBytes([]byte(ret))
}

type TCGetBalance struct{}

func (t *TCGetBalance) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcGetBalance(eng, index, args)
}
func (t *TCGetBalance) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasGetBalance(eng, index, args)
}

//char* TC_GetBalance(char *address)
func tcGetBalance(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	addrTmp, err := vmem.GetString(args[0])
	if err != nil || !types.IsHexAddress(string(addrTmp)) {
		return 0, ErrInvalidApiArgs
	}
	addr := types.HexToAddress(string(addrTmp))
	balance := eng.stateDB.GetBalance(addr)
	eng.Logger().Debug("tcGetBalance", "balance", balance.String())
	return vmem.SetBytes([]byte(balance.String()))
}

type TCTransfer struct{}

func (t *TCTransfer) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcTransfer(eng, index, args)
}
func (t *TCTransfer) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasTransfer(eng, index, args)
}

//void TC_Transfer(char *address, char* amount)
func tcTransfer(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	from := eng.contract.self.Address()
	toTmp, err := vmem.GetString(args[0])
	if err != nil || !types.IsHexAddress(string(toTmp)) {
		return 0, ErrInvalidApiArgs
	}
	to := types.HexToAddress(string(toTmp))
	valTmp, err := vmem.GetString(args[1])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	val, ok := big.NewInt(0).SetString(string(valTmp), 0)
	if !ok || val.Sign() < 0 {
		return 0, ErrInvalidApiArgs
	}

	eng.Logger().Debug("tcTransfer", "from", from.String(), "to", to.String(), "val", val)
	if val.Sign() == 0 {
		return 0, nil
	}
	if eng.stateDB.GetBalance(from).Cmp(val) < 0 {
		return 0, ErrBalanceNotEnough
	}
	eng.stateDB.SubBalance(from, val)
	eng.stateDB.AddBalance(to, val)

	return 0, nil
}

type TCTransferToken struct{}

func (t *TCTransferToken) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcTransferToken(eng, index, args)
}
func (t *TCTransferToken) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasTransferToken(eng, index, args)
}

//void TC_TransferToken(char *address, char* tokenAddress, char* amount)
func tcTransferToken(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	from := eng.contract.self.Address()
	toTmp, err := vmem.GetString(args[0])
	if err != nil || !types.IsHexAddress(string(toTmp)) {
		return 0, ErrInvalidApiArgs
	}
	to := types.HexToAddress(string(toTmp))
	tokenTmp, err := vmem.GetString(args[1])
	if err != nil || !types.IsHexAddress(string(tokenTmp)) {
		return 0, ErrInvalidApiArgs
	}
	token := types.HexToAddress(string(tokenTmp))
	valTmp, err := vmem.GetString(args[2])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	val, ok := big.NewInt(0).SetString(string(valTmp), 0)
	if !ok || val.Sign() < 0 {
		return 0, ErrInvalidApiArgs
	}

	eng.Logger().Debug("tcTransferToken", "from", from.String(), "to", to.String(), "token", token.String(), "val", val)
	if val.Sign() == 0 {
		return 0, nil
	}

	if token == types.EmptyAddress {
		if eng.stateDB.GetBalance(from).Cmp(val) >= 0 {
			eng.stateDB.SubBalance(from, val)
			eng.stateDB.AddBalance(to, val)
		} else {
			eng.Logger().Info("insufficient BaseToken balance")
			return 0, ErrBalanceNotEnough
		}
	} else {
		if eng.stateDB.GetTokenBalance(from, token).Cmp(val) >= 0 {
			eng.stateDB.SubTokenBalance(from, token, val)
			eng.stateDB.AddTokenBalance(to, token, val)
		} else {
			eng.Logger().Info("insufficient token balance")
			return 0, ErrBalanceNotEnough
		}
	}

	return 0, nil
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
	eng.Logger().Debug("tcGetSelfAddress", "t", eng.contract.self.Address().String())
	return vmem.SetBytes([]byte(eng.contract.self.Address().String()))
}

type TCSelfDestruct struct{}

func (t *TCSelfDestruct) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcSelfDestruct(eng, index, args)
}
func (t *TCSelfDestruct) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasSelfDestruct(eng, index, args)
}

//char *TC_SelfDestruct(char* recipient)
func tcSelfDestruct(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	addr := eng.contract.self.Address()
	toTmp, err := vmem.GetString(args[0])
	if err != nil || !types.IsHexAddress(string(toTmp)) {
		return 0, ErrInvalidApiArgs
	}
	to := types.HexToAddress(string(toTmp))
	tv := eng.stateDB.GetTokenBalances(addr)

	for i := 0; i < len(tv); i++ {
		eng.stateDB.AddTokenBalance(to, tv[i].TokenAddr, tv[i].Value)
	}

	//suicideToken(eng, addr, to)
	eng.stateDB.Suicide(addr)
	//delete cache
	eng.appCache.Delete(addr.String())

	return 0, nil
}

type TCLog0 struct{}

func (t *TCLog0) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcLog0(eng, index, args)
}
func (t *TCLog0) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gasFunc := makeGasLog(0)
	return gasFunc(eng, index, args)
}

//void TC_Log0(char* data)
func tcLog0(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)

	topics := make([]types.Hash, 0)
	eng.stateDB.AddLog(&types.Log{
		Address:     eng.contract.self.Address(),
		Topics:      topics,
		Data:        data,
		BlockNumber: eng.Ctx.BlockNumber.Uint64(),
		BlockTime:   eng.Ctx.Time.Uint64(),
	})
	return 0, nil
}

type TCLog1 struct{}

func (t *TCLog1) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcLog1(eng, index, args)
}
func (t *TCLog1) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gasFunc := makeGasLog(1)
	return gasFunc(eng, index, args)
}

//void TC_Log1(char* data, char* topic)
func tcLog1(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)

	topics := make([]types.Hash, 0)
	for i := 1; i < 2; i++ {
		topicTmp, err := vmem.GetString(args[i])
		if err != nil {
			return 0, ErrInvalidApiArgs
		}
		topic := types.BytesToHash(topicTmp)
		topics = append(topics, topic)
	}
	eng.stateDB.AddLog(&types.Log{
		Address:     eng.contract.self.Address(),
		Topics:      topics,
		Data:        data,
		BlockNumber: eng.Ctx.BlockNumber.Uint64(),
		BlockTime:   eng.Ctx.Time.Uint64(),
	})
	return 0, nil
}

type TCLog2 struct{}

func (t *TCLog2) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcLog2(eng, index, args)
}
func (t *TCLog2) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gasFunc := makeGasLog(2)
	return gasFunc(eng, index, args)
}

//void TC_Log2(char* data, char* topic1, char* topic2)
func tcLog2(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)

	topics := make([]types.Hash, 0)
	for i := 1; i < 3; i++ {
		topicTmp, err := vmem.GetString(args[i])
		if err != nil {
			return 0, ErrInvalidApiArgs
		}
		topic := types.BytesToHash(topicTmp)
		topics = append(topics, topic)
	}
	eng.stateDB.AddLog(&types.Log{
		Address:     eng.contract.self.Address(),
		Topics:      topics,
		Data:        data,
		BlockNumber: eng.Ctx.BlockNumber.Uint64(),
		BlockTime:   eng.Ctx.Time.Uint64(),
	})
	return 0, nil
}

type TCLog3 struct{}

func (t *TCLog3) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcLog3(eng, index, args)
}
func (t *TCLog3) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gasFunc := makeGasLog(3)
	return gasFunc(eng, index, args)
}

//void TC_Log3(char* data, char* topic1, char* topic2, char* topic3)
func tcLog3(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)

	topics := make([]types.Hash, 0)
	for i := 1; i < 4; i++ {
		topicTmp, err := vmem.GetString(args[i])
		if err != nil {
			return 0, ErrInvalidApiArgs
		}
		topic := types.BytesToHash(topicTmp)
		topics = append(topics, topic)
	}
	eng.stateDB.AddLog(&types.Log{
		Address:     eng.contract.self.Address(),
		Topics:      topics,
		Data:        data,
		BlockNumber: eng.Ctx.BlockNumber.Uint64(),
		BlockTime:   eng.Ctx.Time.Uint64(),
	})
	return 0, nil
}

type TCLog4 struct{}

func (t *TCLog4) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcLog4(eng, index, args)
}
func (t *TCLog4) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	gasFunc := makeGasLog(4)
	return gasFunc(eng, index, args)
}

//void TC_Log4(char* data, char* topic1, char* topic2, char* topic3, char* topic4)
func tcLog4(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	dataTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	data := make([]byte, len(dataTmp))
	copy(data, dataTmp)

	topics := make([]types.Hash, 0)
	for i := 1; i < 5; i++ {
		topicTmp, err := vmem.GetString(args[i])
		if err != nil {
			return 0, ErrInvalidApiArgs
		}
		topic := types.BytesToHash(topicTmp)
		topics = append(topics, topic)
	}
	eng.stateDB.AddLog(&types.Log{
		Address:     eng.contract.self.Address(),
		Topics:      topics,
		Data:        data,
		BlockNumber: eng.Ctx.BlockNumber.Uint64(),
		BlockTime:   eng.Ctx.Time.Uint64(),
	})
	return 0, nil
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

type TCIssue struct{}

func (t *TCIssue) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcIssue(eng, index, args)
}
func (t *TCIssue) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasIssue(eng, index, args)
}

//void TC_Issue(char* amount);
func tcIssue(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	amountTmp, err := vmem.GetString(args[0])
	if err != nil {
		return 0, ErrInvalidApiArgs
	}
	amount, ok := new(big.Int).SetString(string(amountTmp), 0)
	if !ok {
		return 0, ErrInvalidApiArgs
	}

	if amount.Sign() > 0 {
		contractAddr := eng.contract.self.Address()
		eng.stateDB.AddTokenBalance(contractAddr, *eng.contract.CodeAddr, amount)
	}

	return 0, nil
}

type TCTokenBalance struct{}

func (t *TCTokenBalance) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcTokenBalance(eng, index, args)
}
func (t *TCTokenBalance) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasTokenBalance(eng, index, args)
}

//char* TC_TokenBalance(char* addr, char* token);
func tcTokenBalance(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	addrTmp, err := vmem.GetString(args[0])
	if err != nil || !types.IsHexAddress(string(addrTmp)) {
		return 0, ErrInvalidApiArgs
	}
	addr := types.HexToAddress(string(addrTmp))
	tokenTmp, err := vmem.GetString(args[1])
	if err != nil || !types.IsHexAddress(string(tokenTmp)) {
		return 0, ErrInvalidApiArgs
	}
	token := types.HexToAddress(string(tokenTmp))

	var balance *big.Int
	if token == types.EmptyAddress {
		balance = eng.stateDB.GetBalance(addr)
	} else {
		balance = eng.stateDB.GetTokenBalance(addr, token)
	}
	return vmem.SetBytes([]byte(balance.String()))
}

type TCTokenAddress struct{}

func (t *TCTokenAddress) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcTokenAddress(eng, index, args)
}
func (t *TCTokenAddress) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasTokenAddress(eng, index, args)
}

//char* TC_TokenAddress();
func tcTokenAddress(eng *Engine, index int64, args []uint64) (uint64, error) {
	runningFrame, _ := eng.RunningAppFrame()
	vmem := runningFrame.VM.VMemory()
	if eng.Ctx.Token == types.EmptyAddress {
		return vmem.SetBytes([]byte(types.Address{}.String()))
	}
	return vmem.SetBytes([]byte(eng.Ctx.Token.String()))
}
