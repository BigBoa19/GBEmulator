package main

import (
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var colorMap = map[uint8]uint32{
	0: 0xFFFFFFFF, // White
	1: 0xFFAAAAAA, // Light gray
	2: 0xFF555555, // Dark gray
	3: 0xFF000000, // Black
}

type Game struct {
	ppu        *PPU
	cpu        *CPU
	cycleCount int

	window      *sdl.Window
	renderer    *sdl.Renderer
	texture     *sdl.Texture
	pixelBuffer []uint32
	running     bool
}

func newGame(ppu *PPU, cpu *CPU) *Game {
	sdl.Init(sdl.INIT_VIDEO)
	window, _ := sdl.CreateWindow(
		"Gameboy Emulator",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		640, 576,
		sdl.WINDOW_SHOWN,
	)

	renderer, _ := sdl.CreateRenderer(window, -1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	texture, _ := renderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_STREAMING,
		160, 144,
	)

	return &Game{
		cpu:         cpu,
		ppu:         ppu,
		window:      window,
		renderer:    renderer,
		texture:     texture,
		pixelBuffer: make([]uint32, 23040),
		running:     true,
	}
}

// Handles Quit Detection
func (g *Game) handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		if _, ok := event.(*sdl.QuitEvent); ok {
			g.running = false
		}
	}
}

func (g *Game) update() {
	for g.cycleCount < 70224 {
		cycles := g.cpu.Step()
		g.cycleCount += cycles
		g.cpu.mmu.UpdateScanline(cycles) // Update LY register and render scanlines

		// Check for interrupts after each instruction
		intCycles := g.cpu.HandleInterrupts()
		g.cycleCount += intCycles
	}

	g.cycleCount -= 70224
}

func (g *Game) draw() {
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			pixelVal := g.ppu.framebuffer[y][x]
			g.pixelBuffer[y*160+x] = colorMap[pixelVal]
		}
	}

	g.texture.Update(nil, unsafe.Pointer(&g.pixelBuffer[0]), 160*4)

	g.renderer.Clear()
	g.renderer.Copy(g.texture, nil, nil)
	g.renderer.Present()
}

func (g *Game) run() {
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for g.running {
		g.handleEvents()
		g.update()
		g.draw()
		<-ticker.C
	}
}

func (g *Game) cleanup() {
	g.texture.Destroy()
	g.renderer.Destroy()
	g.window.Destroy()
	sdl.Quit()
}
