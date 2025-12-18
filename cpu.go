package main

import "fmt"

type CPU struct {
	Reg    *Registers
	mmu    *MMU
	halted bool
}

func newCPU(mmu *MMU) *CPU {
	return &CPU{
		Reg: &Registers{},
		mmu: mmu,
	}
}

func (cpu *CPU) HandleInterrupts() int {
	IF := cpu.mmu.Read(0xFF0F) // Interrupt Flag
	IE := cpu.mmu.Read(0xFFFF) // Interrupt Enable

	// Check if there's a pending interrupt (this wakes from HALT even if IME is false)
	interrupts := IF & IE
	if interrupts != 0 && cpu.halted {
		cpu.halted = false
	}

	if !cpu.Reg.IME {
		return 0
	}

	if interrupts == 0 {
		return 0
	}

	// Check VBLANK interrupt (bit 0)
	if interrupts&0x01 != 0 {
		cpu.Reg.IME = false            // Disable interrupts
		cpu.mmu.Write(0xFF0F, IF&0xFE) // Clear VBLANK bit in IF

		// Push PC to stack
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8((cpu.Reg.PC&0xFF00)>>8))
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8(cpu.Reg.PC&0x00FF))

		// Jump to VBLANK handler
		cpu.Reg.PC = 0x0040
		return 20 // Interrupt handling takes 20 cycles
	}

	return 0
}

func (cpu *CPU) Step() int {
	// If halted, don't execute, just consume cycles
	if cpu.halted {
		return 4
	}

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
		cpu.Reg.PC = uint16(int32(cpu.Reg.PC) + int32(offset))
		return 12
	case 0x28: // JR Z - Jump relative if zero flag is set
		offset := int8(cpu.mmu.Read(cpu.Reg.PC))
		cpu.Reg.PC++
		if cpu.Reg.GetZero() {
			cpu.Reg.PC = uint16(int32(cpu.Reg.PC) + int32(offset))
			return 12
		}
		return 8
	case 0x19: // ADD HL,DE - Add DE to HL
		de := cpu.Reg.GetDE()
		cpu.add16(de)
		return 8
	case 0x20: // JR NZ,e - Jump relative if not zero (COPILOT)
		offset := int8(cpu.mmu.Read(cpu.Reg.PC))
		cpu.Reg.PC++
		if !cpu.Reg.GetZero() {
			cpu.Reg.PC = uint16(int32(cpu.Reg.PC) + int32(offset))
			return 12
		}
		return 8
	case 0x23: // INC HL - Increment 16-bit register HL
		hl := cpu.Reg.GetHL()
		cpu.Reg.SetHL(hl + 1)
		return 8
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

	case 0x0B: // DEC BC - Decrement 16-bit register BC
		bc := cpu.Reg.GetBC()
		bc--
		cpu.Reg.SetBC(bc)
		return 8
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
	case 0x03: // INC BC
		bc := cpu.Reg.GetBC()
		cpu.Reg.SetBC(bc + 1)
		return 8
	case 0x13: // INC DE
		de := cpu.Reg.GetDE()
		cpu.Reg.SetDE(de + 1)
		return 8
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
		cpu.halted = true
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
	case 0x2F: // CPL - Complement A (flip all bits) (COPILOT)
		cpu.Reg.A = ^cpu.Reg.A
		cpu.Reg.SetSubtract(true)
		cpu.Reg.SetHalfCarry(true)
		return 4
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
		if offset == 0x40 { // LCD Control
			fmt.Printf("LCD Control write at PC: 0x%04X, value: 0x%02X\n", cpu.Reg.PC-2, cpu.Reg.A)
		}
		cpu.mmu.Write(0xFF00|uint16(offset), cpu.Reg.A)
		return 12
	case 0xC1: // POP BC
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.SetBC((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0xD1: // POP DE
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.SetDE((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0xE1: // POP HL - Pop 16-bit value from stack into HL (COPILOT)
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.SetHL((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0xF1: // POP HL - Pop 16-bit value from stack into HL (COPILOT)
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.SetAF((uint16(hi) << 8) | uint16(lo))
		return 12
	case 0xE6: // AND n - AND A with immediate value
		val := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.and(val)
		return 8
	case 0xE9:
		hl := cpu.Reg.GetHL()
		cpu.Reg.PC = hl
		return 4
	case 0xEF: // RST 28H - Call to address 0x0028 (COPILOT)
		// Push return address onto stack
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8(cpu.Reg.PC>>8))
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8(cpu.Reg.PC&0xFF))
		// Jump to 0x0028
		cpu.Reg.PC = 0x0028
		return 16
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
	case 0xC2: // JP NZ - Jump if zero flag is not set
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		if !cpu.Reg.GetZero() {
			cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
			return 16
		}
		return 12
	case 0xC3: // JP
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
		return 16
	case 0xCA: // JP Z - Jump if zero flag is set
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		if cpu.Reg.GetZero() {
			cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
			return 16
		}
		return 12
	case 0xCD: // CALL nn - Call subroutine (COPILOT)
		lo := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		hi := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		// Push return address onto stack
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8(cpu.Reg.PC>>8)) // High byte
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, uint8(cpu.Reg.PC&0xFF)) // Low byte
		// Jump to target address
		cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
		return 24
	case 0xC5:
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.B)
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.C)
		return 16
	case 0xD5: // PUSH DE - Push DE onto stack (COPILOT)
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.D)
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.E)
		return 16
	case 0xE5: // PUSH HL
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.H)
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.L)
		return 16
	case 0xF5: // PUSH AF
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.A)
		cpu.Reg.SP--
		cpu.mmu.Write(cpu.Reg.SP, cpu.Reg.F)
		return 16
	case 0xC0: // RET NZ - Return if zero flag not set
		if !cpu.Reg.GetZero() {
			lo := cpu.mmu.Read(cpu.Reg.SP)
			cpu.Reg.SP++
			hi := cpu.mmu.Read(cpu.Reg.SP)
			cpu.Reg.SP++
			cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
			return 20
		}
		return 8
	case 0xC8: // RET Z - Return if zero flag set
		if cpu.Reg.GetZero() {
			lo := cpu.mmu.Read(cpu.Reg.SP)
			cpu.Reg.SP++
			hi := cpu.mmu.Read(cpu.Reg.SP)
			cpu.Reg.SP++
			cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
			return 20
		}
		return 8
	case 0xC9: // RET - Return from subroutine
		// Pop return address from stack
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
		return 16
	case 0xD9: // RETI - Return from interrupt (COPILOT)
		// Pop return address from stack
		lo := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		hi := cpu.mmu.Read(cpu.Reg.SP)
		cpu.Reg.SP++
		cpu.Reg.PC = (uint16(hi) << 8) | uint16(lo)
		cpu.Reg.IME = true // Re-enable interrupts
		return 16
	case 0xCB: // CB prefix - extended instructions
		return cpu.executeCB()
	case 0xF3: // DI - Disable interrupts (COPILOT)
		cpu.Reg.IME = false
		return 4
	case 0xFB: // EI - Enable interrupts
		cpu.Reg.IME = true
		return 4
	case 0xFE: // CP n - Compare A with immediate value (COPILOT)
		val := cpu.mmu.Read(cpu.Reg.PC)
		cpu.Reg.PC++
		cpu.cp(val)
		return 8
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

