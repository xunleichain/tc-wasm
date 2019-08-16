package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/go-interpreter/wagon/disasm"
	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/xunleichain/tc-wasm/mock/log"
)

const (
	/*
		 For c/c++:
			 char * thunderchain_main(char *action, char *arg);

	*/
	APPEntry = "thunderchain_main"
)

var (
	ErrNoAppEntry           = fmt.Errorf("no AppEntry(%s)", APPEntry)
	ErrAppInput             = fmt.Errorf("invalid app input")
	WasmBytes               = []byte{0x00, 0x61, 0x73, 0x6d}
	wasmID           uint32 = 0x6d736100
	wasmIDLength            = 4
	initArgsIDLength        = 4
	initArgsID              = []byte("XLTC")
)

type APP struct {
	logger    log.Logger
	Name      string
	Debug     bool
	Module    *wasm.Module
	VM        *exec.VM
	VmProcess *exec.Process
	Eng       *Engine
	IsPreRun  bool
	EntryFunc string

	result interface{}
}

// Clone just copy
func (app *APP) Clone(eng *Engine) *APP {
	vm := app.VM.Clone(eng)
	vm.RecoverPanic = true
	return &APP{
		logger:    app.logger,
		Name:      app.Name,
		Module:    app.Module,
		Eng:       eng,
		VM:        vm,
		VmProcess: exec.NewProcess(vm),
		//VM:     app.VM.Clone(eng),
		EntryFunc: app.EntryFunc,
	}
}

// NewApp new wasm app module
func NewApp(name string, code []byte, eng *Engine, debug bool, logger log.Logger) (*APP, error) {
	if debug {
		disasm.SetLogger(logger)
		wasm.SetLogger(logger)
		validate.SetLogger(logger)
	}

	reader := bytes.NewReader(code)
	m, err := wasm.ReadModule(reader, eng.EnvTable().resolveImport)
	if err != nil {
		return nil, fmt.Errorf("wasm.ReadMoudle fail: %s", err)
	}

	err = validate.VerifyModule(m)
	if err != nil {
		return nil, fmt.Errorf("validate.VerifyMoudle fail: %s", err)
	}

	app := &APP{
		logger:    logger,
		Name:      name,
		Debug:     debug,
		Module:    m,
		Eng:       eng,
		EntryFunc: APPEntry,
	}

	vm, err := exec.NewVM(m, eng)
	if err != nil {
		return nil, fmt.Errorf("exec.NewVM fail: %s", err)
	}
	app.VM = vm

	return app, nil
}

// GetStartFunction Get Function Index of Start function.
func (app *APP) GetStartFunction() int64 {
	if app.Module.Start != nil {
		return int64(app.Module.Start.Index)
	}
	return -1
}

// GetExportFunction Get Function Index of specific fnName
func (app *APP) GetExportFunction(fnName string) int64 {
	if app.Module.Export != nil {
		entry, ok := app.Module.Export.Entries[fnName]
		if ok && entry.Kind == wasm.ExternalFunction {
			return int64(entry.Index)
		}
	}
	return -1
}

// GetEntryFunction Get Function Index for APPEntry
func (app *APP) GetEntryFunction() int64 {
	return app.GetExportFunction(app.EntryFunc)
}

// IsWasmContract check contract's id
func IsWasmContract(code []byte) bool {
	if len(code) > wasmIDLength {
		if wasmID == (uint32(code[0]) | uint32(code[1])<<8 | uint32(code[2])<<16 | uint32(code[3])<<24) {
			return true
		}
	}
	return false
}

func ParseInput(rinput []byte) (string, string, error) {
	if len(rinput) == 0 {
		return "", "", nil
	}

	arr := bytes.Split(rinput, []byte("|"))
	if len(arr) < 2 {
		return "", "", ErrAppInput
	}
	lenAction := len(arr[0])
	action := string(rinput[:lenAction])
	args := string(rinput[lenAction+1:])

	return action, args, nil
}

func ParseInitArgsAndCode(data []byte) ([]byte, []byte, error) {
	input := []byte("Init|{}")
	code := data
	offset := 0

	if IsWasmContract(data[offset : offset+wasmIDLength+1]) {
		//if bytes.Equal(data[offset:offset+wasmIDLength], wasmID) {
		offset += wasmIDLength
		if bytes.Equal(data[offset:offset+initArgsIDLength], initArgsID) {
			offset += initArgsIDLength

			var argsLen uint16
			if err := binary.Read(bytes.NewReader(data[offset:offset+2]), binary.BigEndian, &argsLen); err != nil {
				return input, code, err
			}
			offset += 2

			if argsLen > 0 {
				_init := []byte("Init|")
				input = make([]byte, len(_init)+int(argsLen))
				copy(input, _init)
				copy(input[len(_init):], data[offset:offset+int(argsLen)])
				offset += int(argsLen)
			}
			code = data[offset:]
		}
	}
	return input, code, nil
}

// Run execute AppEntry Function
// the input format should be "action | args"
func (app *APP) Run(action, args string) (uint64, error) {
	if app.IsPreRun {
		return app.RunF(-1) // already set
	}

	fnIndex := app.GetEntryFunction()
	if fnIndex < 0 {
		return 0, ErrNoAppEntry
	}

	if action == "" && args == "" {
		return app.RunF(fnIndex)
	} else {
		vmem := app.VM.VMemory()
		actionPointer, err := vmem.SetBytes([]byte(action))
		if err != nil {
			return 0, err
		}
		argsPointer, err := vmem.SetBytes([]byte(args))
		if err != nil {
			return 0, err
		}
		params := []uint64{uint64(actionPointer), uint64(argsPointer)}
		return app.RunF(fnIndex, params...)
	}
}

// RunF execute function code based on specific fnInex.
func (app *APP) RunF(fnIndex int64, args ...uint64) (uint64, error) {
	if !app.IsPreRun {
		if err := app.VM.PreRun(fnIndex, args...); err != nil {
			return 0, err
		}
		app.IsPreRun = true
	}

	ret, err := app.VM.Run()
	if err != nil {
		return 0, err
	}

	v := uint64(0)
	switch ret.(type) {
	case uint32:
		v = uint64(ret.(uint32))
	case uint64:
		v = ret.(uint64)
	}

	return v, nil
}
