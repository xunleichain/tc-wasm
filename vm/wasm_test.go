package vm

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseInitArgs(t *testing.T) {
	var data []byte
	data = append(data, wasmBytes...)
	data = append(data, []byte("XLTC")...)

	args := "{\"num\": 100, \"name\":\"xxxx\"}"
	argsLen := uint16(len(args))
	argsBuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(argsBuf, binary.BigEndian, argsLen); err != nil {
		t.Fatalf("binary.Write fail: %s", err)
	}

	data = append(data, argsBuf.Bytes()...)
	data = append(data, []byte(args)...)

	code := "HelloWorld"
	data = append(data, []byte(code)...)

	tmpInput, tmpCode, err := parseInitArgsAndCode(data)
	if err != nil {
		t.Fatalf("parseInitArgsAndCode fail: %s", err)
	}

	t.Logf("input: %s", string(tmpInput))
	t.Logf("code: %s", string(tmpCode))
	if !bytes.HasPrefix(tmpInput, []byte("Init|")) {
		t.Fatalf("input with prefix(Init|): %s", string(tmpInput))
	}
	if !bytes.Equal(tmpInput[5:], []byte(args)) {
		t.Fatalf("input not match: wanted(%s), got(%s)", args, string(tmpInput[5:]))
	}
	if !bytes.Equal([]byte(code), tmpCode) {
		t.Fatalf("code not match: wanted(%s), got(%s)", string(code), string(tmpCode))
	}
}
