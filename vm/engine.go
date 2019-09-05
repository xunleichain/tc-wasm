package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/xunleichain/tc-wasm/mock/log"
	"github.com/xunleichain/tc-wasm/mock/types"
)

const (
	maxFrames = 64
)

var (
	AppCache *sync.Map
)

func init() {
	AppCache = new(sync.Map)
}

type StateDB interface {
	GetContractCode([]byte) []byte
}

type Engine struct {
	logger       log.Logger
	isTrace      bool
	isZeroAddr   bool
	State        StateDB
	AppCache     *sync.Map
	Env          *EnvTable
	AppFrames    []*APP
	FrameIndex   int
	runningFrame *APP
	gas          uint64
	gasUsed      uint64
	Contract     *Contract
	Ctx          interface{}
	fee          uint64

	jsonCache []map[string]json.RawMessage
}

func NewEngine(c *Contract, gas uint64, db StateDB, logger log.Logger) *Engine {
	eng := &Engine{
		logger:     logger,
		State:      db,
		AppCache:   AppCache,
		Env:        NewEnvTable(),
		AppFrames:  make([]*APP, maxFrames),
		FrameIndex: -1,
		gas:        gas,
		Contract:   c,
		jsonCache:  make([]map[string]json.RawMessage, 0, 64),
	}

	return eng
}

func (eng *Engine) EnvTable() *EnvTable {
	return eng.Env
}

func (eng *Engine) Logger() log.Logger {
	return eng.logger
}

func (eng *Engine) AppByName(name string) *APP {
	app, _ := eng.AppCache.Load(name)
	if app != nil {
		return app.(*APP)
	}
	return nil
}

// UseGas implement Backend
func (eng *Engine) UseGas(cost uint64) bool {
	if eng.gas < cost {
		return false
	}
	eng.gas -= cost
	eng.gasUsed += cost
	return true
}

func (eng *Engine) Gas() uint64 {
	return eng.gas
}

func (eng *Engine) GasUsed() uint64 {
	return eng.gasUsed
}

// Caller implement Backend
func (eng *Engine) Caller() []byte {
	caller := eng.Contract.CallerAddress
	return caller.Bytes()
}

func (eng *Engine) SetTrace(isTrace bool) {
	eng.isTrace = isTrace
}

// IsTracing implement Backend
func (eng *Engine) IsTracing() bool {
	return eng.isTrace
}

// Trace implement Backend
func (eng *Engine) Trace(msg string, v ...interface{}) {
	if eng.isTrace {
		eng.logger.Info(msg, v...)
	}
}

func (eng *Engine) NewApp(name string, code []byte, debug bool) (*APP, error) {

	if app := eng.AppByName(name); app != nil {
		return app.Clone(eng), nil
	}

	if len(code) == 0 {
		// addr, err := strToAddress(name)
		// if err != nil {
		// 	return nil, err
		// }
		// types.HexToAddress(app.Name)

		code = eng.State.GetContractCode(types.HexToAddress(name).Bytes())
		if len(code) == 0 {
			return nil, ErrContractNoCode
		}
	}

	app, err := NewApp(name, code, eng, debug, eng.logger)

	if err != nil {
		return nil, err
	}

	eng.AppCache.Store(name, app)

	return app.Clone(eng), nil
}

func (eng *Engine) PushAppFrame(app *APP) (int, error) {
	if eng.FrameIndex >= (maxFrames - 1) {
		return 0, ErrOverFrame
	}

	eng.FrameIndex++
	eng.AppFrames[eng.FrameIndex] = app
	return eng.FrameIndex, nil
}

func (eng *Engine) PopAppFrame() (*APP, int) {
	if eng.FrameIndex < 0 {
		return nil, eng.FrameIndex
	}

	app := eng.AppFrames[eng.FrameIndex]
	eng.AppFrames[eng.FrameIndex] = nil
	eng.FrameIndex--
	return app, eng.FrameIndex
}

func (eng *Engine) RunningAppFrame() (*APP, int) {
	return eng.runningFrame, eng.FrameIndex
}

func (eng *Engine) Run(app *APP, input []byte) (uint64, error) {
	action, args, err := ParseInput(input)
	if err != nil {
		return 0, err
	}
	return eng.run(app, action, args)
}

func (eng *Engine) run(app *APP, action, args string) (ret uint64, err error) {
	if string(action) == "Init" || string(action) == "init" {
		if !eng.Contract.CreateCall {
			return 0, ErrInitEngine
		}
	}

	defer func() {
		if r := recover(); r != nil {
			eng.logger.Debug("[Engine] run recover", "frame_index", eng.FrameIndex, "running_app", eng.runningFrame.Name)
			switch e := r.(type) {
			case error:
				err = e
			default:
				err = fmt.Errorf("exec: %v", e)
			}
		}
	}()

	if eng.runningFrame != nil {
		if eng.runningFrame.Name == app.Name {
			panic("Do not call itself")
		}
		if _, err := eng.PushAppFrame(eng.runningFrame); err != nil {
			return 0, err
		}
	}

	eng.logger.Debug("[Engine] Run begin", "frame_index", eng.FrameIndex, "app", app.Name)
	eng.runningFrame = app
	ret, err = app.Run(action, args)
	eng.runningFrame, _ = eng.PopAppFrame()
	eng.logger.Debug("[Engine] Run end", "frame_index", eng.FrameIndex, "app", app.Name, "ret", ret, "err", err)

	return ret, err
}

