package game

import (
	"math"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type DemoMenu struct {
	mgr    *renderer.Manager
	active bool

	// held vars that should not get updated in real-time
	newRenderScale float32
	newFovDegrees  float32
}

func mainMenu() DemoMenu {
	mgr := renderer.New(nil)
	return DemoMenu{
		mgr:    mgr,
		active: false,
	}
}

func (g *Game) openMenu() {
	g.menu.active = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	// setup initial values for held vars that should not get updated in real-time
	g.menu.newRenderScale = float32(g.renderScale)
	g.menu.newFovDegrees = float32(g.fovDegrees)
}

func (g *Game) closeMenu() {
	g.menu.active = false
}

func (m *DemoMenu) layout(w, h int) {
	m.mgr.SetDisplaySize(float32(w), float32(h))
}

func (m *DemoMenu) update(g *Game) {
	if !m.active {
		return
	}

	m.mgr.Update(1.0 / float32(ebiten.MaxTPS()))

	m.mgr.BeginFrame()
	imgui.Begin("Demo Menu")
	{
		if imgui.Button("Resume") {
			g.closeMenu()
		}
		imgui.SameLineV(0, 50)
		if imgui.Button("Exit") {
			exit(0)
		}

		if imgui.SliderFloatV("Render Scaling", &m.newRenderScale, 0.25, 1.0, "%.2f", imgui.SliderFlagsNone) {
			// only allow values of 0.25, 0.5, 0.75, 1.0
			scaleMod := math.Mod(float64(m.newRenderScale), 0.25)
			if scaleMod != 0 {
				m.newRenderScale = float32(math.Round(float64(m.newRenderScale)/0.25) * 0.25)
			}
		}
		if imgui.IsItemDeactivatedAfterEdit() {
			g.setRenderScale(float64(m.newRenderScale))
		}

		if imgui.Checkbox("Fullscreen", &g.fullscreen) {
			g.setFullscreen(g.fullscreen)
		}

		if imgui.Checkbox("Use VSync", &g.vsync) {
			g.setVsyncEnabled(g.vsync)
		}

		if imgui.SliderFloatV("FOV", &m.newFovDegrees, 40, 140, "%.0f", imgui.SliderFlagsNone) {
			g.setFovAngle(float64(m.newFovDegrees))
		}
	}
	imgui.End()
	m.mgr.EndFrame()
}

func (m *DemoMenu) draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.mgr.Draw(screen)
}
