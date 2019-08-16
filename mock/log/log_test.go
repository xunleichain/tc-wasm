package log

import (
	"math/big"
	"testing"
)

type innerS struct {
	bf *big.Float
}

type outerS struct {
	in innerS
	e1 uint32
}

var (
	i0         = 13
	i1  int32  = 520
	i2  int64  = 521
	i3  uint64 = 528
	bi0        = big.NewInt(i2)

	ss = outerS{
		in: innerS{bf: big.NewFloat(3.14)},
		e1: 65535,
	}
)

func TestAll(t *testing.T) {
	Debug("Test Debug", "ss", ss)
	Info("Test Info", "i0", i0, "i1", i1)
	Error("Test Error", "i2", i2, "i3", i3, "bi0", bi0)

	logger := With("Key", "Value")
	logger.Debug("Test KeyValue Debug", "ss", ss)
	logger.Info("Test KeyValue Info", "i0", i0, "i1", i1)
	logger.Error("Test KeyValue Error", "i2", i2, "i3", i3, "bi0", bi0)

	logger = With("KeyError")
	logger.Debug("Test KeyError Debug", "ss", ss)
	logger.Info("Test KeyError Info", "i0", i0, "i1", i1)
	logger.Error("Test KeyError Error", "i2", i2, "i3", i3, "bi0", bi0)
}
