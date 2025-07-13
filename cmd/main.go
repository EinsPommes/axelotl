package main

import (
	"axelot/pkg/player"
	"axelot/pkg/slime"
	"axelot/pkg/world"

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
	world.DrawWorld()
	player.DrawPlayerTexture()
	slime.DrawSlimeTexture()
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "axolotl - a game by joeel56")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)

	world.InitWorld()

	world.LoadMap("assets/map.json")
	player.InitPlayer()
	slime.InitSlime()
}

func input() {
	if rl.IsKeyDown(rl.KeyF10) {
		display := rl.GetCurrentMonitor()
		if rl.IsWindowFullscreen() {
			rl.SetWindowSize(screenWidth, screenHeight)
		} else {
			rl.SetWindowSize(rl.GetMonitorWidth(display), rl.GetMonitorHeight(display))
		}

		rl.ToggleFullscreen()
	}

	player.PlayerInput()

	if rl.IsKeyPressed(rl.KeyEscape) {
		running = false
	}
}

func update() {
	running = !rl.WindowShouldClose()

	if player.IsPlayerDead() {
		player.ResetPlayer()
		slime.ResetSlimes()
		return
	}

	player.PlayerMoving()

	playerPos := rl.NewVector2(player.PlayerDest.X, player.PlayerDest.Y)
	attackPlayerFunc := func() {
		player.SetPlayerDamageState()
		player.TakeDamage(1.1)
	}

	slime.SlimeMoving(playerPos, attackPlayerFunc)
	slime.UpdateSlimeSpawning()

	if slime.IsSlimeAlive() {
		closestSlimeIndex := slime.GetClosestSlimeIndex(playerPos)
		if closestSlimeIndex >= 0 {
			slimePos := slime.GetSlimePositionByIndex(closestSlimeIndex)
			player.TryAttack(slimePos, func(damage float32) {
				slime.DamageSlime(closestSlimeIndex, damage, player.IncrementKillCount)
			})
		}
	}
}

func render() {
	var cam = player.Cam

	rl.BeginDrawing()
	rl.ClearBackground(bgColor)

	rl.BeginMode2D(cam)

	drawScene()
	rl.EndMode2D()

	player.DrawHealthBar()
	player.DrawKillCounter()

	rl.EndDrawing()
}

func quit() {
	player.UnloadPlayerTexture()
	slime.UnloadSlimeTexture()
	world.UnloadWorldTexture()
	rl.CloseWindow()
}

func main() {
	for running {
		input()
		update()
		render()
	}

	quit()
}
