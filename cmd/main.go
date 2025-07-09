package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	running = true
	bgColor = rl.NewColor(147, 211, 196, 255)
)

func drawScene() {
	rl.DrawRectangle(0, 0, screenWidth, screenHeight, bgColor)
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "axolotl - a game by joeel56")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)

	// world.InitWorld()
	// player.InitPlayer()
}

func update() {
	running = !rl.WindowShouldClose()
}

func render() {
	rl.BeginDrawing()
	rl.ClearBackground(bgColor)
	drawScene()
	rl.EndDrawing()
}

func quit() {
	rl.CloseWindow()
}

func main() {
	for running {
		update()
		render()
	}

	quit()
}