// ---------------------------------------------
type TCCallContract struct{}

func (t *TCCallContract) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcCallContract(eng, index, args)
}
func (t *TCCallContract) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasCallContract(eng, index, args)
}

// char * TC_CallContract(char *app, char *action. char *arg)
func tcCallContract(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) < 2 {
		return 0, ErrAppInput
	}

	runningFrame, _ := eng.RunningAppFrame()
	if runningFrame == nil {
		return 0, ErrEmptyFrame
	}

	vmem := runningFrame.VM.VMemory()
	appName, err := vmem.GetString(args[0])
	if err != nil {
		return 0, err
	}
	action, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	var params []byte
	if len(args) == 3 {
		params, err = vmem.GetString(args[2])
		if err != nil {
			return 0, err
		}
	}

	toFrame, err := eng.NewApp(string(appName), nil, false)
	if err != nil {
		return 0, err
	}
	preContract := eng.Contract
	//callContract not support transfer
	eng.Contract = NewContractInner(preContract, AccountRef(types.HexToAddress(string(appName))), big.NewInt(0), eng.Gas())
	eng.Contract.Input = make([]byte, len(action)+len(params)+1)
	copy(eng.Contract.Input[0:], action)
	copy(eng.Contract.Input[len(action):], []byte{'|'})
	copy(eng.Contract.Input[1+len(action):], params)
	eng.logger.Debug("[Engine] TC_CallContract", "app", string(appName), "action", string(action), "params", string(params))
	retPointer, err := eng.run(toFrame, string(action), string(params))
	if err != nil {
		return 0, err
	}
	eng.Contract = preContract

	if retPointer != 0 {
		ret, err := toFrame.VM.VMemory().GetString(uint64(retPointer))
		if err != nil {
			return 0, err
		}

		_ret, err := vmem.SetBytes(ret)
		if err != nil {
			return 0, err
		}
		retPointer = uint64(_ret)
	}

	return retPointer, nil
}

type TCDelegateCallContract struct{}

func (t *TCDelegateCallContract) Call(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return tcDelegateCallContract(eng, index, args)
}
func (t *TCDelegateCallContract) Gas(index int64, ops interface{}, args []uint64) (uint64, error) {
	eng := ops.(*Engine)
	return gasDelegateCallContract(eng, index, args)
}

// char * TC_DelegateCallContract(char *app, char *action. char *arg)
func tcDelegateCallContract(eng *Engine, index int64, args []uint64) (uint64, error) {
	if len(args) < 2 {
		return 0, ErrAppInput
	}

	runningFrame, _ := eng.RunningAppFrame()
	if runningFrame == nil {
		return 0, ErrEmptyFrame
	}

	vmem := runningFrame.VM.VMemory()
	appName, err := vmem.GetString(args[0])
	if err != nil {
		return 0, err
	}
	action, err := vmem.GetString(args[1])
	if err != nil {
		return 0, err
	}

	var params []byte
	if len(args) == 3 {
		params, err = vmem.GetString(args[2])
		if err != nil {
			return 0, err
		}
	}

	toFrame, err := eng.NewApp(string(appName), nil, false)
	if err != nil {
		return 0, err
	}
	preContract := eng.Contract
	eng.Contract = NewContractInner(preContract, AccountRef(preContract.Address()), nil, eng.Gas()).AsDelegate()
	eng.Contract.Input = make([]byte, len(action)+len(params)+1)
	copy(eng.Contract.Input[0:], action)
	copy(eng.Contract.Input[len(action):], []byte{'|'})
	copy(eng.Contract.Input[1+len(action):], params)
	eng.logger.Debug("[Engine] TC_DelegateCallContract", "app", string(appName), "action", string(action), "params", string(params))
	retPointer, err := eng.run(toFrame, string(action), string(params))
	if err != nil {
		return 0, err
	}
	eng.Contract = preContract

	if retPointer != 0 {
		ret, err := toFrame.VM.VMemory().GetString(uint64(retPointer))
		if err != nil {
			return 0, err
		}

		_ret, err := vmem.SetBytes(ret)
		if err != nil {
			return 0, err
		}
		retPointer = uint64(_ret)
	}

	return retPointer, nil
}

func (e *Engine) AddFee(fee uint64) {
	e.fee += fee
}

func (e *Engine) GetFee() uint64 {
	return e.fee
}

func (e *Engine) SetFee(fee uint64) {
	e.fee = fee
}

func (e *Engine) CalFee(realCost uint64, currentFee uint64) {
	e.fee -= currentFee
	if e.gas-realCost > 0 {
		e.fee += e.gas - realCost
	}
}
