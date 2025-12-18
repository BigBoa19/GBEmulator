package main

import (
	"fmt"
	"os"
)

func loadROM(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Pad ROM to 32KB minimum
	rom := make([]byte, 0x8000)
	copy(rom, data)
	return rom, nil
}

func main() {
	fmt.Println("GameBoy emulator starting...")

	var rom []byte
	var err error

	if len(os.Args) > 1 {
		fmt.Printf("Loading ROM: %s\n", os.Args[1])
		rom, err = loadROM(os.Args[1])
		if err != nil {
			fmt.Printf("Error loading ROM: %v\n", err)
			rom = nil
		}
	}

	if rom == nil {
		fmt.Println("Using test ROM...")
		testRom := []byte{
			0x3E, 0x01, // 0x00: LD A, 1
			0x06, 0x00, // 0x02: LD B, 0
			0x80,       // 0x04: ADD A, B (A = A + B)
			0x47,       // 0x05: LD B, A  (Copy result to B)
			0x18, 0xFC, // 0x06: JR -4    (Jump back to 0x04)
		}

		rom = make([]byte, 0x8000)
		copy(rom, testRom)
	}

	mmu := newMMU(rom)
	ppu := newPPU(mmu)
	cpu := newCPU(mmu)
	game := newGame(ppu, cpu)
	defer game.cleanup()

	// Create a visible test pattern - solid black tile (tile 0)
	for i := 0; i < 16; i++ {
		mmu.vram[i] = 0xFF
	}

	// Create a checkerboard tile (tile 1)
	for i := 0; i < 16; i += 2 {
		mmu.vram[16+i] = 0xAA   // 10101010
		mmu.vram[16+i+1] = 0x55 // 01010101
	}

	// Create a striped tile (tile 2)
	for i := 0; i < 16; i += 2 {
		mmu.vram[32+i] = 0xFF   // All bits set
		mmu.vram[32+i+1] = 0x00 // All bits clear
	}

	// Create a test pattern in tile map
	for y := 0; y < 18; y++ {
		for x := 0; x < 20; x++ {
			var tileIndex uint8
			if y < 6 {
				tileIndex = 0 // Top third: black
			} else if y < 12 {
				tileIndex = 1 // Middle third: checkerboard
			} else {
				tileIndex = 2 // Bottom third: stripes
			}
			mmu.vram[0x1800+y*32+x] = tileIndex
		}
	}

	game.run()
	fmt.Println("Gameboy emulator stopping...")
}
