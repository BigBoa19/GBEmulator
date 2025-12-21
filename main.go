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

	mmu := newMMU(rom)
	ppu := newPPU(mmu)
	mmu.SetPPU(ppu) // Connect MMU to PPU for scanline rendering
	cpu := newCPU(mmu)
	game := newGame(ppu, cpu)
	defer game.cleanup()

	game.run()
	fmt.Println("Gameboy emulator stopping...")
}
