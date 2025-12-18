package main

type PPU struct {
	mmu *MMU
	framebuffer [144][160]uint8
}

func newPPU(mmu *MMU) *PPU {
	return &PPU{
		mmu: mmu,
	}
}

func (ppu *PPU) getTile(tileIndex uint8) [8][8]uint8 {
	byteIndex := 16 * tileIndex
	tileRam := ppu.mmu.vram[byteIndex: byteIndex + 16]

	var tile [8][8]uint8
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			bit := 7 - j
			bitLow := (tileRam[2*i] & 1 << bit) >> bit
			bitHigh := (tileRam[2*i + 1] & 1 << bit) >> bit
			tile[i][j] = bitHigh << 1 | bitLow
		}
	}

	return tile
}

func (ppu *PPU) Render() {
	for tileY := 0; tileY < 18; tileY++ {
		for tileX := 0; tileX < 20; tileX++ {
			mapIndex := tileY * 32 + tileX
			tileIndex := ppu.mmu.vram[0x1800 + mapIndex]
			tile := ppu.getTile(tileIndex)

			for py := 0; py < 8; py++ {
				for px := 0; px < 8; px++ {
					ppu.framebuffer[tileY*8 + py][tileX*8 + px] = tile[py][px]
				}
			}
		}
	}
}