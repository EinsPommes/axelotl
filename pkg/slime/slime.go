package slime

import (
	"axelot/pkg/world"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

type Slime struct {
	Sprite       rl.Texture2D
	OldX, OldY   float32
	Src          rl.Rectangle
	Dest         rl.Rectangle
	Dir          int
	Frame        int
	HitBox       rl.Rectangle
	FrameCount   int
	LastAttack   int
	IsAttacking  bool
	AttackTimer  int
	MaxHealth    float32
	Health       float32
	HealthbarDir int
	IsDead       bool
	DeathTimer   int

	// Enhanced AI
	aiState      AIState
	stateTimer   int
	targetX      float32
	targetY      float32
	aggroRange   float32
	patrolRadius float32
	wanderTimer  int
}

var (
	slimeSprite           rl.Texture2D
	slimeHealthBarTexture rl.Texture2D
	slimeHealthBarSrc     rl.Rectangle
	slimes                []Slime

	slimeHitBoxYOffset   float32 = 3
	attackCooldown       int     = 60
	attackRange          float32 = 25
	attackDuration       int     = 20
	slimeHealthBarWidth  float32 = 32
	slimeHealthBarHeight float32 = 8
	slimeHealthBarOffset float32 = 3
	deathDuration        int     = 120

	spawnTimer    int = 0
	spawnInterval int = 300 // 500ms at 60 FPS

	globalFrameCount int
)

type AIState int

const (
	Wandering AIState = iota
	Chasing
	Attacking
	Retreating
	Stunned
)

func InitSlime() {
	slimeSprite = rl.LoadTexture("assets/slime/jellyfish_slime.png")
	slimeHealthBarTexture = rl.LoadTexture("assets/axolotl/Health_Bars_001.png")
	slimeHealthBarSrc = rl.NewRectangle(0, 0, 128, 32)

	rand.Seed(time.Now().UnixNano())

	SpawnSlime()
}

func SpawnSlime() {
	waterTiles := world.WaterTiles

	if len(waterTiles) == 0 {
		return
	}

	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		randomIndex := rand.Intn(len(waterTiles))
		selectedTile := waterTiles[randomIndex]

		x := float32(selectedTile.X * world.WorldMap.TileSize)
		y := float32(selectedTile.Y * world.WorldMap.TileSize)

		if !IsLocationOnGround(x, y) {
			newSlime := Slime{
				Sprite:       slimeSprite,
				Src:          rl.NewRectangle(0, 0, 32, 32),
				Dest:         rl.NewRectangle(x, y, 32, 32),
				Dir:          5,
				Frame:        0,
				HitBox:       rl.NewRectangle(0, 0, 10, 10),
				FrameCount:   0,
				LastAttack:   0,
				IsAttacking:  false,
				AttackTimer:  0,
				MaxHealth:    5.0,
				Health:       5.0,
				HealthbarDir: 0,
				IsDead:       false,
				DeathTimer:   0,

				// AI stuff
				aiState:      Wandering,
				stateTimer:   0,
				targetX:      x,
				targetY:      y,
				aggroRange:   120.0 + rand.Float32()*80.0, // random aggro range
				patrolRadius: 50.0 + rand.Float32()*30.0,
				wanderTimer:  rand.Intn(120) + 60,
			}

			slimes = append(slimes, newSlime)
			return
		}
	}
}

func IsLocationOnGround(x, y float32) bool {
	groundTiles := world.GroundTiles
	tileSize := float32(world.WorldMap.TileSize)

	slimeRect := rl.NewRectangle(x, y, 32, 32)

	for _, tile := range groundTiles {
		tileRect := rl.NewRectangle(
			float32(tile.X)*tileSize,
			float32(tile.Y)*tileSize,
			tileSize,
			tileSize,
		)

		if rl.CheckCollisionRecs(slimeRect, tileRect) {
			return true
		}
	}

	return false
}

func UpdateSlimeSpawning() {
	spawnTimer++
	if spawnTimer >= spawnInterval {
		SpawnSlime()
		spawnTimer = 0
	}
}

