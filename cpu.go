package main

import "fmt"

type CPU struct {
	Reg *Registers
	mmu *MMU
}

func newCPU(mmu *MMU) *CPU {
	return &CPU{
		Reg: &Registers{},
		mmu: mmu,
	}
}

func (cpu *CPU) Step() int {
	opcode := cpu.mmu.Read(cpu.Reg.PC)
	cpu.Reg.PC++

	switch opcode {
	case 0x00:
		return 4
	case 0x3E: // LD A
		r := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.A = r
		cpu.Reg.PC++
		return 8
	case 0x06: // LD B
		r := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.B = r
		cpu.Reg.PC++
		return 8
	case 0x47:
		cpu.Reg.B = cpu.Reg.A
		return 4
	case 0x0E: // LD C
		r := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.C = r
		cpu.Reg.PC++
		return 8
	case 0x01:
		r1 := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.Reg.C = r1
		r2 := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.B = r2
		cpu.Reg.PC++
		return 12
	case 0x18: // JR
		offset := int8(cpu.mmu.Read(cpu.Reg.PC))
		cpu.Reg.PC++
		cpu.Reg.PC = uint16((int32(cpu.Reg.PC) + int32(offset)))
		return 12
	// ^ initial opcodes

	// ADD
	case 0x80:
		cpu.add(cpu.Reg.B)
		return 4
	case 0x81:
		cpu.add(cpu.Reg.C)
		return 4
	case 0x82:
		cpu.add(cpu.Reg.D)
		return 4
	case 0x83:
		cpu.add(cpu.Reg.E)
		return 4
	case 0x84:
		cpu.add(cpu.Reg.H)
		return 4
	case 0x85:
		cpu.add(cpu.Reg.L)
		return 4
	case 0x86:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.add(val)
		return 8
	case 0x87:
		cpu.add(cpu.Reg.A)
		return 4
	// ADC
	case 0x88:
		cpu.adc(cpu.Reg.B)
		return 4
	case 0x89:
		cpu.adc(cpu.Reg.C)
		return 4
	case 0x8A:
		cpu.adc(cpu.Reg.D)
		return 4
	case 0x8B:
		cpu.adc(cpu.Reg.E)
		return 4
	case 0x8C:
		cpu.adc(cpu.Reg.H)
		return 4
	case 0x8D:
		cpu.adc(cpu.Reg.L)
		return 4
	case 0x8E:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.adc(val)
		return 8
	case 0x8F:
		cpu.adc(cpu.Reg.A)
		return 4
	case 0xC6:
		val := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.add(val)
		return 8
	// SUB
	case 0x90:
		cpu.sub(cpu.Reg.B)
		return 4
	case 0x91:
		cpu.sub(cpu.Reg.C)
		return 4
	case 0x92:
		cpu.sub(cpu.Reg.D)
		return 4
	case 0x93:
		cpu.sub(cpu.Reg.E)
		return 4
	case 0x94:
		cpu.sub(cpu.Reg.H)
		return 4
	case 0x95:
		cpu.sub(cpu.Reg.L)
		return 4
	case 0x96:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.sub(val)
		return 8
	case 0x97:
		cpu.sub(cpu.Reg.A)
		return 4

	case 0x98: // SBC
		cpu.sbc(cpu.Reg.B)
		return 4
	case 0x99:
		cpu.sbc(cpu.Reg.C)
		return 4
	case 0x9A:
		cpu.sbc(cpu.Reg.D)
		return 4
	case 0x9B:
		cpu.sbc(cpu.Reg.E)
		return 4
	case 0x9C:
		cpu.sbc(cpu.Reg.H)
		return 4
	case 0x9D:
		cpu.sbc(cpu.Reg.L)
		return 4
	case 0x9E:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.sbc(val)
		return 8
	case 0x9F:
		cpu.sbc(cpu.Reg.A)
		return 4

	case 0xA0: // AND
		cpu.and(cpu.Reg.B)
		return 4
	case 0xA1:
		cpu.and(cpu.Reg.C)
		return 4
	case 0xA2:
		cpu.and(cpu.Reg.D)
		return 4
	case 0xA3:
		cpu.and(cpu.Reg.E)
		return 4
	case 0xA4:
		cpu.and(cpu.Reg.H)
		return 4
	case 0xA5:
		cpu.and(cpu.Reg.L)
		return 4
	case 0xA6:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.and(val)
		return 8
	case 0xA7:
		cpu.and(cpu.Reg.A)
		return 4

	case 0xA8: // XOR
		cpu.xor(cpu.Reg.B)
		return 4
	case 0xA9:
		cpu.xor(cpu.Reg.C)
		return 4
	case 0xAA:
		cpu.xor(cpu.Reg.D)
		return 4
	case 0xAB:
		cpu.xor(cpu.Reg.E)
		return 4
	case 0xAC:
		cpu.xor(cpu.Reg.H)
		return 4
	case 0xAD:
		cpu.xor(cpu.Reg.L)
		return 4
	case 0xAE:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.xor(val)
		return 8
	case 0xAF:
		cpu.xor(cpu.Reg.A)
		return 4

	case 0xB0: // OR
		cpu.or(cpu.Reg.B)
		return 4
	case 0xB1:
		cpu.or(cpu.Reg.C)
		return 4
	case 0xB2:
		cpu.or(cpu.Reg.D)
		return 4
	case 0xB3:
		cpu.or(cpu.Reg.E)
		return 4
	case 0xB4:
		cpu.or(cpu.Reg.H)
		return 4
	case 0xB5:
		cpu.or(cpu.Reg.L)
		return 4
	case 0xB6:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.or(val)
		return 8
	case 0xB7:
		cpu.or(cpu.Reg.A)
		return 4

	case 0xB8: // CP
		cpu.cp(cpu.Reg.B)
		return 4
	case 0xB9:
		cpu.cp(cpu.Reg.C)
		return 4
	case 0xBA:
		cpu.cp(cpu.Reg.D)
		return 4
	case 0xBB:
		cpu.cp(cpu.Reg.E)
		return 4
	case 0xBC:
		cpu.cp(cpu.Reg.H)
		return 4
	case 0xBD:
		cpu.cp(cpu.Reg.L)
		return 4
	case 0xBE:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		cpu.cp(val)
		return 8
	case 0xBF:
		cpu.cp(cpu.Reg.A)
		return 4

	case 0x05: // DEC
		cpu.Reg.B = cpu.dec(cpu.Reg.B)
		return 4
	case 0x0D:
		cpu.Reg.C = cpu.dec(cpu.Reg.C)
		return 4
	case 0x15:
		cpu.Reg.D = cpu.dec(cpu.Reg.D)
		return 4
	case 0x1D:
		cpu.Reg.E = cpu.dec(cpu.Reg.E)
		return 4
	case 0x25:
		cpu.Reg.H = cpu.dec(cpu.Reg.H)
		return 4
	case 0x2D:
		cpu.Reg.L = cpu.dec(cpu.Reg.L)
		return 4
	case 0x35:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		val = cpu.dec(val)
		cpu.mmu.Write(addr, val)
		return 12
	case 0x3D:
		cpu.Reg.A = cpu.dec(cpu.Reg.A)
		return 4

	case 0x04: // INC
		cpu.Reg.B = cpu.inc(cpu.Reg.B)
		return 4
	case 0x0C:
		cpu.Reg.C = cpu.inc(cpu.Reg.C)
		return 4
	case 0x14:
		cpu.Reg.D = cpu.inc(cpu.Reg.D)
		return 4
	case 0x1C:
		cpu.Reg.E = cpu.inc(cpu.Reg.E)
		return 4
	case 0x24:
		cpu.Reg.H = cpu.inc(cpu.Reg.H)
		return 4
	case 0x2C:
		cpu.Reg.L = cpu.inc(cpu.Reg.L)
		return 4
	case 0x34:
		addr := cpu.Reg.GetHL()
		val := cpu.mmu.Read(addr)
		val = cpu.inc(val)
		cpu.mmu.Write(addr, val)
		return 12
	case 0x3C:
		cpu.Reg.A = cpu.inc(cpu.Reg.A)
		return 4

		// LD B,r
	case 0x40:
		return 4 // LD B,B (NOP-like)
	case 0x41:
		cpu.Reg.B = cpu.Reg.C
		return 4 // LD B,C
	case 0x42:
		cpu.Reg.B = cpu.Reg.D
		return 4 // LD B,D
	case 0x43:
		cpu.Reg.B = cpu.Reg.E
		return 4 // LD B,E
	case 0x44:
		cpu.Reg.B = cpu.Reg.H
		return 4 // LD B,H
	case 0x45:
		cpu.Reg.B = cpu.Reg.L
		return 4 // LD B,L
	case 0x46:
		cpu.Reg.B = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8 // LD B,(HL)

	// LD C,r
	case 0x48:
		cpu.Reg.C = cpu.Reg.B
		return 4
	case 0x49:
		return 4 // NOP
	case 0x4A:
		cpu.Reg.C = cpu.Reg.D
		return 4
	case 0x4B:
		cpu.Reg.C = cpu.Reg.E
		return 4
	case 0x4C:
		cpu.Reg.C = cpu.Reg.H
		return 4
	case 0x4D:
		cpu.Reg.C = cpu.Reg.L
		return 4
	case 0x4E:
		cpu.Reg.C = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x4F:
		cpu.Reg.C = cpu.Reg.A
		return 4

	// LD D,r
	case 0x50:
		cpu.Reg.D = cpu.Reg.B
		return 4
	case 0x51:
		cpu.Reg.D = cpu.Reg.C
		return 4
	case 0x52:
		return 4 // NOP
	case 0x53:
		cpu.Reg.D = cpu.Reg.E
		return 4
	case 0x54:
		cpu.Reg.D = cpu.Reg.H
		return 4
	case 0x55:
		cpu.Reg.D = cpu.Reg.L
		return 4
	case 0x56:
		cpu.Reg.D = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x57:
		cpu.Reg.D = cpu.Reg.A
		return 4

	// LD E,r
	case 0x58:
		cpu.Reg.E = cpu.Reg.B
		return 4
	case 0x59:
		cpu.Reg.E = cpu.Reg.C
		return 4
	case 0x5A:
		cpu.Reg.E = cpu.Reg.D
		return 4
	case 0x5B:
		return 4 // NOP
	case 0x5C:
		cpu.Reg.E = cpu.Reg.H
		return 4
	case 0x5D:
		cpu.Reg.E = cpu.Reg.L
		return 4
	case 0x5E:
		cpu.Reg.E = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x5F:
		cpu.Reg.E = cpu.Reg.A
		return 4

	// LD H,r
	case 0x60:
		cpu.Reg.H = cpu.Reg.B
		return 4
	case 0x61:
		cpu.Reg.H = cpu.Reg.C
		return 4
	case 0x62:
		cpu.Reg.H = cpu.Reg.D
		return 4
	case 0x63:
		cpu.Reg.H = cpu.Reg.E
		return 4
	case 0x64:
		return 4 // NOP
	case 0x65:
		cpu.Reg.H = cpu.Reg.L
		return 4
	case 0x66:
		cpu.Reg.H = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x67:
		cpu.Reg.H = cpu.Reg.A
		return 4

	// LD L,r
	case 0x68:
		cpu.Reg.L = cpu.Reg.B
		return 4
	case 0x69:
		cpu.Reg.L = cpu.Reg.C
		return 4
	case 0x6A:
		cpu.Reg.L = cpu.Reg.D
		return 4
	case 0x6B:
		cpu.Reg.L = cpu.Reg.E
		return 4
	case 0x6C:
		cpu.Reg.L = cpu.Reg.H
		return 4
	case 0x6D:
		return 4 // NOP
	case 0x6E:
		cpu.Reg.L = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x6F:
		cpu.Reg.L = cpu.Reg.A
		return 4

	// LD (HL),r - Store register into memory at HL
	case 0x70:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.B)
		return 8
	case 0x71:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.C)
		return 8
	case 0x72:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.D)
		return 8
	case 0x73:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.E)
		return 8
	case 0x74:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.H)
		return 8
	case 0x75:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.L)
		return 8
	case 0x76: // HALT - special instruction, not a load
		return 4
	case 0x77:
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.A)
		return 8

	// LD A,r
	case 0x78:
		cpu.Reg.A = cpu.Reg.B
		return 4
	case 0x79:
		cpu.Reg.A = cpu.Reg.C
		return 4
	case 0x7A:
		cpu.Reg.A = cpu.Reg.D
		return 4
	case 0x7B:
		cpu.Reg.A = cpu.Reg.E
		return 4
	case 0x7C:
		cpu.Reg.A = cpu.Reg.H
		return 4
	case 0x7D:
		cpu.Reg.A = cpu.Reg.L
		return 4
	case 0x7E:
		cpu.Reg.A = cpu.mmu.Read(cpu.Reg.GetHL())
		return 8
	case 0x7F:
		return 4 // NOP

	case 0x16: // LD D,n
		cpu.Reg.D = cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		return 8
	case 0x1E: // LD E,n
		cpu.Reg.E = cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		return 8
	case 0x26: // LD H,n
		cpu.Reg.H = cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		return 8
	case 0x2E: // LD L,n
		cpu.Reg.L = cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		return 8
	case 0x36: // LD (HL),n
		val := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.mmu.Write(cpu.Reg.GetHL(), val)
		return 12

	// LD A with special addressing
	case 0x0A: // LD A,(BC)
		cpu.Reg.A = cpu.mmu.Read(cpu.Reg.GetBC())
		return 8
	case 0x1A: // LD A,(DE)
		cpu.Reg.A = cpu.mmu.Read(cpu.Reg.GetDE())
		return 8
	case 0xFA: // LD A,(nn)
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		addr := (uint16(hi) << 8) | uint16(lo)
		cpu.Reg.A = cpu.mmu.Read(addr)
		return 16

	case 0x02: // LD (BC),A
		cpu.mmu.Write(cpu.Reg.GetBC(), cpu.Reg.A)
		return 8
	case 0x12: // LD (DE),A
		cpu.mmu.Write(cpu.Reg.GetDE(), cpu.Reg.A)
		return 8
	case 0xEA: // LD (nn),A
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		addr := (uint16(hi) << 8) | uint16(lo)
		cpu.mmu.Write(addr, cpu.Reg.A)
		return 16

	// LD with HL increment/decrement
	case 0x22: // LD (HL+),A or LDI (HL),A
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.A)
		cpu.Reg.SetHL(cpu.Reg.GetHL() + 1)
		return 8
	case 0x2A: // LD A,(HL+) or LDI A,(HL)
		cpu.Reg.A = cpu.mmu.Read(cpu.Reg.GetHL())
		cpu.Reg.SetHL(cpu.Reg.GetHL() + 1)
		return 8
	case 0x32: // LD (HL-),A or LDD (HL),A
		cpu.mmu.Write(cpu.Reg.GetHL(), cpu.Reg.A)
		cpu.Reg.SetHL(cpu.Reg.GetHL() - 1)
		return 8
	case 0x3A: // LD A,(HL-) or LDD A,(HL)
		cpu.Reg.A = cpu.mmu.Read(cpu.Reg.GetHL())
		cpu.Reg.SetHL(cpu.Reg.GetHL() - 1)
		return 8

	// 16-bit loads
	case 0x11: // LD DE,nn
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.Reg.SetDE((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0x21: // LD HL,nn
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.Reg.SetHL((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0x31: // LD SP,nn
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.Reg.SP = (uint16(hi) << 8) | uint16(lo)
		return 12

	// High memory operations
	case 0xE0: // LDH (n),A
		offset := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.mmu.Write(0xFF00|uint16(offset), cpu.Reg.A)
		return 12
	case 0xF0: // LDH A,(n)
		offset := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.Reg.A = cpu.mmu.Read(0xFF00 | uint16(offset))
		return 12
	case 0xE2: // LD (C),A
		cpu.mmu.Write(0xFF00|uint16(cpu.Reg.C), cpu.Reg.A)
		return 8
	case 0xF2: // LD A,(C)
		cpu.Reg.A = cpu.mmu.Read(0xFF00 | uint16(cpu.Reg.C))
		return 8

	case 0xC3: // JP
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
		return 16
	default:
		panic(fmt.Sprintf("Unknown Opcode: 0x%02X at PC: 0x%04X", opcode, cpu.Reg.PC-1))
	}
}

func (cpu *CPU) add(val uint8) {
	sum := uint16(cpu.Reg.A) + uint16(val)

	cpu.Reg.SetZero((sum & 0xFF) == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(sum > 0xFF)

	nibbleSum := (cpu.Reg.A & 0x0F) + (val & 0x0F)
	cpu.Reg.SetHalfCarry(nibbleSum > 0x0F)

	cpu.Reg.A = uint8(sum)
}

func (cpu *CPU) adc(val uint8) {
	carry := uint16(0)
	if cpu.Reg.GetCarry() {
		carry = 1
	}
	sum := uint16(cpu.Reg.A) + uint16(val) + carry

	cpu.Reg.SetZero((sum & 0xFF) == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(sum > 0xFF)

	nibbleSum := (cpu.Reg.A & 0x0F) + (val & 0x0F) + uint8(carry)
	cpu.Reg.SetHalfCarry(nibbleSum > 0x0F)

	cpu.Reg.A = uint8(sum)
}

func (cpu *CPU) sub(val uint8) {
	res := uint16(cpu.Reg.A) - uint16(val)

	cpu.Reg.SetZero((res & 0xFF) == 0)
	cpu.Reg.SetSubtract(true)
	cpu.Reg.SetCarry(val > cpu.Reg.A)

	cpu.Reg.SetHalfCarry((val & 0x0F) > (cpu.Reg.A & 0x0F))

	cpu.Reg.A = uint8(res)
}

func (cpu *CPU) sbc(val uint8) {
	carry := uint16(0)
	if cpu.Reg.GetCarry() {
		carry = 1
	}
	res := uint16(cpu.Reg.A) - uint16(val) - carry
	cpu.Reg.SetZero((res & 0xFF) == 0)
	cpu.Reg.SetSubtract(true)
	cpu.Reg.SetCarry(uint16(cpu.Reg.A) < (uint16(val) + carry))
	cpu.Reg.SetHalfCarry((uint16(cpu.Reg.A) & 0x0F) < ((uint16(val) & 0x0F) + carry))

	cpu.Reg.A = uint8(res)
}

func (cpu *CPU) xor(val uint8) {
	cpu.Reg.A ^= val

	cpu.Reg.SetZero(cpu.Reg.A == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(false)
	cpu.Reg.SetHalfCarry(false)
}

func (cpu *CPU) and(val uint8) {
	cpu.Reg.A &= val

	cpu.Reg.SetZero(cpu.Reg.A == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(false)
	cpu.Reg.SetHalfCarry(true)
}

func (cpu *CPU) or(val uint8) {
	cpu.Reg.A |= val

	cpu.Reg.SetZero(cpu.Reg.A == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(false)
	cpu.Reg.SetHalfCarry(false)
}

func (cpu *CPU) cp(val uint8) {
	res := uint16(cpu.Reg.A) - uint16(val)

	cpu.Reg.SetZero((res & 0xFF) == 0)
	cpu.Reg.SetSubtract(true)
	cpu.Reg.SetCarry(val > cpu.Reg.A)
	cpu.Reg.SetHalfCarry((val & 0x0F) > (cpu.Reg.A & 0x0F))
}

func (cpu *CPU) inc(reg uint8) uint8 {
	reg += 1
	cpu.Reg.SetZero(reg == 0)
	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetHalfCarry((reg & 0x0F) == 0x00)
	return reg
}

func (cpu *CPU) dec(reg uint8) uint8 {
	reg -= 1
	cpu.Reg.SetZero(reg == 0)
	cpu.Reg.SetSubtract(true)
	cpu.Reg.SetHalfCarry((reg & 0x0F) == 0x0F)
	return reg
}
