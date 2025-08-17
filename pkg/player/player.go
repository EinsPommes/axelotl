package player

import (
	"axelot/pkg/world"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	playerSprite rl.Texture2D
	oldX, oldY   float32

	playerSrc                                     rl.Rectangle
	PlayerDest                                    rl.Rectangle
	playerMoving                                  bool
	playerDir                                     int
	playerUp, playerDown, playerLeft, playerRight bool
	playerFrame                                   int
	PlayerHitBox                                  rl.Rectangle
	playerHitBoxYOffset                           float32 = 3

	frameCount int

	playerSpeed float32 = 1.4

	Cam rl.Camera2D

	healthBarTexture rl.Texture2D
	maxHealth        float32 = 10.0
	currentHealth    float32 = 10.0
	healthBarWidth   float32 = 48
	healthBarHeight  float32 = 96
	healthBarX       float32 = 530
	healthBarY       float32 = 480
	healthbarDir     int     = 5
	healthBarSrc     rl.Rectangle

	lastAttackTime int
	attackCooldown int     = 30
	attackRange    float32 = 40
	isAttacking    bool
	attackDuration int = 15
	attackTimer    int
	attackPressed  bool

	healthRegenTimer    int = 0
	healthRegenInterval int = 120

	slimeKillCount int = 0
)

func InitPlayer() {
	playerSprite = rl.LoadTexture("assets/axolotl/spritesheet.png")
	healthBarTexture = rl.LoadTexture("assets/axolotl/Health_bar.png")

	playerSrc = rl.NewRectangle(0, 0, 32, 32)

	healthBarSrc = rl.NewRectangle(0, 0, 32, 64)

	PlayerDest = rl.NewRectangle(600, 400, 32, 32)
	PlayerHitBox = rl.NewRectangle(0, 0, 10, 10)

	Cam = rl.NewCamera2D(rl.NewVector2(float32(screenWidth/2), float32(screenHeight/2)),
		rl.NewVector2(float32(PlayerDest.X-(PlayerDest.Width/2)), float32(PlayerDest.Y-(PlayerDest.Height/2))), 0, 2)
}

func DrawPlayerTexture() {
	rl.DrawTexturePro(playerSprite, playerSrc, PlayerDest, rl.NewVector2(0, 0), 0, rl.White)
}

func PlayerInput() {
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		playerMoving = true
		playerDir = 0
		playerUp = true
	}

	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		playerMoving = true
		playerDir = 1
		playerDown = true
	}

	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		playerMoving = true
		playerDir = 2
		playerLeft = true
	}

	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		playerMoving = true
		playerDir = 3
		playerRight = true
	}

	if rl.IsKeyPressed(rl.KeyQ) {
		attackPressed = true
	}

	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		playerSpeed = 2
	} else {
		playerSpeed = 1.4
	}
}

func TryAttack(targetPos rl.Vector2, attackFunc func(float32)) bool {
	if attackPressed && frameCount-lastAttackTime >= attackCooldown && !isAttacking {
		playerPos := rl.NewVector2(PlayerDest.X, PlayerDest.Y)
		dist := rl.Vector2Distance(playerPos, targetPos)

		if dist <= attackRange {
			attackFunc(1.2)
			lastAttackTime = frameCount
			isAttacking = true
			attackTimer = attackDuration
			playerDir = 4
			attackPressed = false
			return true
		}
	}
	attackPressed = false
	return false
}

func PlayerMoving() {
	oldX, oldY = PlayerDest.X, PlayerDest.Y
	playerSrc.X = playerSrc.Width * float32(playerFrame)

	if isAttacking {
		attackTimer--
		if attackTimer <= 0 {
			isAttacking = false
		}
	}

	RegenerateHealth()

	if playerMoving {
		if playerUp {
			PlayerDest.Y -= playerSpeed

		}
		if playerDown {
			PlayerDest.Y += playerSpeed
		}
		if playerLeft {
			PlayerDest.X -= playerSpeed
		}
		if playerRight {
			PlayerDest.X += playerSpeed
		}

		if frameCount%8 == 1 {
			playerFrame++
		}
	} else if frameCount%45 == 1 {
		playerFrame++

	}

	frameCount++
	if playerFrame >= 8 {
		playerFrame = 0
	}

	if !playerMoving && playerFrame > 1 {
		playerFrame = 0
	}

	playerSrc.Y = playerSrc.Height * float32(playerDir)
	playerSrc.X = playerSrc.Width * float32(playerFrame)

	PlayerHitBox.X = PlayerDest.X + (PlayerDest.Width / 2) - PlayerHitBox.Width/2
	PlayerHitBox.Y = PlayerDest.Y + (PlayerDest.Height / 2) + playerHitBoxYOffset

	PlayerCollision(world.GroundTiles)

	Cam.Target = rl.NewVector2(float32(PlayerDest.X-(PlayerDest.Width/2)), float32(PlayerDest.Y-(PlayerDest.Height/2)))

	playerMoving = false
	playerUp, playerDown, playerLeft, playerRight = false, false, false, false

}

