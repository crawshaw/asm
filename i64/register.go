package i64

// Register is an amd64 register.
type Register int

// Addr makes an Addr representing a register.
func (r Register) Addr() Addr {
	addrType := Reg
	if r >= X0 && r <= X15 {
		addrType = Xmm
	}
	return Addr{addrType, r, 0, ""}
}

// Ind makes an Addr representing a memory address pointed to by the given register.
func (r Register) Ind(disp uint64) Addr { return Addr{Ind, r, disp, ""} }

// String returns the name of the register.
func (r Register) String() string { return registerName[r] }

const (
	AX Register = iota
	CX
	DX
	BX
	SP
	BP
	SI
	DI
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15

	X0
	X1
	X2
	X3
	X4
	X5
	X6
	X7
	// TODO(crawshaw): encoding of X8-15 is untested.
	X8
	X9
	X10
	X11
	X12
	X13
	X14
	X15
)

var registerName = map[Register]string{
	AX:  "AX",
	CX:  "CX",
	DX:  "DX",
	BX:  "BX",
	SP:  "SP",
	BP:  "BP",
	SI:  "SI",
	DI:  "DI",
	R8:  "R8",
	R9:  "R9",
	R10: "R10",
	R11: "R11",
	R12: "R12",
	R13: "R13",
	R14: "R14",
	R15: "R15",
	X0:  "X0",
	X1:  "X1",
	X2:  "X2",
	X3:  "X3",
	X4:  "X4",
	X5:  "X5",
	X6:  "X6",
	X7:  "X7",
	X8:  "X8",
	X9:  "X9",
	X10: "X10",
	X11: "X11",
	X12: "X12",
	X13: "X13",
	X14: "X14",
	X15: "X15",
}
