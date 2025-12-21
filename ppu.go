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
    // Only render visible scanlines (0-143)
    if ly >= 144 {
        return
    }

    // Render this scanline to the back buffer
    tileY := int(ly) / 8
    pixelY := int(ly) % 8

    for tileX := 0; tileX < 20; tileX++ {
        mapIndex := tileY*32 + tileX
        tileIndex := ppu.mmu.vram[0x1800+mapIndex]
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