package main

type MMU struct {
	rom             []byte     // Load the file into here
	vram            [8192]byte // 8KB
	eram            [8192]byte // 8KB
	wram            [8192]byte // 8KB
	oam             [160]byte  // Sprite memory
	io              [128]byte  // I/O ports
	hram            [127]byte  // High RAM
	ie              byte       // Interrupt Enable (just 1 byte)
	scanlineCounter int        // Track cycles for LY register
	ppu             *PPU       // Reference to PPU for rendering
}

func newMMU(rom []byte) *MMU {
	return &MMU{
		rom: rom,
	}
}

func (m *MMU) SetPPU(ppu *PPU) {
	m.ppu = ppu
}

func (m *MMU) UpdateScanline(cycles int) {
	m.scanlineCounter += cycles
	if m.scanlineCounter >= 456 { // 456 cycles per scanline
		m.scanlineCounter -= 456
		currentLY := m.io[0x44] // LY is at offset 0x44

		// Render the current scanline before incrementing
		if m.ppu != nil && currentLY < 144 {
			m.ppu.RenderScanline(currentLY)
		}

		currentLY++
		if currentLY == 144 {
			// Entering VBlank - swap buffers so drawing sees a complete frame
			if m.ppu != nil {
				m.ppu.SwapBuffers()
			}
			m.io[0x0F] |= 0x01 // Set VBLANK interrupt flag
		}
		if currentLY > 153 {
			currentLY = 0
		}
		m.io[0x44] = currentLY
	}
}

func (m *MMU) Read(addr uint16) byte {
	switch {
	case addr < 0x8000: // ROM
		return m.rom[addr]
	case addr < 0xA000: // VRAM
		return m.vram[addr-0x8000]
	case addr < 0xC000: // ERAM
		return 0
	case addr < 0xE000: //WRAM
		return m.wram[addr-0xC000]
	case addr < 0xFE00: // Echo ram
		return 0
	case addr < 0xFEA0: // oam
		return m.oam[addr-0xFE00]
	case addr < 0xFF00: // not usable
		return 0
	case addr < 0xFF80: // IO
		if addr == 0xFF44 { // LY register - current scanline
			// Return the actual LY value from io array
			return m.io[0x44]
		}
		return m.io[addr-0xFF00]
	case addr < 0xFFFF: // HRAM
		return m.hram[addr-0xFF80]
	case addr == 0xFFFF: // IE
		return m.ie
	default:
		return 0
	}
}

func (m *MMU) Write(addr uint16, b byte) {
	switch {
	case addr < 0x8000:
		return
	case addr < 0xA000: // VRAM
		m.vram[addr-0x8000] = b
	case addr < 0xC000: // ERAM
		m.eram[addr-0xA000] = b
	case addr < 0xE000: //WRAM
		m.wram[addr-0xC000] = b
	case addr < 0xFE00: // Echo ram
		return
	case addr < 0xFEA0: // oam
		m.oam[addr-0xFE00] = b
	case addr < 0xFF00: // not usable
		return
	case addr < 0xFF80: // IO
		if addr == 0xFF44 { // LY (scanline)
			return // LY is read-only, writes are ignored
		}
		m.io[addr-0xFF00] = b
	case addr < 0xFFFF: // HRAM
		m.hram[addr-0xFF80] = b
	case addr == 0xFFFF: // IE
		m.ie = b
	}
}
