package vm

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <stdint.h>
#include <dlfcn.h>

typedef struct {
	void *ctx;
	uint64_t gas;
	uint64_t gas_used;
	int32_t pages;
	uint8_t *mem;

	// internal temp member
	void *_ff;
	uint32_t _findex;
} vm_t;

static inline void get_gas(vm_t *vm, uint64_t *gas_used, uint64_t *gas) {
	*gas_used = vm->gas_used;
	*gas = vm->gas;
}

static inline void update_mem(vm_t *vm, int32_t pages, void *mem) {
	vm->pages = pages;
	vm->mem = (uint8_t *)(mem);
}

static inline void copy_args(uint64_t *dst, uint64_t *src, int32_t n) {
	memcpy(dst, src, sizeof(uint64_t) * n);
}

typedef uint32_t (*tc_main_t)(vm_t*, uint32_t, uint32_t);

static inline tc_main_t get_main_func(void *dl) {
	return (uint32_t (*)(vm_t*, uint32_t, uint32_t))dlsym(dl, "thunderchain_main");
}

static inline int32_t has_main_func(void *dl) {
	tc_main_t _main = get_main_func(dl);
	return _main ? 1 : 0;
}

static uint32_t call_main(void *__ptrs) {
	void *_ptrs[6];
	memcpy(&_ptrs[0], __ptrs, 6 * sizeof(void *));

	void *dl = _ptrs[0];
	void *ctx = _ptrs[1];
	uint64_t *data = (uint64_t *)(_ptrs[2]);
	void *mem = _ptrs[3];
	uint64_t *gas_used = (uint64_t *)(_ptrs[4]);
	uint64_t *gas = (uint64_t *)(_ptrs[5]);

	vm_t vm;
	vm.ctx = ctx;
	vm.gas = data[0];
	vm.gas_used = data[1];
	vm.pages = (int32_t)(data[2]);
	vm.mem = (uint8_t *)(mem);

	uint32_t action = (uint32_t)(data[3]);
	uint32_t args = (uint32_t)(data[4]);

	tc_main_t _main = get_main_func(dl);
	if (_main == NULL) {
		return 0;
	}

	// printf("call_main begin: gas:%lu, gas_used:%lu\n", vm.gas, vm.gas_used);
	uint32_t ret = _main(&vm, action, args);

	get_gas(&vm, gas_used, gas);
	return ret;
}

*/
import "C"
import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/memory"
	"github.com/xunleichain/tc-wasm/mock/log"
)

const wasmPageSize = 65536 // (64 KB)

// Native --
type Native struct {
	app    *APP
	logger log.Logger
	dl     *dynamicLib
	t      time.Time
	ret    uint64
}

// NewNative --
func NewNative(app *APP, file string) (*Native, error) {
	cfile := C.CString(file)
	handle, err := C.dlopen(cfile, C.RTLD_LAZY|C.RTLD_LOCAL)
	C.free(unsafe.Pointer(cfile))
	if handle == nil {
		app.logger.Info("[Native] C.dlopen fail", "file", file, "err", err)
		return nil, fmt.Errorf("C.dlopen")
	}

	native := &Native{
		app:    app,
		logger: app.logger,
		t:      time.Now(),
	}

	if ret := C.has_main_func(handle); ret < 0 {
		C.dlclose(handle)
		return nil, fmt.Errorf("%s Without CMain function", file)
	}

	native.dl = newDynamicLib(file, handle, app.logger)
	return native, nil
}

func (native *Native) close() {
	if native != nil {
		native.dl.free()
	}
}

func (native *Native) remove() {
	native.dl.setDelete()
	native.dl.free()
}

func (native *Native) count() uint64 {
	return native.dl.count()
}

func (native *Native) clone(app *APP) *Native {
	if native == nil {
		return nil
	}

	dl := native.dl.get()
	if dl == nil {
		return nil
	}

	t := time.Now()
	native.t = t
	return &Native{
		app:    app,
		logger: app.logger,
		dl:     dl,
		t:      t,
	}
}

