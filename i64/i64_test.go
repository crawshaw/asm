package i64

import (
	"bytes"
	"reflect"

	"testing"
)

var i64tests = []struct {
	ins  Instruction
	text string
	want []byte
}{
	{
		Instruction{ADDQ, BP.Addr(), BX.Addr()},
		"ADDQ  BP,BX",
		[]byte{0x48, 0x01, 0xeb},
	},
	{
		Instruction{Op: RET},
		"RET   ,",
		[]byte{0xc3},
	},
	{
		Instruction{Op: PUSHQ, From: Addr{Type: Imm8, Value: 0}},
		"PUSHQ 0x0,",
		[]byte{0x6a, 0x00},
	},
	{
		Instruction{Op: PUSHQ, From: Imm(uint32(0x9d42))},
		"PUSHQ 0x9d42,",
		[]byte{0x68, 0x42, 0x9d, 0x00, 0x00},
	},
	{
		Instruction{MOVL, Imm(uint32(72)), AX.Addr()},
		"MOVL  0x48,AX",
		[]byte{0xb8, 0x48, 0x00, 0x00, 0x00},
	},
	{
		Instruction{MOVQ, SP.Ind(8), BX.Addr()},
		"MOVQ  8+(SP),BX",
		[]byte{0x48, 0x8b, 0x5c, 0x24, 0x08},
	},
	{
		Instruction{MOVQ, Imm(uint32(1)), SP.Ind(0)},
		"MOVQ  0x1,(SP)",
		[]byte{0x48, 0xc7, 0x04, 0x24, 0x01, 0x00, 0x00, 0x00},
	},
	{
		Instruction{MOVQ, Imm(uint64(0xabcd1234abcd)), BP.Addr()},
		"MOVQ  0xabcd1234abcd,BP",
		[]byte{0x48, 0xbd, 0xcd, 0xab, 0x34, 0x12, 0xcd, 0xab, 0x00, 0x00},
	},
	{
		Instruction{CMPQ, SP.Addr(), CX.Ind(0)},
		"CMPQ  SP,(CX)",
		[]byte{0x48, 0x3b, 0x21},
	},
	{
		Instruction{Op: POPQ, To: AX.Addr()},
		"POPQ  ,AX",
		[]byte{0x58},
	},
	{
		Instruction{Op: PUSHQ, From: BX.Addr()},
		"PUSHQ BX,",
		[]byte{0x53},
	},
	{
		Instruction{Op: JHI, To: Rel(int8(0x0a))},
		"JHI   ,:(a)",
		[]byte{0x77, 0x0a},
	},
	{
		Instruction{Op: JHI, To: Addr{Rel8, 0x0a, 0, "labelname"}},
		"JHI   ,labelname:(a)",
		[]byte{0x77, 0x0a},
	},
	{
		Instruction{Op: CALL, To: Rel(int32(-0x113))},
		"CALL  ,:(-113)",
		[]byte{0xe8, 0xed, 0xfe, 0xff, 0xff},
	},
	{
		Instruction{MOVSD, SP.Ind(8), X0.Addr()},
		"MOVSD 8+(SP),X0",
		[]byte{0xf2, 0x0f, 0x10, 0x44, 0x24, 0x08},
	},
	{
		Instruction{MOVSD, X0.Addr(), SP.Ind(8)},
		"MOVSD X0,8+(SP)",
		[]byte{0xf2, 0x0f, 0x11, 0x44, 0x24, 0x08},
	},
	{
		Instruction{ADDSD, X0.Addr(), X1.Addr()},
		"ADDSD X0,X1",
		[]byte{0xf2, 0x0f, 0x58, 0xc1},
	},
	{
		Instruction{Op: IDIVL, To: BX.Addr()},
		"IDIVL ,BX",
		[]byte{0xf7, 0xfb},
	},
	{
		Instruction{Op: IDIVQ, To: BX.Addr()},
		"IDIVQ ,BX",
		[]byte{0x48, 0xf7, 0xfb},
	},
}

func TestI64(t *testing.T) {
	for _, test := range i64tests {
		c := new(ins)
		if err := c.make(&test.ins); err != nil {
			t.Errorf("%v: %v", test.ins, err)
			continue
		}
		buf := new(bytes.Buffer)
		if _, err := c.writeTo(buf); err != nil {
			t.Errorf("%v: %v", test.ins, err)
			continue
		}
		if got := buf.Bytes(); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%v.Bytes()=%x, want %x", test.ins, got, test.want)
			continue // no point seeing text error too
		}
		buf = new(bytes.Buffer)
		c.printText(buf)
		if got := buf.String(); got != test.text {
			t.Errorf("%v.printText()=%q, want %q", test.ins, got, test.text)
		}
	}
}

func TestOpName(t *testing.T) {
	names := make(map[string]Op)
	for i := LABEL; i < lastOp; i++ {
		name, ok := opName[i]
		if !ok {
			t.Errorf("op %d has no name", i)
			continue
		}
		if other, ok := names[name]; ok {
			t.Errorf("op %d and op %d share name %q", i, other, name)
		}
		names[name] = i
	}
}
