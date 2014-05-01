package i64

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Instruction is an amd64 asssembly instruction.
type Instruction struct {
	Op   Op   // instruction opcode
	From Addr // source address
	To   Addr // destination address
}

const (
	rexW = 0x08
	rexR = 0x04
	rexX = 0x02
	rexB = 0x01
)

// ins is a laid out Instruction. After calling make, an ins looks very
// similar to the laid out bytes for an op. (Passing ins to printf is helpful
// in debugging.)
type ins struct {
	ins          *Instruction
	codeblock    int
	codeblockEnd int

	rex       uint8 // just the lower four bits.
	c0        uint8 // only used if 66, f3, or f2.
	c1        uint8
	c2        uint8 // only used if code==0x0f.
	modRM     bool
	modRMmod  uint8 // 2 bits, 11b for direct addressing.
	modRMreg  uint8 // 3 bits, fourth bit is in REX.R
	modRMrm   uint8 // 3 bits, fourth bit is in REX.B
	sib       bool
	sibScale  uint8
	sibIndex  uint8 // 3 bits, fourth bit is set as REX.X
	sibBase   uint8 // 3 bits, fourth bit is set as REX.B
	dispWidth int   // num bytes, 8, 16, 32.
	disp      uint64
	immWidth  int // num bytes, 8, 16, 32, 64.
	imm       uint64
}

func (c *ins) make(p *Instruction) error {
	c.ins = p
	if p.Op == LABEL {
		return nil
	}
	optabKey := opKey{p.Op, p.From.Type, p.To.Type}
	optabVal, ok := optab[optabKey]
	if !ok {
		return fmt.Errorf("unknown combination: %v", optabKey)
	}
	c.c0 = optabVal.c0
	c.c1 = optabVal.c1
	c.c2 = optabVal.c2
	if optabVal.addReg {
		if err := c.addRegToOp(); err != nil {
			return err
		}
	}
	if optabVal.rex {
		c.rex |= rexW
	}
	if err := c.makeMod(optabVal.mod); err != nil {
		return err
	}
	c.makeImm(p.From)
	c.makeImm(p.To)
	return nil
}

func (c *ins) addRegToOp() error {
	var r Addr
	if c.ins.To.Type == Reg {
		r = c.ins.To
	} else {
		r = c.ins.From
	}
	reg, ok := r.Value.(Register)
	if !ok {
		return fmt.Errorf("address must be register")
	}
	if reg >= R8 && reg <= R15 {
		return fmt.Errorf("R8-15 unsupported for addReg")
	}
	if reg >= X0 && reg <= X15 {
		return fmt.Errorf("Xmm unsupported for addReg")
	}
	c.c1 += uint8(reg - AX)
	return nil
}

func (c *ins) makeMod(bits modBits) error {
	if bits == modNone {
		return nil
	}
	c.modRM = true

	// r1 is direct address, r2 may be indirect address
	var r1, r2 Addr
	if c.ins.From.Type == Ind {
		r1 = c.ins.To
		r2 = c.ins.From
	} else {
		r1 = c.ins.From
		r2 = c.ins.To
	}
	if r1.Type == Ind {
		return fmt.Errorf("only one register can be indirect")
	}

	switch bits {
	case modDefault:
		if reg, ok := r1.Value.(Register); ok {
			switch {
			case reg >= AX && reg <= DI:
				c.modRMreg = uint8(reg - AX)
			case reg >= R8 && reg <= R15:
				c.rex |= rexR
				c.modRMreg = uint8(reg - R8)
			case reg >= X0 && reg <= X7:
				c.modRMreg = uint8(reg - X0)
			case reg >= X8 && reg <= X15:
				c.rex |= rexR // TODO is this right?
				c.modRMreg = uint8(reg - X8)
			}
		}
	case mod0, mod1, mod2, mod3, mod4, mod5, mod6, mod7, mod8:
		c.modRMreg = uint8(bits - mod0)
	case modNone:
		panic("unreachable")
	}

	if r2.Type == Ind {
		if err := c.indirectAddress(r2); err != nil {
			return err
		}
	} else {
		if err := c.directAddress(r2); err != nil {
			return err
		}
	}
	return nil
}

