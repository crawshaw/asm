package i64

// Op is an amd64 operation. Mnemonics closely follow the amd64 manual.
type Op int

const (
	LABEL Op = iota // not a real Op

	ADD
	OR
	ADC
	SBB
	AND
	SUB
	XOR
	CMP

	ADDL
	ORL
	ADCL
	SBBL
	ANDL
	SUBL
	XORL
	CMPL

	ADDQ
	ORQ
	ADCQ
	SBBQ
	ANDQ
	SUBQ
	XORQ
	CMPQ

	IMULL
	IMULQ

	IDIVL
	IDIVQ

	MOVB
	MOVL
	MOVQ

	LEAL
	LEAQ

	CALL
	RET
	JMP
	JE
	JNE
	JHI
	PUSHL
	PUSHQ
	POPL
	POPQ
	LEA

	MOVSS
	ADDSS
	MULSS
	SUBSS
	MINSS
	DIVSS
	MAXSS

	MOVSD
	ADDSD
	MULSD
	SUBSD
	MINSD
	DIVSD
	MAXSD

	lastOp
)

var opName = map[Op]string{
	LABEL: "LABEL",

	ADD: "ADD",
	OR:  "OR",
	ADC: "ADC",
	SBB: "SBB",
	AND: "AND",
	SUB: "SUB",
	XOR: "XOR",
	CMP: "CMP",

	ADDL: "ADDL",
	ORL:  "ORL",
	ADCL: "ADCL",
	SBBL: "SBBL",
	ANDL: "ANDL",
	SUBL: "SUBL",
	XORL: "XORL",
	CMPL: "CMPL",

	ADDQ: "ADDQ",
	ORQ:  "ORQ",
	ADCQ: "ADCQ",
	SBBQ: "SBBQ",
	ANDQ: "ANDQ",
	SUBQ: "SUBQ",
	XORQ: "XORQ",
	CMPQ: "CMPQ",

	IMULL: "IMULL",
	IMULQ: "IMULQ",

	IDIVL: "IDIVL",
	IDIVQ: "IDIVQ",

	MOVB: "MOVB",
	MOVL: "MOVL",
	MOVQ: "MOVQ",

	LEAL: "LEAL",
	LEAQ: "LEAQ",

	CALL:  "CALL",
	RET:   "RET",
	JMP:   "JMP",
	JE:    "JE",
	JNE:   "JNE",
	JHI:   "JHI",
	PUSHL: "PUSHL",
	PUSHQ: "PUSHQ",
	POPL:  "POPL",
	POPQ:  "POPQ",
	LEA:   "LEA",

	MOVSS: "MOVSS",
	ADDSS: "ADDSS",
	MULSS: "MULSS",
	SUBSS: "SUBSS",
	MINSS: "MINSS",
	DIVSS: "DIVSS",
	MAXSS: "MAXSS",

	MOVSD: "MOVSD",
	ADDSD: "ADDSD",
	MULSD: "MULSD",
	SUBSD: "SUBSD",
	MINSD: "MINSD",
	DIVSD: "DIVSD",
	MAXSD: "MAXSD",
}

func (op Op) String() string {
	return opName[op]
}