func DrawSlimeTexture() {
	for i := range slimes {
		if slimes[i].Health > 0 || slimes[i].IsDead {
			rl.DrawTexturePro(slimes[i].Sprite, slimes[i].Src, slimes[i].Dest, rl.NewVector2(0, 0), 0, rl.White)
			if slimes[i].Health > 0 {
				DrawSlimeHealthBar(i)
			}
		}
	}
}

func SlimeMoving(playerPos rl.Vector2, attackPlayerFunc func()) {
	globalFrameCount++

	for i := range slimes {
		if slimes[i].Health <= 0 && !slimes[i].IsDead {
			continue
		}

		slimes[i].OldX, slimes[i].OldY = slimes[i].Dest.X, slimes[i].Dest.Y
		slimes[i].Src.X = slimes[i].Src.Width * float32(slimes[i].Frame)

		if slimes[i].FrameCount%12 == 1 {
			slimes[i].Frame++
		}

		if slimes[i].IsDead {
			if slimes[i].Frame >= 3 {
				slimes[i].Frame = 0
			}
		} else {
			if slimes[i].Frame >= 5 {
				slimes[i].Frame = 0
			}
		}

		if slimes[i].IsAttacking {
			if slimes[i].Frame >= 6 {
				slimes[i].Frame = 0
			}
		} else {
			if slimes[i].Frame >= 5 {
				slimes[i].Frame = 0
			}
		}

		slimes[i].FrameCount++

		if slimes[i].IsDead {
			slimes[i].Dir = 4
			slimes[i].DeathTimer++
			if slimes[i].DeathTimer >= deathDuration {
				slimes[i].IsDead = false
				slimes[i].DeathTimer = 0
			}
		} else if slimes[i].IsAttacking {
			slimes[i].Dir = 3
		} else {
			slimes[i].Dir = 2
		}

		slimes[i].Src.Y = slimes[i].Src.Height * float32(slimes[i].Dir)

		if !slimes[i].IsDead {
			UpdateSlimeAI(i, playerPos, attackPlayerFunc)
		}

		slimes[i].HitBox.X = slimes[i].Dest.X + (slimes[i].Dest.Width / 2) - slimes[i].HitBox.Width/2
		slimes[i].HitBox.Y = slimes[i].Dest.Y + (slimes[i].Dest.Height / 2) + slimeHitBoxYOffset

		SlimeCollision(i, world.GroundTiles)
	}
}

func SlimeCollision(slimeIndex int, tiles []world.Tile) {
	var jsonMap = world.WorldMap

	for i := 0; i < len(tiles); i++ {
		if slimes[slimeIndex].HitBox.X < float32(tiles[i].X*jsonMap.TileSize+jsonMap.TileSize) &&
			slimes[slimeIndex].HitBox.X+slimes[slimeIndex].HitBox.Width > float32(tiles[i].X*jsonMap.TileSize) &&
			slimes[slimeIndex].HitBox.Y < float32(tiles[i].Y*jsonMap.TileSize+jsonMap.TileSize) &&
			slimes[slimeIndex].HitBox.Y+slimes[slimeIndex].HitBox.Height > float32(tiles[i].Y*jsonMap.TileSize) {

			slimes[slimeIndex].Dest.X = slimes[slimeIndex].OldX
			slimes[slimeIndex].Dest.Y = slimes[slimeIndex].OldY
		}
	}
}

func UnloadSlimeTexture() {
	rl.UnloadTexture(slimeSprite)
	rl.UnloadTexture(slimeHealthBarTexture)
}

