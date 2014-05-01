// Package program contains i64 tests that depend on package call.
package program

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"

	"github.com/crawshaw/asm/call"
	"github.com/crawshaw/asm/i64"
)

var (
	num1 = new(uint64)
	num2 = new(uint64)
	num3 = new(int32)
	num4 = new(uint8)

	num1ptr = uint64(reflect.ValueOf(num1).Pointer())
	num2ptr = uint64(reflect.ValueOf(num2).Pointer())
	num3ptr = uint64(reflect.ValueOf(num3).Pointer())
	num4ptr = uint64(reflect.ValueOf(num4).Pointer())
)

var progtests = []struct {
	program i64.Program
	init1   uint64
	init2   uint64
	init3   int32
	init4   uint8
	want1   uint64
	want2   uint64
	want3   int32
	want4   uint8
}{
	{
		i64.Program{
			{i64.MOVQ, i64.Imm(uint64(num1ptr)), i64.BX.Addr()},
			{i64.MOVQ, i64.BX.Ind(0), i64.BP.Addr()},
			{i64.ADDQ, i64.Imm(uint8(5)), i64.BP.Addr()},
			{i64.MOVQ, i64.BP.Addr(), i64.BX.Ind(0)},
			{Op: i64.RET},
		},
		2, 0, 0, 0,
		7, 0, 0, 0,
	},
	{
		i64.Program{
			{i64.MOVQ, i64.Imm(uint64(num2ptr)), i64.BX.Addr()},
			{i64.MOVQ, i64.BX.Ind(0), i64.BP.Addr()},
			{i64.MOVQ, i64.Imm(uint64(num1ptr)), i64.BX.Addr()},
			{i64.MOVQ, i64.BX.Ind(0), i64.CX.Addr()},
			{i64.IMULQ, i64.CX.Addr(), i64.BP.Addr()},
			{i64.MOVQ, i64.CX.Addr(), i64.BX.Ind(0)},
			{i64.MOVQ, i64.Imm(uint64(num3ptr)), i64.BX.Addr()},
			{i64.MOVQ, i64.CX.Addr(), i64.BX.Ind(0)},
			{i64.MOVQ, i64.Imm(uint64(num4ptr)), i64.BX.Addr()},
			{i64.MOVB, i64.CX.Addr(), i64.BX.Ind(0)},
			{Op: i64.RET},
		},
		5, 3, 0, 0,
		15, 3, 15, 15,
	},
	{
		i64.Program{
			{i64.MOVQ, i64.Imm(uint64(num1ptr)), i64.CX.Addr()},
			{i64.MOVQ, i64.Imm(uint32(7)), i64.BP.Addr()},
			{i64.MOVQ, i64.Imm(uint32(14)), i64.BX.Addr()},
			{Op: i64.LABEL, From: i64.LabelAddr("loop")},
			{i64.ADDQ, i64.Imm(uint32(1)), i64.BP.Addr()},
			{i64.MOVQ, i64.BP.Addr(), i64.CX.Ind(0)},
			{i64.CMPQ, i64.BP.Addr(), i64.BX.Addr()},
			{Op: i64.JNE, To: i64.LabelAddr("loop")},
			{Op: i64.RET},
		},
		7, 0, 0, 0,
		14, 0, 0, 0,
	},
	{
		i64.Program{
			// *num1 = add_one(8)
			{i64.MOVQ, i64.Imm(uint64(8)), i64.BX.Addr()},
			{i64.SUBQ, i64.Imm(uint8(16)), i64.SP.Addr()},
			{i64.MOVQ, i64.BX.Addr(), i64.SP.Ind(0)},
			{Op: i64.CALL, To: i64.LabelAddr("add_one")},
			{i64.MOVQ, i64.SP.Ind(0), i64.BX.Addr()},
			{i64.MOVQ, i64.Imm(uint64(num1ptr)), i64.CX.Addr()},
			{i64.MOVQ, i64.BX.Addr(), i64.CX.Ind(0)},
			{i64.ADDQ, i64.Imm(uint8(16)), i64.SP.Addr()},
			{Op: i64.RET},

			// add_one(x int64) int64 { return x + 1 }
			{Op: i64.LABEL, From: i64.LabelAddr("add_one")},
			{i64.MOVQ, i64.SP.Ind(8), i64.AX.Addr()},
			{i64.ADDQ, i64.Imm(uint8(1)), i64.AX.Addr()},
			{i64.MOVQ, i64.AX.Addr(), i64.SP.Ind(8)},
			{Op: i64.RET},
		},
		0, 0, 0, 0,
		9, 0, 0, 0,
	},
}

func TestProgram(t *testing.T) {
	var programText string
	defer func() {
		if x := recover(); x != nil {
			t.Fatalf("%s\n\n%v", programText, x)
		}
	}()

	for i, test := range progtests {
		// Print now in case we panic.
		buf := new(bytes.Buffer)
		if err := test.program.PrintText(buf); err != nil {
			t.Errorf("%d: print: %v", i, err)
		}
		programText = buf.String()

		// Run the program.
		code, err := test.program.Bytes()
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		fn := unsafe.Pointer(&code[0])
		*num1, *num2, *num3, *num4 = test.init1, test.init2, test.init3, test.init4
		call.Call(fn, nil, 0)
		if *num1 != test.want1 || *num2 != test.want2 || *num3 != test.want3 || *num4 != test.want4 {
			t.Errorf("%d: got %d,%d,%d,%d, want %d,%d,%d,%d\n%s", i, *num1, *num2, *num3, *num4, test.want1, test.want2, test.want3, test.want4, programText)
		}
	}
}
