package types

import (
	"bytes"
	"encoding/hex"
	"github.com/xunleichain/tc-wasm/mock/deps/hexutil"
	"github.com/xunleichain/tc-wasm/mock/deps/rlp"
	"math/big"
	"testing"
)

var (
	testAddrHex = "970e8128ab834e8eac17ab8e3812f010678cf791"
	testmsg     = hexutil.MustDecode("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig     = hexutil.MustDecode("0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301")
	testpubkey  = hexutil.MustDecode("0x04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")
)

func TestCreateAddress(t *testing.T) {
	addr := HexToAddress(testAddrHex)
	oaddr0 := oldCreateAddress(addr, 0)
	oaddr1 := oldCreateAddress(addr, 1)
	oaddr2 := oldCreateAddress(addr, 2)
	checkAddr(t, HexToAddress("333c3310824b7c685133f2bedb2ca4b8b4df633d"), oaddr0)
	checkAddr(t, HexToAddress("8bda78331c916a08481428e4b07c96d3e916d165"), oaddr1)
	checkAddr(t, HexToAddress("c9ddedf451bc62ce88bf9292afb13df35b670699"), oaddr2)
	caddr0 := CreateAddress(addr, 0, nil)
	caddr1 := CreateAddress(addr, 1, nil)
	caddr2 := CreateAddress(addr, 2, nil)
	checkAddr(t, HexToAddress("c3422dbc55e9a331c6114858ee53ef6e7964ef18"), caddr0)
	checkAddr(t, HexToAddress("91d3aba19c8225cda8e1a2b0cebbb109dd8f12e1"), caddr1)
	checkAddr(t, HexToAddress("a847bed53d0c23ddf8c81db663b35362f6a3bfee"), caddr2)
}

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestKeccak256Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := Keccak256Hash(in); return h[:] }, msg, exp)
}

func TestEcrecover(t *testing.T) {
	pubkey, err := Ecrecover(testmsg, testsig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}
	if !bytes.Equal(pubkey, testpubkey) {
		t.Errorf("pubkey mismatch: want: %x have: %x", testpubkey, pubkey)
	}
}

func TestValidateSignatureValues(t *testing.T) {
	check := func(expected bool, v byte, r, s *big.Int) {
		if ValidateSignatureValues(v, r, s, false) != expected {
			t.Errorf("mismatch for v: %d r: %d s: %d want: %v", v, r, s, expected)
		}
	}
	minusOne := big.NewInt(-1)
	one := Big1
	zero := big.NewInt(0)
	secp256k1nMinus1 := new(big.Int).Sub(secp256k1N, Big1)

	// correct v,r,s
	check(true, 0, one, one)
	check(true, 1, one, one)
	// incorrect v, correct r,s,
	check(false, 2, one, one)
	check(false, 3, one, one)

	// incorrect v, combinations of incorrect/correct r,s at lower limit
	check(false, 2, zero, zero)
	check(false, 2, zero, one)
	check(false, 2, one, zero)
	check(false, 2, one, one)

	// correct v for any combination of incorrect r,s
	check(false, 0, zero, zero)
	check(false, 0, zero, one)
	check(false, 0, one, zero)

	check(false, 1, zero, zero)
	check(false, 1, zero, one)
	check(false, 1, one, zero)

	// correct sig with max r,s
	check(true, 0, secp256k1nMinus1, secp256k1nMinus1)
	// correct v, combinations of incorrect r,s at upper limit
	check(false, 0, secp256k1N, secp256k1nMinus1)
	check(false, 0, secp256k1nMinus1, secp256k1N)
	check(false, 0, secp256k1N, secp256k1N)

	// current callers ensures r,s cannot be negative, but let's test for that too
	// as crypto package could be used stand-alone
	check(false, 0, minusOne, one)
	check(false, 0, one, minusOne)
}

// oldCreateAddress creates an contract address given the bytes and nonce
func oldCreateAddress(b Address, nonce uint64) Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	return BytesToAddress(Keccak256(data)[12:])
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if !bytes.Equal(exp, sum) {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 Address) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %x have: %x", addr0, addr1)
	}
}