// RunCMain --
func (native *Native) RunCMain(action, args string) (ret uint64, err error) {
	eng := native.engine()
	mem := native.memory()

	actionP, err := mem.SetBytes([]byte(action))
	if err != nil {
		return 0, err
	}
	argsP, err := mem.SetBytes([]byte(args))
	if err != nil {
		return 0, err
	}

	var gas uint64
	var gasUsed uint64
	pages := uint64(mem.HeapSize() / wasmPageSize)
	data := []uint64{eng.gas, eng.gasUsed, pages, actionP, argsP}

	ptrs := make([]uintptr, 6)
	ptrs[0] = uintptr(native.dl.so)
	ptrs[1] = uintptr(unsafe.Pointer(native))
	ptrs[2] = uintptr(unsafe.Pointer(&data[0]))
	ptrs[3] = uintptr(unsafe.Pointer(&mem.Memory[0]))
	ptrs[4] = uintptr(unsafe.Pointer(&gasUsed))
	ptrs[5] = uintptr(unsafe.Pointer(&gas))

	defer func() {
		if r := recover(); r != nil {
			eng.logger.Debug("[Native] RunCMain recover", "frame_index", eng.FrameIndex, "running_app", eng.runningFrame.String(), "err", err, "bt", string(debug.Stack()))
			switch e := r.(type) {
			case error:
				err = e
				if err == ErrExecutionExit {
					ret = native.ret
					err = nil
				}
			default:
				err = fmt.Errorf("exec: %v", e)
			}
		}
	}()

	iret := C.call_main(unsafe.Pointer(&ptrs[0]))
	native.updateGas(gas, gasUsed)
	native.app.logger.Debug("[Native] RunCMain done", "app", native.name(), "ret", iret, "gas", gas, "gas_used", gasUsed)
	return uint64(iret), nil
}

// Printf --
func (native *Native) Printf(f string, args ...interface{}) {
	native.logger.Debug(fmt.Sprintf(f, args...))
}

func (native *Native) name() string {
	return native.app.String()
	// return native.app.Name
}

func (native *Native) engine() *Engine {
	return native.app.Eng
}

func (native *Native) memory() *memory.MemManager {
	return native.app.VM.VMemory()
}

func (native *Native) envTable() *EnvTable {
	return native.engine().Env
}

func (native *Native) getFuncByName(name string) EnvFunc {
	return native.envTable().GetFuncByName(name)
}

func (native *Native) updateGas(gas, gasUsed uint64) {
	eng := native.engine()
	eng.gas = gas
	eng.gasUsed = gasUsed
}

func (native *Native) useGas(cost uint64) bool {
	eng := native.engine()
	return eng.UseGas(cost)
}

func updateGas(cvm *C.vm_t, gas, gasUsed uint64) {
	cvm.gas = C.uint64_t(gas)
	cvm.gas_used = C.uint64_t(gasUsed)
}

func updateMem(cvm *C.vm_t, native *Native) {
	mem := native.memory()
	pages := int32(mem.HeapSize() / wasmPageSize)
	if int32(cvm.pages) != pages {
		C.update_mem(cvm, C.int32_t(pages), unsafe.Pointer(&mem.Memory[0]))
	}
}

// -------------------------------------------------------

// GoPanic --
//export GoPanic
func GoPanic(cvm *C.vm_t, cmsg *C.char) {
	native := (*Native)(cvm.ctx)
	msg := C.GoString(cmsg)

	native.updateGas(uint64(cvm.gas), uint64(cvm.gas_used))
	native.Printf("[GoPanic] app:%s, msg:%s", native.name(), msg)

	switch msg {
	case "Abort":
		panic(ErrContractAbort)
	case "OutOfGas":
		panic("[vm] execCode: OutOfGas") // panic(ErrOutOfGas)
	case "Unreachable":
		panic(exec.ErrUnreachable)
	case "ElemIndexOverflow":
		panic(exec.ErrUndefinedElementIndex)
	default:
		panic(msg)
	}
}

// GoRevert --
//export GoRevert
func GoRevert(cvm *C.vm_t, cmsg *C.char) {
	native := (*Native)(cvm.ctx)
	msg := C.GoString(cmsg)

	native.updateGas(uint64(cvm.gas), uint64(cvm.gas_used))
	native.Printf("[GoRevert] app:%s, msg:%s", native.name(), msg)
	panic(ErrExecutionReverted)
}

// GoExit --
//export GoExit
func GoExit(cvm *C.vm_t, cstatus C.int32_t) {
	native := (*Native)(cvm.ctx)
	status := int32(cstatus)

	native.updateGas(uint64(cvm.gas), uint64(cvm.gas_used))
	native.Printf("[GoExit] app:%s, status:%d", native.name(), status)
	native.ret = uint64(status)
	panic(ErrExecutionExit)
}

