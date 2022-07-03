package game

import (
	"fmt"
	"image/color"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type DemoMenu struct {
	mgr    *renderer.Manager
	active bool

	// held vars that should not get updated in real-time
	newRenderWidth  int32
	newRenderHeight int32
	newRenderScale  float32
	newFovDegrees   float32
	newMinLightRGB  *[3]float32
	newMaxLightRGB  *[3]float32
}

func mainMenu() DemoMenu {
	mgr := renderer.New(nil)
	return DemoMenu{
		mgr:    mgr,
		active: false,
	}
}

func (g *Game) openMenu() {
	g.paused = true
	g.menu.active = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	// setup initial values for held vars that should not get updated in real-time
	g.menu.newRenderWidth = int32(g.screenWidth)
	g.menu.newRenderHeight = int32(g.screenHeight)
	g.menu.newRenderScale = float32(g.renderScale)
	g.menu.newFovDegrees = float32(g.fovDegrees)

	if g.menu.newMinLightRGB == nil {
		// for demo purposes only retaining these values in the menu
		g.menu.newMinLightRGB = &[3]float32{0, 0, 0}
		g.menu.newMaxLightRGB = &[3]float32{1, 1, 1} // represents NRGBA{255, 255, 255}
	}
}

func (g *Game) closeMenu() {
	g.paused = false
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
		// Set resolution by using int input fields and button to set it
		{
			imgui.Text("Resolution:")

			imgui.Indent()
			imgui.Text(" Width")
			imgui.SameLine()
			imgui.InputInt("##renderWidth", &m.newRenderWidth)

			imgui.Text("Height")
			imgui.SameLine()
			imgui.InputInt("##renderHeight", &m.newRenderHeight)

			if imgui.Button("Apply") {
				g.setResolution(int(m.newRenderWidth), int(m.newRenderHeight))
			}

			imgui.Unindent()
		}

		// Render scaling: +/- buttons to constrict values (0.25 <= s <= 1.0 in 0.25 increments only)
		{
			imgui.Text(fmt.Sprintf("Render Scaling: %0.2f", m.newRenderScale))
			imgui.SameLine()

			if imgui.Button("-") {
				m.newRenderScale -= 0.25
				if m.newRenderScale < 0.25 {
					m.newRenderScale = 0.25
				}
				g.setRenderScale(float64(m.newRenderScale))
			}

			imgui.SameLine()
			if imgui.Button("+") {
				m.newRenderScale += 0.25
				if m.newRenderScale > 1.0 {
					m.newRenderScale = 1.0
				}
				g.setRenderScale(float64(m.newRenderScale))
			}
		}

		if imgui.SliderFloatV("FOV", &m.newFovDegrees, 40, 140, "%.0f", imgui.SliderFlagsNone) {
			g.setFovAngle(float64(m.newFovDegrees))
		}

		if imgui.Checkbox("Fullscreen", &g.fullscreen) {
			g.setFullscreen(g.fullscreen)
		}

		if imgui.Checkbox("Use VSync", &g.vsync) {
			g.setVsyncEnabled(g.vsync)
		}

		imgui.Checkbox("Floor Texturing", &g.tex.renderFloorTex)

		imgui.Separator()

		imgui.Text("Lighting:")
		lightColorChanged := false
		if imgui.ColorEdit3("Min Lighting", m.newMinLightRGB) {
			lightColorChanged = true
		}
		if imgui.ColorEdit3("Max Lighting", m.newMaxLightRGB) {
			lightColorChanged = true
		}

		if lightColorChanged {
			// need to handle menu derived value as a fraction of 1/255
			minLightRGB := color.NRGBA{R: byte(m.newMinLightRGB[0] * 255), G: byte(m.newMinLightRGB[1] * 255), B: byte(m.newMinLightRGB[2] * 255)}
			maxLightRGB := color.NRGBA{R: byte(m.newMaxLightRGB[0] * 255), G: byte(m.newMaxLightRGB[1] * 255), B: byte(m.newMaxLightRGB[2] * 255)}
			g.camera.SetLightRGB(minLightRGB, maxLightRGB)
		}

		// Just some extra spacing
		imgui.Dummy(imgui.Vec2{X: 10, Y: 10})
		imgui.Separator()
		{
			if imgui.ButtonV("Resume", imgui.Vec2{X: 100, Y: 25}) {
				g.closeMenu()
			}
			imgui.SameLineV(0, imgui.WindowWidth()-200)
			if imgui.ButtonV("Exit", imgui.Vec2{X: 100, Y: 25}) {
				exit(0)
			}
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
