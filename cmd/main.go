package main

import (
	"axelot/pkg/player"
	"axelot/pkg/slime"
	"axelot/pkg/ui"
	"axelot/pkg/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	running      = true
	bgColor      = rl.NewColor(147, 211, 196, 255)
	gameStarted  = false
	survivalTime = 0
	maxCombo     = 0
)

func drawScene() {
	world.DrawWorld()
	player.DrawChargeEffects() // Draw effects behind player
	player.DrawPlayerTexture()
	slime.DrawSlimeTexture()
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "axolotl - a game by joeel56")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)

	// Initialize UI system
	ui.SetGameState(ui.MainMenu)

	world.InitWorld()
	world.LoadMap("assets/map.json")
	player.InitPlayer()
	slime.InitSlime()
}

func input() {
	currentState := ui.GetCurrentState()

	// Handle menu input
	if currentState != ui.Playing {
		action := ui.HandleMenuInput()
		handleMenuAction(action)
		return
	}

	// In-game input
	if rl.IsKeyPressed(rl.KeyP) {
		ui.SetGameState(ui.Paused)
		return
	}

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
		ui.SetGameState(ui.Paused)
	}
}

func handleMenuAction(action ui.MenuOption) {
	// Ignore no-action
	if action == ui.MenuOption(-1) {
		return
	}

	switch action {
	case ui.StartGame:
		ui.SetGameState(ui.Playing)
		gameStarted = true
		survivalTime = 0
		maxCombo = 0
		player.ResetPlayer()
		slime.ResetSlimes()

	case ui.SettingsMenu:
		ui.SetGameState(ui.Settings)

	case ui.QuitGame:
		running = false

	case ui.ResumeGame:
		ui.SetGameState(ui.Playing)

	case ui.BackToMenu:
		ui.SetGameState(ui.MainMenu)
		gameStarted = false

	case ui.ToggleFullscreen:
		display := rl.GetCurrentMonitor()
		if rl.IsWindowFullscreen() {
			rl.SetWindowSize(screenWidth, screenHeight)
		} else {
			rl.SetWindowSize(rl.GetMonitorWidth(display), rl.GetMonitorHeight(display))
		}
		rl.ToggleFullscreen()
		ui.SetFullscreen(!ui.IsFullscreenEnabled())
	}
}

func update() {
	// Check if window was closed, but don't override menu quit
	if rl.WindowShouldClose() {
		running = false
	}

	// Only update game logic when playing
	if ui.GetCurrentState() != ui.Playing {
		return
	}

	// Track survival time and max combo
	if gameStarted {
		survivalTime++
		// Track max combo
		currentCombo := player.GetCurrentCombo()
		if currentCombo > maxCombo {
			maxCombo = currentCombo
		}
	}

	if player.IsPlayerDead() {
		// Game over - show stats
		ui.SetGameOverStats(player.GetKillCount(), survivalTime, maxCombo)
		ui.SetGameState(ui.GameOver)
		return
	}

	player.PlayerMoving()

	playerPos := rl.NewVector2(player.PlayerDest.X, player.PlayerDest.Y)
	attackPlayerFunc := func() {
		player.SetPlayerDamageState()
		player.TakeDamage(0.7)
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
	rl.BeginDrawing()
	rl.ClearBackground(bgColor)

	currentState := ui.GetCurrentState()

	// Always render game world in background
	var cam = player.Cam

	// Apply screen shake to camera
	shakeOffset := player.GetScreenShakeOffset()
	cam.Target.X += shakeOffset.X
	cam.Target.Y += shakeOffset.Y

	rl.BeginMode2D(cam)
	drawScene()
	rl.EndMode2D()

	// Only show HUD when actually playing
	if currentState == ui.Playing {
		player.DrawHealthBar()
		player.DrawKillCounter()
		player.DrawWeaponHUD()
	}

	// Render menu overlay
	if currentState != ui.Playing {
		ui.DrawMenu()
	}

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
