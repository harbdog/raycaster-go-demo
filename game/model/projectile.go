package model

import (
	"image/color"
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Projectile struct {
	*Sprite
	Ricochets    int
	Lifespan     float64
	ImpactEffect Effect
}

func NewProjectile(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	p := &Projectile{
		Sprite:       NewSprite(x, y, scale, img, mapColor, anchor, collisionRadius, collisionHeight),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	// projectiles should not be convergence capable by player focal point
	p.Focusable = false

	// projectiles self illuminate so they do not get dimmed in dark conditions
	p.illumination = 5000

	return p
}

func NewAnimatedProjectile(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	p := &Projectile{
		Sprite:       NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, anchor, collisionRadius, collisionHeight),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	// projectiles should not be convergence capable by player focal point
	p.Focusable = false

	// projectiles self illuminate so they do not get dimmed in dark conditions
	p.illumination = 5000

	return p
}

func (p *Projectile) SpawnEffect(x, y, z, angle, pitch float64) *Effect {
	e := &Effect{}
	s := &Sprite{}
	copier.Copy(e, p.ImpactEffect)
	copier.Copy(s, p.ImpactEffect.Sprite)

	e.Sprite = s
	e.Position = &geom.Vector2{X: x, Y: y}
	e.PositionZ = z
	e.Angle = angle
	e.Pitch = pitch

	// keep track of what spawned it
	e.Parent = p.Parent

	return e
}
