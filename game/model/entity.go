package model

import (
	"image/color"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Entity struct {
	Position        *geom.Vector2
	PositionZ       float64
	Scale           float64
	Anchor          raycaster.SpriteAnchor
	Angle           float64
	Pitch           float64
	Velocity        float64
	CollisionRadius float64
	CollisionHeight float64
	MapColor        color.RGBA
	Parent          *Entity
}

func (e *Entity) Pos() *geom.Vector2 {
	return e.Position
}

func (e *Entity) PosZ() float64 {
	return e.PositionZ
}
