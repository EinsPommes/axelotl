package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameState int

const (
	MainMenu GameState = iota
	Playing
	Paused
	GameOver
	Settings
)

type MenuOption int

const (
	StartGame MenuOption = iota
	SettingsMenu
	QuitGame
	ResumeGame
	BackToMenu
	VolumeUp
	VolumeDown
	ToggleFullscreen
)

var (
	currentState   GameState = MainMenu
	selectedOption int       = 0
	menuOptions    []string
	showStats      bool = false

	// Settings
	masterVolume float32 = 0.7
	isFullscreen bool    = false

	// Game Over stats
	finalKillCount int
	survivalTime   int
	finalComboMax  int
)

func GetCurrentState() GameState {
	return currentState
}

func SetGameState(state GameState) {
	currentState = state
	selectedOption = 0

	// Set appropriate menu options for each state
	switch state {
	case MainMenu:
		menuOptions = []string{"Start Game", "Settings", "Quit"}
	case Paused:
		menuOptions = []string{"Resume", "Settings", "Main Menu"}
	case GameOver:
		menuOptions = []string{"Try Again", "Main Menu", "Quit"}
	case Settings:
		menuOptions = []string{"Volume: " + fmt.Sprintf("%.0f%%", masterVolume*100), "Fullscreen: " + getToggleText(isFullscreen), "Back"}
	}
}

func getToggleText(enabled bool) string {
	if enabled {
		return "ON"
	}
	return "OFF"
}

func HandleMenuInput() MenuOption {
	// Navigation
	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
		selectedOption--
		if selectedOption < 0 {
			selectedOption = len(menuOptions) - 1
		}
	}

	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
		selectedOption++
		if selectedOption >= len(menuOptions) {
			selectedOption = 0
		}
	}

	// Selection
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		switch currentState {
		case MainMenu:
			switch selectedOption {
			case 0:
				return StartGame
			case 1:
				return SettingsMenu
			case 2:
				return QuitGame
			}
		case Paused:
			switch selectedOption {
			case 0:
				return ResumeGame
			case 1:
				return SettingsMenu
			case 2:
				return BackToMenu
			}
		case GameOver:
			switch selectedOption {
			case 0:
				return StartGame
			case 1:
				return BackToMenu
			case 2:
				return QuitGame
			}
		case Settings:
			switch selectedOption {
			case 0:
				return VolumeUp
			case 1:
				return ToggleFullscreen
			case 2:
				return BackToMenu
			}
		}
	}

	// Volume adjustment in settings
	if currentState == Settings && selectedOption == 0 {
		if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyA) {
			masterVolume -= 0.1
			if masterVolume < 0 {
				masterVolume = 0
			}
			menuOptions[0] = "Volume: " + fmt.Sprintf("%.0f%%", masterVolume*100)
		}
		if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeyD) {
			masterVolume += 0.1
			if masterVolume > 1 {
				masterVolume = 1
			}
			menuOptions[0] = "Volume: " + fmt.Sprintf("%.0f%%", masterVolume*100)
		}
	}

	return MenuOption(-1)
}

func DrawMenu() {
	screenWidth := float32(600)
	screenHeight := float32(600)

	// Same dark overlay for all menus
	rl.DrawRectangle(0, 0, int32(screenWidth), int32(screenHeight), rl.NewColor(0, 0, 0, 150))
	DrawStandardMenu(screenWidth, screenHeight)
}

func DrawStandardMenu(screenWidth, screenHeight float32) {
	// Simple title
	var title string
	switch currentState {
	case MainMenu:
		title = "AXOLOTL"
	case Paused:
		title = "PAUSED"
	case GameOver:
		title = "GAME OVER"
	case Settings:
		title = "SETTINGS"
	}

	titleWidth := rl.MeasureText(title, 48)
	rl.DrawText(title, int32(screenWidth/2-float32(titleWidth)/2), 150, 48, rl.White)

	// Simple menu options
	startY := float32(250)
	for i, option := range menuOptions {
		y := startY + float32(i)*50

		var color rl.Color
		if i == selectedOption {
			color = rl.Yellow
			rl.DrawText(">", int32(screenWidth/2-120), int32(y), 24, rl.Yellow)
		} else {
			color = rl.White
		}

		optionWidth := rl.MeasureText(option, 24)
		rl.DrawText(option, int32(screenWidth/2-float32(optionWidth)/2), int32(y), 24, color)
	}

	// Game Over stats
	if currentState == GameOver && showStats {
		DrawGameOverStats()
	}

	// Simple controls hint
	rl.DrawText("WASD/Arrows: Navigate  •  Enter: Select", 120, int32(screenHeight-40), 16, rl.Gray)
}

func DrawGameOverStats() {
	statsY := float32(400)

	rl.DrawText("FINAL STATS:", 220, int32(statsY), 20, rl.White)

	killText := fmt.Sprintf("Jellyfish Defeated: %d", finalKillCount)
	rl.DrawText(killText, 180, int32(statsY+30), 18, rl.White)

	timeText := fmt.Sprintf("Survival Time: %d seconds", survivalTime/60)
	rl.DrawText(timeText, 180, int32(statsY+55), 18, rl.White)

	comboText := fmt.Sprintf("Max Combo: %d", finalComboMax)
	rl.DrawText(comboText, 180, int32(statsY+80), 18, rl.White)
}

func SetGameOverStats(kills, time, maxCombo int) {
	finalKillCount = kills
	survivalTime = time
	finalComboMax = maxCombo
	showStats = true
}

func GetMasterVolume() float32 {
	return masterVolume
}

func IsFullscreenEnabled() bool {
	return isFullscreen
}

func SetFullscreen(enabled bool) {
	isFullscreen = enabled
	if enabled {
		menuOptions[1] = "Fullscreen: ON"
	} else {
		menuOptions[1] = "Fullscreen: OFF"
	}
}
