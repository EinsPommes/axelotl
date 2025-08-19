package player

import (
	"axelot/pkg/world"
	"fmt"
	"math"

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

	// Combat system
	lastAttackTime int
	attackCooldown int     = 30
	attackRange    float32 = 40
	isAttacking    bool
	attackDuration int = 15
	attackTimer    int
	attackPressed  bool

	// Weapon damage (moved to global vars)
	comboCount    int = 0
	lastComboTime int = 0
	comboWindow   int = 45 // frames to continue combo

	// Different attack types
	chargeAttackPressed bool
	chargeStartTime     int
	isCharging          bool
	maxChargeTime       int = 60

	dashAttackPressed bool
	isDashing         bool
	dashTimer         int
	dashDuration      int     = 20
	dashSpeed         float32 = 4.0
	dashDirectionX    float32 = 0
	dashDirectionY    float32 = 0

	healthRegenTimer    int = 0
	healthRegenInterval int = 120

	slimeKillCount int = 0

	// Visual effects
	screenShake      float32 = 0
	screenShakeDecay float32 = 0.9
	chargeParticles  []ChargeParticle
	chargeGlow       float32 = 0
)

type ChargeParticle struct {
	x, y    float32
	vx, vy  float32
	life    float32
	maxLife float32
	size    float32
	color   rl.Color
}

// Weapon stats - just use one weapon for now
var (
	weaponDamage     float32 = 1.2
	weaponRange      float32 = 40
	weaponCooldown   int     = 30
	weaponComboBonus float32 = 0.3
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

	// Basic attack - can interrupt charging
	if rl.IsKeyPressed(rl.KeyQ) {
		attackPressed = true
		// Cancel charging if Q is pressed
		if isCharging {
			isCharging = false
			chargeAttackPressed = false
		}
	}

	// Charge attack (hold E)
	if rl.IsKeyPressed(rl.KeyE) && !isCharging && !isAttacking {
		isCharging = true
		chargeStartTime = frameCount
		chargeAttackPressed = false
	}
	if rl.IsKeyReleased(rl.KeyE) && isCharging {
		chargeAttackPressed = true
		// Don't set isCharging to false here - let TryAttack handle it
	}

	// Dash attack (R key)
	if rl.IsKeyPressed(rl.KeyR) {
		dashAttackPressed = true
	}

	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		playerSpeed = 2
	} else {
		playerSpeed = 1.4
	}
}

