package main

const (
	ZERO_FLAG uint8 = 0x80
	SUBTRACT_FLAG uint8 = 0x40
	HALF_CARRY_FLAG uint8 = 0x20
    CARRY_FLAG     uint8 = 0x10
)

type Registers struct {
	A, B, C, D, E, F, H, L uint8

	SP, PC uint16
	IME bool // Interrupt Master Enable
}

// -----------------------------------------------------------------------------
// VIRTUAL 16-BIT REGISTERS (Pairing Logic)
// -----------------------------------------------------------------------------

func (r *Registers) GetBC() uint16 {
	return uint16(r.B) << 8 | uint16(r.C)
}

func (r *Registers) SetBC(val uint16) {
	r.B = uint8((val & 0xFF00) >> 8)
	r.C = uint8(val & 0x00FF)
}

func (r *Registers) GetDE() uint16 {
	return uint16(r.D) << 8 | uint16(r.E)
}

func (r *Registers) SetDE(val uint16) {
	r.D = uint8((val & 0xFF00) >> 8)
	r.E = uint8(val & 0x00FF)
}

func (r *Registers) GetHL() uint16 {
	return uint16(r.H) << 8 | uint16(r.L)
}

func (r *Registers) SetHL(val uint16) {
	r.H = uint8((val & 0xFF00) >> 8)
	r.L = uint8(val & 0x00FF)
}

func (r* Registers) GetAF() uint16 {
	return uint16(r.A) << 8 | uint16(r.F)
}

func (r* Registers) SetAF(val uint16) {
	r.A = uint8((val & 0xFF00) >> 8)
	r.F = uint8(val & 0x00F0)
}




func (r* Registers) GetZero() bool {
	return (r.F & ZERO_FLAG == ZERO_FLAG)
}

func (r* Registers) SetZero(val bool) {
	if val {
		r.F = r.F | ZERO_FLAG
	} else {
		r.F &= ^ZERO_FLAG
	}
}

func (r* Registers) GetSubtract() bool {
	return (r.F & SUBTRACT_FLAG == SUBTRACT_FLAG)
}

func (r* Registers) SetSubtract(val bool) {
	if val {
		r.F = r.F | SUBTRACT_FLAG
	} else {
		r.F &= ^SUBTRACT_FLAG
	}
}

func (r* Registers) GetHalfCarry() bool {
	return (r.F & HALF_CARRY_FLAG == HALF_CARRY_FLAG)
}

func (r* Registers) SetHalfCarry(val bool) {
	if val {
		r.F = r.F | HALF_CARRY_FLAG
	} else {
		r.F &= ^HALF_CARRY_FLAG
	}
}

func (r* Registers) GetCarry() bool {
	return (r.F & CARRY_FLAG == CARRY_FLAG)
}

func (r* Registers) SetCarry(val bool) {
	if val {
		r.F = r.F | CARRY_FLAG
	} else {
		r.F &= ^CARRY_FLAG
	}
}