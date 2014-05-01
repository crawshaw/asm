package i64

import (
	"fmt"
	"io"
)

// AddrType is the type of an Addr. An address can be a register or data.
//
// The usual use of AddrType is as an enum. However, when building the optab,
// it is convienient to express op addresses as a bit field. So the AddrType
// values are spaced accordingly.
type AddrType int

const (
	None  AddrType = 0
	Reg   AddrType = 1       // standard register
	Ind   AddrType = 1 << 1  // memory address stored in a register
	Xmm   AddrType = 1 << 2  // SSE register, for FP
	Imm8  AddrType = 1 << 3  // immediate data, 8 bits.
	Imm16 AddrType = 1 << 4  // immediate data, 16 bits.
	Imm32 AddrType = 1 << 5  // immediate data, 32 bits.
	Imm64 AddrType = 1 << 6  // immediate data, 64 bits.
	Rel8  AddrType = 1 << 7  // relative address, signed 8 bits.
	Rel16 AddrType = 1 << 8  // relative address, signed 16 bits.
	Rel32 AddrType = 1 << 9  // relative address, signed 32 bits.
	Label AddrType = 1 << 10 // label, no binary representation.
)

// Addr is an address used by an instruction.
// It may represent a register, or data.
type Addr struct {
	Type  AddrType
	Value interface{} // either an int primitive or a Register
	Disp  uint64      // displacement value
	Name  string      // label, or debugging information
}

// Imm builds an Addr that represents immediate data.
// The value of v must be of type uint8, uint32, or uint64.
func Imm(v interface{}) Addr {
	switch v.(type) {
	case uint8:
		return Addr{Type: Imm8, Value: v}
	case uint32:
		return Addr{Type: Imm32, Value: v}
	case uint64:
		return Addr{Type: Imm64, Value: v}
	}
	panic(fmt.Sprintf("unknown Imm value: %T (%v)", v, v))
}

// Rel builds an Addr that represents a relative address.
// The value of v must be of type int8 or int32.
func Rel(v interface{}) Addr {
	switch v.(type) {
	case int8:
		return Addr{Type: Rel8, Value: v}
	case int32:
		return Addr{Type: Rel32, Value: v}
	}
	panic(fmt.Sprintf("unknown Rel value: %T (%v)", v, v))
}

// LabelAddr builds an Addr that represents a label.
func LabelAddr(name string) Addr { return Addr{Type: Label, Name: name} }

func (a *Addr) valueInt64() int64 {
	switch v := a.Value.(type) {
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return int64(v)
	case int:
		return int64(v)
	}
	panic(fmt.Sprintf("%v: value not signed integer", *a))
}

func (a *Addr) valueUint64() uint64 {
	switch v := a.Value.(type) {
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return uint64(v)
	case int8, int16, int32, int64, int:
		return uint64(a.valueInt64())
	}
	panic(fmt.Sprintf("%v: value not integer", *a)) // TODO error?
}

func (p *Addr) printText(w io.Writer, codeblockEnd int) {
	switch p.Type {
	case None:
	case Reg, Xmm:
		name := p.Value.(Register).String()
		io.WriteString(w, name)
	case Ind:
		name := p.Value.(Register).String()
		if p.Disp != 0 {
			fmt.Fprintf(w, "%x+(%s)", p.Disp, name)
		} else {
			fmt.Fprintf(w, "(%s)", name)
		}
	case Rel8, Rel16, Rel32:
		if codeblockEnd != 0 {
			// We know where we are, print the absolute jump point.
			blk := int64(codeblockEnd) + p.valueInt64()
			fmt.Fprintf(w, "%s:(%06x)", p.Name, blk)
		} else {
			fmt.Fprintf(w, "%s:(%x)", p.Name, p.Value)
		}
	case Imm8, Imm16, Imm32, Imm64:
		fmt.Fprintf(w, "0x%x", p.Value)
	case Label:
		fmt.Fprint(w, p.Name)
	default:
		panic(fmt.Sprintf("unknown addr type: %v", p.Type))
	}
}

var addrTypeName = map[AddrType]string{
	None:  "None",
	Reg:   "Reg",
	Ind:   "Ind",
	Xmm:   "Xmm",
	Imm8:  "Imm8",
	Imm16: "Imm16",
	Imm32: "Imm32",
	Imm64: "Imm64",
	Rel8:  "Rel8",
	Rel16: "Rel16",
	Rel32: "Rel32",
	Label: "Label",
}

func (a AddrType) String() string {
	return addrTypeName[a]
}