func TryAttack(targetPos rl.Vector2, attackFunc func(float32)) bool {
	playerPos := rl.NewVector2(PlayerDest.X, PlayerDest.Y)
	dist := rl.Vector2Distance(playerPos, targetPos)

	// Basic attack
	if attackPressed && frameCount-lastAttackTime >= weaponCooldown && !isAttacking {
		if dist <= weaponRange {
			damage := weaponDamage

			// Combo system - more damage if attacking in sequence
			if frameCount-lastComboTime <= comboWindow {
				comboCount++
				damage += weaponComboBonus * float32(comboCount)
			} else {
				comboCount = 1
			}

			lastComboTime = frameCount
			attackFunc(damage)
			lastAttackTime = frameCount
			isAttacking = true
			attackTimer = attackDuration
			playerDir = 4
			attackPressed = false
			return true
		}
	}

	// Charge attack - only trigger when E is released
	if chargeAttackPressed && !isAttacking && isCharging {
		chargeTime := frameCount - chargeStartTime
		if dist <= weaponRange*1.5 {
			// Minimum charge time before it becomes effective
			if chargeTime >= 15 {
				chargeDamage := weaponDamage * (1.0 + float32(chargeTime)/float32(maxChargeTime)*1.5)
				attackFunc(chargeDamage)
			} else {
				// Too quick, just do normal damage
				attackFunc(weaponDamage)
			}

			isCharging = false
			chargeAttackPressed = false
			isAttacking = true
			attackTimer = attackDuration + 5
			playerDir = 4
			lastAttackTime = frameCount

			// Water burst effect on charge release
			SpawnChargeExplosion()
			screenShake = 4.0

			return true
		} else {
			// Out of range, cancel charge
			isCharging = false
			chargeAttackPressed = false
		}
	}

	// Dash attack - water dash towards enemy
	if dashAttackPressed && !isDashing && !isAttacking && !isCharging {
		if dist <= 120 {
			isDashing = true
			dashTimer = dashDuration
			dashAttackPressed = false

			// Calculate dash direction towards target
			dashDirectionX = (targetPos.X - playerPos.X) / dist
			dashDirectionY = (targetPos.Y - playerPos.Y) / dist

			// Spawn water wave effect
			SpawnDashWave()
			screenShake = 2.0

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

	// Handle water dash movement
	if isDashing {
		dashTimer--
		if dashTimer > 0 {
			// Smooth dash movement in calculated direction
			PlayerDest.X += dashDirectionX * dashSpeed
			PlayerDest.Y += dashDirectionY * dashSpeed

			// Spawn water trail particles
			if frameCount%3 == 0 {
				SpawnDashTrail()
			}

			// Set dash animation
			playerDir = 4 // dash animation
		} else {
			isDashing = false
			// Dash impact effect
			SpawnDashImpact()
			screenShake = 3.0
		}
	}

	// Handle charging effects
	if isCharging && !isAttacking {
		// Force player to stand still and face down
		playerDir = 1
		playerFrame = 0
		playerMoving = false

		// Update charge glow effect
		chargeTime := frameCount - chargeStartTime
		chargeGlow = float32(chargeTime) / float32(maxChargeTime)
		if chargeGlow > 1.0 {
			chargeGlow = 1.0
		}

		// Spawn water bubbles around player
		if frameCount%8 == 0 {
			SpawnChargeParticle()
		}

		// Gentle water ripple effect when fully charged
		if chargeTime >= maxChargeTime && frameCount%15 == 0 {
			screenShake = 1.5
		}
	} else {
		chargeGlow = 0
	}

	// Update screen shake
	if screenShake > 0 {
		screenShake *= screenShakeDecay
		if screenShake < 0.1 {
			screenShake = 0
		}
	}

	// Update charge particles
	UpdateChargeParticles()

	RegenerateHealth()

	currentSpeed := playerSpeed
	if isDashing {
		currentSpeed = dashSpeed
	}

	if playerMoving && !isDashing && !isCharging {
		if playerUp {
			PlayerDest.Y -= currentSpeed
		}
		if playerDown {
			PlayerDest.Y += currentSpeed
		}
		if playerLeft {
			PlayerDest.X -= currentSpeed
		}
		if playerRight {
			PlayerDest.X += currentSpeed
		}

		if frameCount%8 == 1 {
			playerFrame++
		}
	} else if frameCount%45 == 1 && !isCharging {
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

	// Reset dash system
	isDashing = false
	dashTimer = 0
	dashAttackPressed = false
	dashDirectionX = 0
	dashDirectionY = 0

	// Reset charge system
	isCharging = false
	chargeAttackPressed = false
	chargeStartTime = 0
	chargeGlow = 0

	// Clear all particles and effects
	chargeParticles = []ChargeParticle{}
	screenShake = 0

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

func SpawnChargeParticle() {
	// Spawn water bubble around player
	distance := float32(rl.GetRandomValue(15, 35))

	particle := ChargeParticle{
		x:       PlayerDest.X + PlayerDest.Width/2 + float32(distance)*float32(rl.GetRandomValue(-1, 1)),
		y:       PlayerDest.Y + PlayerDest.Height/2 + float32(distance)*float32(rl.GetRandomValue(-1, 1)),
		vx:      float32(rl.GetRandomValue(-15, 15)) / 20.0, // Slower, more floaty
		vy:      float32(rl.GetRandomValue(-25, -5)) / 15.0, // Bubbles rise up
		life:    1.2,
		maxLife: 1.2,
		size:    float32(rl.GetRandomValue(3, 8)),
		color:   rl.NewColor(100, uint8(150+rl.GetRandomValue(0, 105)), 255, 255), // Blue water bubbles
	}

	chargeParticles = append(chargeParticles, particle)
}

func SpawnChargeExplosion() {
	// Spawn water splash in all directions
	for i := 0; i < 15; i++ {
		angle := float64(i) * 2.0 * math.Pi / 15.0
		speed := float32(rl.GetRandomValue(30, 80)) / 10.0

		particle := ChargeParticle{
			x:       PlayerDest.X + PlayerDest.Width/2,
			y:       PlayerDest.Y + PlayerDest.Height/2,
			vx:      float32(math.Cos(angle)) * speed,
			vy:      float32(math.Sin(angle)) * speed,
			life:    1.5,
			maxLife: 1.5,
			size:    float32(rl.GetRandomValue(4, 10)),
			color:   rl.NewColor(uint8(50+rl.GetRandomValue(0, 100)), uint8(200+rl.GetRandomValue(0, 55)), 255, 255), // Water splash
		}

		chargeParticles = append(chargeParticles, particle)
	}
}

func SpawnDashWave() {
	// Create expanding water wave with wave-like particles
	for ring := 0; ring < 3; ring++ {
		for i := 0; i < 16; i++ {
			angle := float64(i) * 2.0 * math.Pi / 16.0
			speed := float32(2+ring) * 1.8

			// Add some wave-like irregularity
			waveOffset := math.Sin(angle*4) * 0.3
			finalSpeed := speed * (1.0 + float32(waveOffset))

			particle := ChargeParticle{
				x:       PlayerDest.X + PlayerDest.Width/2,
				y:       PlayerDest.Y + PlayerDest.Height/2,
				vx:      float32(math.Cos(angle)) * finalSpeed,
				vy:      float32(math.Sin(angle)) * finalSpeed,
				life:    2.2 - float32(ring)*0.4,
				maxLife: 2.2 - float32(ring)*0.4,
				size:    float32(3 + ring*3), // Bigger for better wave shapes
				color:   rl.NewColor(80, 180+uint8(ring*20), 255, 180-uint8(ring*40)),
			}

			chargeParticles = append(chargeParticles, particle)
		}
	}
}

func SpawnDashTrail() {
	// Water trail behind dashing player
	for i := 0; i < 3; i++ {
		particle := ChargeParticle{
			x:       PlayerDest.X + PlayerDest.Width/2 + float32(rl.GetRandomValue(-8, 8)),
			y:       PlayerDest.Y + PlayerDest.Height/2 + float32(rl.GetRandomValue(-8, 8)),
			vx:      -dashDirectionX*2.0 + float32(rl.GetRandomValue(-10, 10))/20.0,
			vy:      -dashDirectionY*2.0 + float32(rl.GetRandomValue(-10, 10))/20.0,
			life:    0.8,
			maxLife: 0.8,
			size:    float32(rl.GetRandomValue(3, 7)),
			color:   rl.NewColor(120, 200, 255, 180),
		}

		chargeParticles = append(chargeParticles, particle)
	}
}

func SpawnDashImpact() {
	// Water impact splash when dash ends
	for i := 0; i < 12; i++ {
		angle := float64(i) * 2.0 * math.Pi / 12.0
		speed := float32(rl.GetRandomValue(20, 60)) / 10.0

		particle := ChargeParticle{
			x:       PlayerDest.X + PlayerDest.Width/2,
			y:       PlayerDest.Y + PlayerDest.Height/2,
			vx:      float32(math.Cos(angle)) * speed,
			vy:      float32(math.Sin(angle))*speed - 1.0, // Slight upward bias
			life:    1.2,
			maxLife: 1.2,
			size:    float32(rl.GetRandomValue(5, 12)),
			color:   rl.NewColor(60, 220, 255, 255),
		}

		chargeParticles = append(chargeParticles, particle)
	}
}

func UpdateChargeParticles() {
	// Update existing particles
	for i := len(chargeParticles) - 1; i >= 0; i-- {
		p := &chargeParticles[i]
		p.x += p.vx
		p.y += p.vy
		p.life -= 0.02

		// Fade color
		alpha := uint8(p.life * 255)
		p.color.A = alpha

		// Remove dead particles
		if p.life <= 0 {
			chargeParticles = append(chargeParticles[:i], chargeParticles[i+1:]...)
		}
	}
}

func DrawChargeEffects() {
	// Draw pixel charge aura around player
	if chargeGlow > 0 {
		DrawPixelAura(int32(PlayerDest.X+PlayerDest.Width/2), int32(PlayerDest.Y+PlayerDest.Height/2), chargeGlow)
	}

	// Draw water particles with different shapes
	for _, p := range chargeParticles {
		DrawWaterParticle(p)
	}
}

func DrawWaterParticle(p ChargeParticle) {
	// Pixel-art water particles based on size and velocity
	speed := math.Sqrt(float64(p.vx*p.vx + p.vy*p.vy))

	if speed > 3.0 {
		// Fast moving = pixel droplet
		DrawPixelDroplet(p)
	} else if p.size > 6 {
		// Large = pixel splash
		DrawPixelSplash(p)
	} else {
		// Small = pixel bubble
		DrawPixelBubble(p)
	}
}

func DrawPixelDroplet(p ChargeParticle) {
	// Pixel droplet - hand-drawn pixel pattern
	x, y := int32(p.x), int32(p.y)

	// Main droplet body (3x4 pixel pattern)
	rl.DrawRectangle(x-1, y-1, 3, 2, p.color)
	rl.DrawRectangle(x, y-2, 1, 1, p.color)
	rl.DrawRectangle(x-1, y+1, 3, 1, p.color)

	// Droplet tail (1x2 pixels behind)
	angle := math.Atan2(float64(p.vy), float64(p.vx))
	tailX := x - int32(math.Cos(angle)*4)
	tailY := y - int32(math.Sin(angle)*4)

	fadeColor := p.color
	fadeColor.A = fadeColor.A / 2
	rl.DrawRectangle(tailX, tailY, 1, 2, fadeColor)
}

func DrawPixelSplash(p ChargeParticle) {
	// Pixel splash - scattered pixel pattern
	x, y := int32(p.x), int32(p.y)

	// Main splash body
	rl.DrawRectangle(x-2, y-1, 5, 3, p.color)
	rl.DrawRectangle(x-1, y-2, 3, 1, p.color)
	rl.DrawRectangle(x-1, y+2, 3, 1, p.color)

	// Scattered droplets around splash
	fadeColor := p.color
	fadeColor.A = fadeColor.A / 2

	rl.DrawRectangle(x-4, y, 1, 1, fadeColor)
	rl.DrawRectangle(x+4, y-1, 1, 1, fadeColor)
	rl.DrawRectangle(x, y-4, 1, 1, fadeColor)
	rl.DrawRectangle(x-1, y+4, 1, 1, fadeColor)
}

func DrawPixelBubble(p ChargeParticle) {
	// Pixel bubble - simple but clean
	x, y := int32(p.x), int32(p.y)
	size := int32(p.size)

	if size <= 3 {
		// Small bubble (2x2)
		rl.DrawRectangle(x, y, 2, 2, p.color)
		// Highlight pixel
		highlight := rl.NewColor(255, 255, 255, p.color.A/2)
		rl.DrawRectangle(x, y, 1, 1, highlight)
	} else {
		// Medium bubble (3x3)
		rl.DrawRectangle(x-1, y-1, 3, 3, p.color)
		rl.DrawRectangle(x, y-2, 1, 1, p.color)
		rl.DrawRectangle(x-2, y, 1, 1, p.color)

		// Highlight pixels
		highlight := rl.NewColor(255, 255, 255, p.color.A/2)
		rl.DrawRectangle(x-1, y-1, 1, 1, highlight)
		rl.DrawRectangle(x, y-1, 1, 1, highlight)
	}
}

func DrawPixelAura(centerX, centerY int32, intensity float32) {
	// Pixel-art expanding aura rings

	// Ring 1 (inner) - 5x5 hollow square
	if intensity > 0.3 {
		alpha := uint8(intensity * 150)
		color1 := rl.NewColor(150, 220, 255, alpha)

		// Top and bottom lines
		rl.DrawRectangle(centerX-2, centerY-2, 5, 1, color1)
		rl.DrawRectangle(centerX-2, centerY+2, 5, 1, color1)
		// Left and right lines
		rl.DrawRectangle(centerX-2, centerY-1, 1, 3, color1)
		rl.DrawRectangle(centerX+2, centerY-1, 1, 3, color1)
	}

	// Ring 2 (middle) - 7x7 hollow square
	if intensity > 0.6 {
		alpha := uint8(intensity * 100)
		color2 := rl.NewColor(120, 200, 255, alpha)

		rl.DrawRectangle(centerX-3, centerY-3, 7, 1, color2)
		rl.DrawRectangle(centerX-3, centerY+3, 7, 1, color2)
		rl.DrawRectangle(centerX-3, centerY-2, 1, 5, color2)
		rl.DrawRectangle(centerX+3, centerY-2, 1, 5, color2)
	}

	// Ring 3 (outer) - 9x9 hollow square
	if intensity > 0.9 {
		alpha := uint8(intensity * 80)
		color3 := rl.NewColor(100, 180, 255, alpha)

		rl.DrawRectangle(centerX-4, centerY-4, 9, 1, color3)
		rl.DrawRectangle(centerX-4, centerY+4, 9, 1, color3)
		rl.DrawRectangle(centerX-4, centerY-3, 1, 7, color3)
		rl.DrawRectangle(centerX+4, centerY-3, 1, 7, color3)
	}
}

func GetScreenShakeOffset() rl.Vector2 {
	if screenShake <= 0 {
		return rl.NewVector2(0, 0)
	}

	shakeX := float32(rl.GetRandomValue(int32(-screenShake), int32(screenShake)))
	shakeY := float32(rl.GetRandomValue(int32(-screenShake), int32(screenShake)))

	return rl.NewVector2(shakeX, shakeY)
}

func DrawWeaponHUD() {
	// Combo counter
	if comboCount > 1 {
		comboText := fmt.Sprintf("Combo x%d", comboCount)
		rl.DrawText(comboText, 10, 40, 18, rl.Yellow)
	}

	// Charge indicator
	if isCharging {
		chargeTime := frameCount - chargeStartTime
		chargePercent := float32(chargeTime) / float32(maxChargeTime)
		if chargePercent > 1.0 {
			chargePercent = 1.0
		}

		// Draw charge bar
		barWidth := float32(100)
		barHeight := float32(8)
		barX := float32(10)
		barY := float32(70)

		// Bar background
		rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth), int32(barHeight), rl.DarkGray)

		// Water charge progress - changes color when effective
		var barColor rl.Color
		if chargeTime >= 15 {
			barColor = rl.NewColor(100, 220, 255, 200) // Bright water blue
		} else {
			barColor = rl.NewColor(150, 180, 220, 180) // Light blue building up
		}
		rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth*chargePercent), int32(barHeight), barColor)

		// Text
		if chargeTime >= 15 {
			rl.DrawText("WATER POWER READY!", 10, 85, 12, rl.NewColor(100, 220, 255, 255))
		} else {
			rl.DrawText("Gathering water energy...", 10, 85, 12, rl.NewColor(150, 180, 220, 255))
		}
	}

	// Controls reminder
	rl.DrawText("Controls: Q-Attack, E-Charge, R-Dash", 10, screenHeight-25, 12, rl.Gray)
}
