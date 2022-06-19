package game

import (
	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type Menu struct {
	mgr    *renderer.Manager
	active bool
}

func MainMenu() Menu {
	mgr := renderer.New(nil)
	return Menu{
		mgr:    mgr,
		active: false,
	}
}

func (m *Menu) Layout(w, h int) {
	m.mgr.SetDisplaySize(float32(w), float32(h))
}

func (m *Menu) Update() {
	if !m.active {
		return
	}

	m.mgr.Update(1.0 / float32(ebiten.MaxTPS()))

	m.mgr.SetText("Menu")
	m.mgr.BeginFrame()
	{
		imgui.Text("Hello, world!")
	}
	m.mgr.EndFrame()
}

func (m *Menu) Draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.mgr.Draw(screen)
}