func (c *ins) indirectAddress(r2 Addr) error {
	// TODO handle a displacement with no register. (i.e. mod=00, r/m=100b)

	if r2.Disp == 0 {
		c.modRMmod = 0
	} else if r2.Disp <= 0xff {
		c.modRMmod = 0x01
		c.dispWidth = 8
		c.disp = r2.Disp
	} else if r2.Disp <= 0xffffffff {
		c.modRMmod = 0x02
		c.dispWidth = 32
		c.disp = r2.Disp
	} else {
		return fmt.Errorf("TODO handle disp with SIB scaling: %x", r2.Disp)
	}
	// Assuming we are not using the scale and index, (TODO above) let's proceed.

	// Skip SIB if rm != 100b.
	reg := r2.Value.(Register)
	switch {
	case reg >= R8 && reg <= R15 && reg-R8 != 0x4:
		c.rex |= rexB
		c.modRMrm = uint8(reg - R8)
		return nil
	case reg >= X0 && reg <= X7 && reg-X0 != 0x4:
		c.rex |= rexB
		c.modRMrm = uint8(reg - X0)
		return nil
	case reg != 0x4:
		c.modRMrm = uint8(reg)
		return nil
	}

	// Add a SIB.
	c.sib = true
	c.modRMrm = 0x4

	if reg >= R8 && reg <= R15 {
		c.rex |= rexB
		reg -= R8
	} else if reg >= X0 && reg <= X7 {
		reg -= X0
	} else if reg >= X8 && reg <= X15 {
		c.rex |= rexB
		reg -= X8
	}
	c.sibIndex = 0x4 // no index register
	c.sibBase = uint8(reg)
	return nil
}

func (c *ins) directAddress(r2 Addr) error {
	c.modRMmod = 0x3
	if reg, ok := r2.Value.(Register); ok {
		if reg >= R8 && reg <= R15 {
			c.rex |= rexB
			reg -= R8
		} else if reg >= X0 && reg <= X7 {
			reg -= X0
		} else if reg >= X8 && reg <= X15 {
			c.rex |= rexB
			reg -= X8
		}
		c.modRMrm = uint8(reg)
	}
	return nil
}

func (c *ins) makeImm(r Addr) {
	switch r.Type {
	case Imm8, Rel8:
		c.immWidth = 8
		c.imm = r.valueUint64()
	case Imm16, Rel16:
		c.immWidth = 16
		c.imm = r.valueUint64()
	case Imm32, Rel32:
		c.immWidth = 32
		c.imm = r.valueUint64()
	case Imm64:
		c.immWidth = 64
		c.imm = r.valueUint64()
	}
}

func (c *ins) writeTo(w io.Writer) (n int64, err error) {
	if c.ins.Op == LABEL {
		return
	}
	var bufArray [6]byte
	buf := bufArray[:0]
	if c.rex != 0 {
		buf = append(buf, 0x48|c.rex)
	}
	if c.c0 != 0 {
		buf = append(buf, c.c0)
	}
	buf = append(buf, c.c1)
	if c.c1 == 0x0f {
		buf = append(buf, c.c2)
	}
	if c.modRM {
		buf = append(buf, c.modRMmod<<6|c.modRMreg<<3|c.modRMrm)
	}
	if c.sib {
		buf = append(buf, c.sibScale<<6|c.sibIndex<<3|c.sibBase)
	}
	if _, err := w.Write(buf); err != nil {
		return n, err
	}
	n += int64(len(buf))
	if c.dispWidth > 0 {
		if err := writeBytes(w, c.dispWidth, c.disp); err != nil {
			return n, err
		}
		n += int64(c.dispWidth)
	}
	if c.immWidth > 0 {
		if err := writeBytes(w, c.immWidth, c.imm); err != nil {
			return n, err
		}
		n += int64(c.immWidth)
	}
	return n, nil
}

func (c *ins) printText(w io.Writer) {
	name := opName[c.ins.Op]
	fmt.Fprint(w, name)
	const namePad = "      "
	if len(name) < len(namePad) {
		fmt.Fprint(w, namePad[len(name):])
	}

	c.ins.From.printText(w, c.codeblockEnd)
	fmt.Fprint(w, ",")
	c.ins.To.printText(w, c.codeblockEnd)
}

func writeBytes(w io.Writer, width int, v uint64) error {
	switch width {
	case 8:
		return binary.Write(w, binary.LittleEndian, uint8(v))
	case 16:
		return binary.Write(w, binary.LittleEndian, uint16(v))
	case 32:
		return binary.Write(w, binary.LittleEndian, uint32(v))
	case 64:
		return binary.Write(w, binary.LittleEndian, uint64(v))
	}
	panic(fmt.Sprintf("unknown width: %d", width))
}