// GoGrowMemory --
//export GoGrowMemory
func GoGrowMemory(cvm *C.vm_t, pages C.int32_t) {
	native := (*Native)(cvm.ctx)
	mem := native.memory()

	if err := mem.GrowMem(int(pages) * wasmPageSize); err != nil {
		native.updateGas(uint64(cvm.gas), uint64(cvm.gas_used))
		native.Printf("[GoGrowMem] fail: app:%s, pages:%d, err:%s", native.name(), pages, err)
		panic(err)
	}
	C.update_mem(cvm, C.int32_t(pages), unsafe.Pointer(&mem.Memory[0]))
	native.Printf("[GoGrowMemory] ok: app:%s, pages:%d", native.name(), int(pages))
}

// GoFunc --
//export GoFunc
func GoFunc(cvm *C.vm_t, cname *C.char, cArgn C.int32_t, cArgs *C.uint64_t) uint64 {
	native := (*Native)(cvm.ctx)
	eng := native.engine()

	args := make([]uint64, int(cArgn))
	if len(args) > 0 {
		C.copy_args((*C.uint64_t)(unsafe.Pointer(&args[0])), cArgs, cArgn)
	}
	index := int64(-1)
	name := C.GoString(cname)

	native.updateGas(uint64(cvm.gas), uint64(cvm.gas_used))
	envFunc := native.getFuncByName(name)
	if envFunc == nil {
		native.Printf("[GoFunc] Not Exist: app:%s, name:%s", native.name(), name)
		panic(fmt.Sprintf("[GoFunc] Not Exist: app:%s, name:%s", native.name(), name))
	}

	preFee := eng.GetFee()
	cost, err := envFunc.Gas(index, eng, args)
	if err != nil {
		native.Printf("[GoFunc] Gas() fail: app:%s, name:%s, err%s", native.name(), name, err)
		eng.SetFee(preFee)
		panic(fmt.Sprintf("[vm] execCode: calc gas fail: %s", err))
	}

	if !native.useGas(cost) {
		native.Printf("[GoFunc] OutOfGas: app:%s, name:%s, gas_used:%d, gas:%d, cost:%d",
			native.name(), name, eng.gasUsed, eng.gas, cost)
		currentFee := eng.GetFee() - preFee
		realCost := cost - currentFee
		eng.CalFee(realCost, currentFee)
		panic("[vm] execCode: OutOfGas")
	}

	ret, err := envFunc.Call(index, eng, args)
	if err != nil {
		native.Printf("[GoFunc] Call() fail: app:%s, name:%s, err:%s", native.name(), name, err)
		panic(err)
	}

	updateGas(cvm, eng.gas, eng.gasUsed)
	updateMem(cvm, native)
	// native.app.logger.Debug("[GoFunc] Call() ok", "app", native.name(), "name", name, "cost", cost)
	return ret
}

// -----------------------------

type dynamicLib struct {
	sync.Mutex
	isDelete bool
	ref      uint64
	file     string
	so       unsafe.Pointer
	logger   log.Logger
}

func newDynamicLib(file string, so unsafe.Pointer, logger log.Logger) *dynamicLib {
	return &dynamicLib{
		ref:    1,
		file:   file,
		so:     so,
		logger: logger,
	}
}

func (dl *dynamicLib) count() uint64 {
	if dl == nil {
		return 0
	}

	dl.Lock()
	ref := dl.ref
	dl.Unlock()
	return ref
}

func (dl *dynamicLib) get() *dynamicLib {
	if dl == nil {
		return nil
	}

	dl.Lock()
	defer dl.Unlock()

	if dl.isDelete {
		dl.logger.Printf("[dynamicLib] deleting %s", dl.file)
		return nil
	}

	dl.ref++
	return dl
}

func (dl *dynamicLib) free() {
	if dl == nil {
		return
	}

	dl.Lock()
	defer dl.Unlock()

	dl.ref--
	if dl.ref == 0 {
		if dl.so != nil {
			C.dlclose(dl.so)
			dl.so = nil
			dl.logger.Printf("[dynamicLib] dlclose %s", dl.file)
		}
		if dl.isDelete {
			os.Remove(dl.file)
			// dl.isDelete = false
			dl.logger.Printf("[dynamicLib] remove %s", dl.file)
		}
	}

}

func (dl *dynamicLib) setDelete() {
	if dl == nil {
		return
	}

	dl.Lock()
	dl.isDelete = true
	dl.Unlock()
}