func DrawSlimeHealthBar(slimeIndex int) {
	if slimes[slimeIndex].Health <= 0 {
		return
	}

	slimeHealthBarSrc.Y = slimeHealthBarSrc.Height * float32(slimes[slimeIndex].HealthbarDir)

	healthBarX := slimes[slimeIndex].Dest.X + (slimes[slimeIndex].Dest.Width / 2) - (slimeHealthBarWidth / 2)
	healthBarY := slimes[slimeIndex].Dest.Y - slimeHealthBarOffset

	slimeHealthBarDest := rl.NewRectangle(healthBarX, healthBarY, slimeHealthBarWidth, slimeHealthBarHeight)

	rl.DrawTexturePro(slimeHealthBarTexture, slimeHealthBarSrc, slimeHealthBarDest, rl.NewVector2(0, 0), 0, rl.White)
}

func GetSlimePositions() []rl.Vector2 {
	var positions []rl.Vector2
	for i := range slimes {
		if slimes[i].Health > 0 && !slimes[i].IsDead {
			positions = append(positions, rl.NewVector2(slimes[i].Dest.X, slimes[i].Dest.Y))
		}
	}
	return positions
}

func GetSlimePosition() rl.Vector2 {
	for i := range slimes {
		if slimes[i].Health > 0 && !slimes[i].IsDead {
			return rl.NewVector2(slimes[i].Dest.X, slimes[i].Dest.Y)
		}
	}
	return rl.NewVector2(0, 0)
}

func GetSlimePositionByIndex(index int) rl.Vector2 {
	if index < 0 || index >= len(slimes) || slimes[index].Health <= 0 || slimes[index].IsDead {
		return rl.NewVector2(0, 0)
	}
	return rl.NewVector2(slimes[index].Dest.X, slimes[index].Dest.Y)
}

func IsSlimeAlive() bool {
	for i := range slimes {
		if slimes[i].Health > 0 && !slimes[i].IsDead {
			return true
		}
	}
	return false
}

func GetClosestSlimeIndex(playerPos rl.Vector2) int {
	closestIndex := -1
	closestDistance := float32(999999)

	for i := range slimes {
		if slimes[i].Health > 0 && !slimes[i].IsDead {
			distance := rl.Vector2Distance(playerPos, rl.NewVector2(slimes[i].Dest.X, slimes[i].Dest.Y))
			if distance < closestDistance {
				closestDistance = distance
				closestIndex = i
			}
		}
	}

	return closestIndex
}

func DamageSlime(slimeIndex int, damage float32, killCounterFunc func()) {
	if slimeIndex < 0 || slimeIndex >= len(slimes) {
		return
	}

	wasAlive := slimes[slimeIndex].Health > 0

	slimes[slimeIndex].Health -= damage
	if slimes[slimeIndex].Health < 0 {
		slimes[slimeIndex].Health = 0
	}

	if wasAlive && slimes[slimeIndex].Health <= 0 {
		slimes[slimeIndex].IsDead = true
		slimes[slimeIndex].DeathTimer = 0
		killCounterFunc()
	}

	healthPercentage := slimes[slimeIndex].Health / slimes[slimeIndex].MaxHealth
	if healthPercentage > 0.875 {
		slimes[slimeIndex].HealthbarDir = 0
	} else if healthPercentage > 0.75 {
		slimes[slimeIndex].HealthbarDir = 1
	} else if healthPercentage > 0.625 {
		slimes[slimeIndex].HealthbarDir = 2
	} else if healthPercentage > 0.5 {
		slimes[slimeIndex].HealthbarDir = 3
	} else if healthPercentage > 0.375 {
		slimes[slimeIndex].HealthbarDir = 4
	} else if healthPercentage > 0.25 {
		slimes[slimeIndex].HealthbarDir = 5
	} else if healthPercentage > 0.125 {
		slimes[slimeIndex].HealthbarDir = 6
	} else {
		slimes[slimeIndex].HealthbarDir = 7
	}
}