func (cpu *CPU) add16(val uint16) {
	hl := cpu.Reg.GetHL()
	result := uint32(hl) + uint32(val)

	cpu.Reg.SetSubtract(false)
	cpu.Reg.SetCarry(result > 0xFFFF)
	cpu.Reg.SetHalfCarry(((hl & 0x0FFF) + (val & 0x0FFF)) > 0x0FFF)

	cpu.Reg.SetHL(uint16(result))
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

func (cpu *CPU) executeCB() int {
	opcode := cpu.mmu.Read(cpu.Reg.PC)
	cpu.Reg.PC++

	switch opcode {
	case 0x11: // RL C - Rotate C left through carry
		oldCarry := uint8(0)
		if cpu.Reg.GetCarry() {
			oldCarry = 1
		}
		cpu.Reg.SetCarry((cpu.Reg.C & 0x80) != 0)
		cpu.Reg.C = (cpu.Reg.C << 1) | oldCarry
		cpu.Reg.SetZero(cpu.Reg.C == 0)
		cpu.Reg.SetSubtract(false)
		cpu.Reg.SetHalfCarry(false)
		return 8
	case 0x37: // SWAP A - Swap upper and lower nibbles of A
		cpu.Reg.A = (cpu.Reg.A << 4) | (cpu.Reg.A >> 4)
		cpu.Reg.SetZero(cpu.Reg.A == 0)
		cpu.Reg.SetSubtract(false)
		cpu.Reg.SetHalfCarry(false)
		cpu.Reg.SetCarry(false)
		return 8
	case 0x7C: // BIT 7,H - Test bit 7 of H
		cpu.Reg.SetZero((cpu.Reg.H & 0x80) == 0)
		cpu.Reg.SetSubtract(false)
		cpu.Reg.SetHalfCarry(true)
		return 8
	case 0x87: // RES (reset) 0,A - Reset bit 0 of A
		cpu.Reg.A &= ^uint8(0x01)
		return 8
	default:
		panic(fmt.Sprintf("Unknown CB Opcode: 0xCB%02X at PC: 0x%04X", opcode, cpu.Reg.PC-1))
	}
}
