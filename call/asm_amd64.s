// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "zasm_GOOS_GOARCH.h"

// Called from reflection library.  Mimics morestack,
// reuses stack growth code to create a frame
// with the desired args running the desired function.
//
// func Call(fn *byte, arg *byte, argsize uint32).
TEXT call·Call(SB), 7, $0
	get_tls(CX)
	MOVQ	m(CX), BX

	// Save our caller's state as the PC and SP to
	// restore when returning from f.
	MOVQ	0(SP), AX	// our caller's PC
	MOVQ	AX, (m_morebuf+gobuf_pc)(BX)
	LEAQ	8(SP), AX	// our caller's SP
	MOVQ	AX, (m_morebuf+gobuf_sp)(BX)
	MOVQ	g(CX), AX
	MOVQ	AX, (m_morebuf+gobuf_g)(BX)

	// Set up morestack arguments to call f on a new stack.
	// We set f's frame size to 1, as a hint to newstack
	// that this is a call from reflect·call.
	// If it turns out that f needs a larger frame than
	// the default stack, f's usual stack growth prolog will
	// allocate a new segment (and recopy the arguments).
	MOVQ	8(SP), AX	// fn
	MOVQ	16(SP), DX	// arg frame
	MOVL	24(SP), CX	// arg size

	MOVQ	AX, m_morepc(BX)	// f's PC
	MOVQ	DX, m_moreargp(BX)	// argument frame pointer
	MOVL	CX, m_moreargsize(BX)	// f's argument size
	MOVL	$1, m_moreframesize(BX)	// f's frame size

	// Call newstack on m->g0's stack.
	MOVQ	m_g0(BX), BP
	get_tls(CX)
	MOVQ	BP, g(CX)
	MOVQ	(g_sched+gobuf_sp)(BP), SP
	CALL	runtime·newstack(SB)
	MOVQ	$0, 0x1103	// crash if newstack returns
	RET