func UpdateSlimeAI(slimeIndex int, playerPos rl.Vector2, attackPlayerFunc func()) {
	slime := &slimes[slimeIndex]
	slimePos := rl.NewVector2(slime.Dest.X, slime.Dest.Y)
	dist := rl.Vector2Distance(slimePos, playerPos)

	slime.stateTimer++

	switch slime.aiState {
	case Wandering:
		// Random movement around spawn area
		slime.wanderTimer--
		if slime.wanderTimer <= 0 {
			// Pick new random target within patrol radius
			radius := rand.Float32() * slime.patrolRadius
			slime.targetX = slime.Dest.X + float32(radius)*float32(rl.GetRandomValue(-1, 1))
			slime.targetY = slime.Dest.Y + float32(radius)*float32(rl.GetRandomValue(-1, 1))
			slime.wanderTimer = rand.Intn(120) + 60
		}

		// Move towards wander target slowly
		if slime.targetX != slime.Dest.X || slime.targetY != slime.Dest.Y {
			dirX := slime.targetX - slime.Dest.X
			dirY := slime.targetY - slime.Dest.Y
			length := rl.Vector2Length(rl.NewVector2(dirX, dirY))
			if length > 2 {
				slime.Dest.X += (dirX / length) * 0.3
				slime.Dest.Y += (dirY / length) * 0.3
			}
		}

		// Switch to chasing if player gets close
		if dist < slime.aggroRange {
			slime.aiState = Chasing
			slime.stateTimer = 0
		}

	case Chasing:
		// Chase player aggressively
		if dist <= attackRange && globalFrameCount-slime.LastAttack >= attackCooldown {
			slime.aiState = Attacking
			slime.stateTimer = 0
		} else if dist < 200 && dist > 5 {
			directionX := playerPos.X - slime.Dest.X
			directionY := playerPos.Y - slime.Dest.Y

			length := rl.Vector2Length(rl.NewVector2(directionX, directionY))
			if length > 0 {
				directionX /= length
				directionY /= length
			}

			// Slightly unpredictable movement
			zigzag := float32(rl.GetRandomValue(-10, 10)) / 100.0
			moveSpeed := float32(0.9) + rand.Float32()*0.4 // variable speed

			slime.Dest.X += (directionX + zigzag) * moveSpeed
			slime.Dest.Y += (directionY + zigzag) * moveSpeed
		} else if dist > slime.aggroRange*1.5 {
			// Lost player, go back to wandering
			slime.aiState = Wandering
			slime.wanderTimer = 30
		}

		// Sometimes retreat if low health
		if slime.Health < slime.MaxHealth*0.3 && rand.Intn(100) < 15 {
			slime.aiState = Retreating
			slime.stateTimer = 0
		}

	case Attacking:
		if !slime.IsAttacking {
			slime.LastAttack = globalFrameCount
			slime.IsAttacking = true
			slime.AttackTimer = attackDuration
		}

		if slime.IsAttacking {
			slime.AttackTimer--
			if slime.AttackTimer <= attackDuration-3 && slime.AttackTimer > attackDuration-6 {
				attackPlayerFunc()
			}
			if slime.AttackTimer <= 0 {
				slime.IsAttacking = false
				// After attacking, either chase more or retreat
				if rand.Intn(100) < 70 {
					slime.aiState = Chasing
				} else {
					slime.aiState = Retreating
				}
				slime.stateTimer = 0
			}
		}

	case Retreating:
		// Move away from player
		if dist < 80 {
			directionX := slime.Dest.X - playerPos.X
			directionY := slime.Dest.Y - playerPos.Y

			length := rl.Vector2Length(rl.NewVector2(directionX, directionY))
			if length > 0 {
				directionX /= length
				directionY /= length
			}

			retreatSpeed := float32(1.2)
			slime.Dest.X += directionX * retreatSpeed
			slime.Dest.Y += directionY * retreatSpeed
		} else {
			// Far enough, go back to wandering
			slime.aiState = Wandering
			slime.wanderTimer = rand.Intn(60) + 30
		}

		// If retreat timer runs out, go back to chasing
		if slime.stateTimer > 180 {
			slime.aiState = Chasing
			slime.stateTimer = 0
		}
	}
}

func ResetSlimes() {
	slimes = []Slime{}
	spawnTimer = 0
	globalFrameCount = 0

	SpawnSlime()
}