func RegenerateHealth() {
	healthRegenTimer++

	if healthRegenTimer >= healthRegenInterval {
		if currentHealth < maxHealth {
			currentHealth += 1.0
			if currentHealth > maxHealth {
				currentHealth = maxHealth
			}

			UpdateHealthBar()
		}
		healthRegenTimer = 0
	}
}

func UpdateHealthBar() {
	healthPercentage := currentHealth / maxHealth
	if healthPercentage > 0.8 {
		healthbarDir = 5
	} else if healthPercentage > 0.6 {
		healthbarDir = 4
	} else if healthPercentage > 0.4 {
		healthbarDir = 3
	} else if healthPercentage > 0.2 {
		healthbarDir = 2
	} else if healthPercentage > 0.1 {
		healthbarDir = 1
	} else {
		healthbarDir = 0
	}
}

func PlayerCollision(tiles []world.Tile) {
	var jsonMap = world.WorldMap

	for i := 0; i < len(tiles); i++ {
		if PlayerHitBox.X < float32(tiles[i].X*jsonMap.TileSize+jsonMap.TileSize) &&
			PlayerHitBox.X+PlayerHitBox.Width > float32(tiles[i].X*jsonMap.TileSize) &&
			PlayerHitBox.Y < float32(tiles[i].Y*jsonMap.TileSize+jsonMap.TileSize) &&
			PlayerHitBox.Y+PlayerHitBox.Height > float32(tiles[i].Y*jsonMap.TileSize) {

			PlayerDest.X = oldX
			PlayerDest.Y = oldY
		}
	}
}

func UnloadPlayerTexture() {
	rl.UnloadTexture(playerSprite)
	rl.UnloadTexture(healthBarTexture)
}

func SetPlayerDamageState() {
	playerDir = 5
}

func TakeDamage(damage float32) {
	currentHealth -= damage
	if currentHealth < 0 {
		currentHealth = 0
	}

	UpdateHealthBar()
}

func DrawHealthBar() {
	healthBarSrc.Y = healthBarSrc.Height * float32(healthbarDir)

	healthBarDest := rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, healthBarHeight)

	rl.DrawTexturePro(healthBarTexture, healthBarSrc, healthBarDest, rl.NewVector2(0, 0), 0, rl.White)
}

func GetCurrentHealth() float32 {
	return currentHealth
}

func GetMaxHealth() float32 {
	return maxHealth
}

func IsPlayerDead() bool {
	return currentHealth <= 0
}

func ResetPlayer() {
	currentHealth = maxHealth
	PlayerDest.X = 600
	PlayerDest.Y = 400
	playerDir = 1
	playerFrame = 0
	playerMoving = false
	playerUp, playerDown, playerLeft, playerRight = false, false, false, false
	isAttacking = false
	attackTimer = 0
	attackPressed = false
	frameCount = 0
	lastAttackTime = 0
	healthRegenTimer = 0
	slimeKillCount = 0

	Cam.Target = rl.NewVector2(float32(PlayerDest.X-(PlayerDest.Width/2)), float32(PlayerDest.Y-(PlayerDest.Height/2)))

	UpdateHealthBar()
}

func IncrementKillCount() {
	slimeKillCount++
}

func GetKillCount() int {
	return slimeKillCount
}

func DrawKillCounter() {
	killText := fmt.Sprintf("Jellyfish Killed: %d", slimeKillCount)
	rl.DrawText(killText, 10, 10, 20, rl.White)
}
