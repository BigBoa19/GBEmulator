package main

type PPU struct {
    mmu         *MMU
    framebuffer [144][160]uint8
    backBuffer  [144][160]uint8 // Double buffering to prevent tearing
}

func newPPU(mmu *MMU) *PPU {
    return &PPU{
        mmu: mmu,
    }
}

func (ppu *PPU) getTile(tileIndex uint8) [8][8]uint8 {
    // Check LCD control register bit 4 for addressing mode
    lcdControl := ppu.mmu.io[0x40] // LCDC register

    var byteIndex int
    if lcdControl&0x10 != 0 || lcdControl == 0 {
        // Mode 1: 8000 method - unsigned addressing (also default if LCDC not set)
        byteIndex = int(tileIndex) * 16
    } else {
        // Mode 0: 8800 method - signed addressing
        signedIndex := int8(tileIndex)
        byteIndex = 0x1000 + (int(signedIndex) * 16) // Base at 0x9000 in VRAM
    }

    // Bounds check
    if byteIndex < 0 || byteIndex+16 > len(ppu.mmu.vram) {
        // Return blank tile if out of bounds
        return [8][8]uint8{}
    }

    tileRam := ppu.mmu.vram[byteIndex : byteIndex+16]

    var tile [8][8]uint8
    for i := 0; i < 8; i++ {
        for j := 0; j < 8; j++ {
            bit := 7 - j
            bitLow := (tileRam[2*i] >> bit) & 1
            bitHigh := (tileRam[2*i+1] >> bit) & 1
            tile[i][j] = (bitHigh << 1) | bitLow
        }
    }

    return tile
}

func (ppu *PPU) RenderScanline(ly uint8) {
    lcdControl := ppu.mmu.io[0x40]

    // FIX 1: Check LCD Enable (Bit 7). If off, stop rendering.
    // (Note: Your display loop should handle drawing "white" when this bit is 0)
    if lcdControl&0x80 == 0 {
        return
    }

    // Only render visible scanlines (0-143)
    if ly >= 144 {
        return
    }

    // FIX 2: Check Background Tile Map Display Select (Bit 3)
    // 0 = 0x9800 (Offset 0x1800), 1 = 0x9C00 (Offset 0x1C00)
    mapOffset := 0x1800
    if lcdControl&0x08 != 0 {
        mapOffset = 0x1C00
    }

    tileY := int(ly) / 8
    pixelY := int(ly) % 8

    for tileX := 0; tileX < 20; tileX++ {
        mapIndex := tileY*32 + tileX
        
        // Use the calculated mapOffset instead of hardcoded 0x1800
        tileIndex := ppu.mmu.vram[mapOffset+mapIndex]
        tile := ppu.getTile(tileIndex)

        for px := 0; px < 8; px++ {
            screenX := tileX*8 + px
            ppu.backBuffer[ly][screenX] = tile[pixelY][px]
        }
    }
}

func (ppu *PPU) SwapBuffers() {
    // Copy back buffer to front buffer at VBlank
    ppu.framebuffer = ppu.backBuffer
}