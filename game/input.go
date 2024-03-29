package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseMode int

const (
	MouseModeLook MouseMode = iota
	MouseModeMove
	MouseModeCursor
)

func (g *Game) handleInput() {

	menuKeyPressed := inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF1)
	if menuKeyPressed {
		if g.menu.active {
			if g.osType == osTypeBrowser && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				// do not allow Esc key close menu in browser, since Esc key releases browser mouse capture
			} else {
				g.closeMenu()
			}
		} else {
			g.openMenu()
		}
	}

	if g.paused {
		// currently only paused when menu is active, one could consider other pauses not the subject of this demo
		return
	}

	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && g.osType == osTypeDesktop {
		// debug cursor mode not intended for browser purposes
		if g.mouseMode != MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			g.mouseMode = MouseModeCursor
		}
	} else if inpututil.IsKeyJustReleased(ebiten.KeyControl) {
		if g.mouseMode == MouseModeCursor {
			g.mouseMode = MouseModeLook
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		if g.mouseMode != MouseModeMove {
			g.mouseMode = MouseModeMove
		}
	} else if inpututil.IsKeyJustReleased(ebiten.KeyAlt) {
		if g.mouseMode == MouseModeMove {
			g.mouseMode = MouseModeLook
		}
	}

	if (g.mouseMode == MouseModeLook || g.mouseMode == MouseModeMove) && ebiten.CursorMode() != ebiten.CursorModeCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)

		// reset initial mouse capture position
		g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	}

	switch g.mouseMode {
	case MouseModeCursor:
		g.mouseX, g.mouseY = ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			fmt.Printf("mouse left clicked: (%v, %v)\n", g.mouseX, g.mouseY)
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", g.mouseX, g.mouseY)
		}

	case MouseModeMove:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		isStrafeMove := false
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click in this mode to strafe instead of rotate with mouse X axis
			isStrafeMove = true
		}

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// set initial mouse capture position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				if isStrafeMove {
					g.Strafe(-0.01 * float64(dx) * moveModifier)
				} else {
					g.Rotate(0.005 * float64(dx) * moveModifier)
				}
			}

			if dy != 0 {
				g.Move(0.01 * float64(dy) * moveModifier)
			}
		}
	case MouseModeLook:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click to zoom view in this mode
			if g.camera.FovDepth() != g.zoomFovDepth {
				zoomFovDegrees := g.fovDegrees / g.zoomFovDepth
				g.camera.SetFovAngle(zoomFovDegrees, g.zoomFovDepth)
				g.camera.SetPitchAngle(g.player.Pitch)
			}
		} else if g.camera.FovDepth() == g.zoomFovDepth {
			// unzoom
			g.camera.SetFovAngle(g.fovDegrees, 1.0)
			g.camera.SetPitchAngle(g.player.Pitch)
		}

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				g.Rotate(0.005 * float64(dx) * moveModifier)
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy))
			}
		}
	}

	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		g.player.NextWeapon(wheelY > 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit1) {
		g.player.SelectWeapon(0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit2) {
		g.player.SelectWeapon(1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		// put away/holster weapon
		g.player.SelectWeapon(-1)
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rotLeft = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		rotRight = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		forward = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		backward = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.Crouch()
	} else if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.Prone()
	} else if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Jump()
	} else if !g.IsStanding() {
		g.Stand()
	}

	if forward {
		g.Move(0.06 * moveModifier)
	} else if backward {
		g.Move(-0.06 * moveModifier)
	}

	if g.mouseMode == MouseModeLook || g.mouseMode == MouseModeMove {
		// strafe instead of rotate
		if rotLeft {
			g.Strafe(-0.05 * moveModifier)
		} else if rotRight {
			g.Strafe(0.05 * moveModifier)
		}
	} else {
		if rotLeft {
			g.Rotate(0.03 * moveModifier)
		} else if rotRight {
			g.Rotate(-0.03 * moveModifier)
		}
	}
}
