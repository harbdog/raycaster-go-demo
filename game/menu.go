package game

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type DemoMenu struct {
	active bool
	ui     *ebitenui.UI
	res    *uiResources
	game   *Game

	// held vars that should not get updated in real-time
	newRenderWidth    int32
	newRenderHeight   int32
	newRenderScale    float32
	newFovDegrees     float32
	newRenderDistance float32

	newGlobalIllumination float32
	newLightFalloff       float32
	newMinLightRGB        [3]float32
	newMaxLightRGB        [3]float32
}

func createMenu(g *Game) *DemoMenu {
	res, err := NewUIResources()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// using empty background container since settings will be in a window
	bg := widget.NewContainer()
	var ui *ebitenui.UI = &ebitenui.UI{
		Container: bg,
	}

	menu := &DemoMenu{
		game:   g,
		ui:     ui,
		res:    res,
		active: false,
	}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			// It is using a GridLayout with a single column
			widget.GridLayoutOpts.Columns(1),
			// It uses the Stretch parameter to define how the rows will be layed out.
			// - a fixed sized header
			// - a content row that stretches to fill all remaining space
			// - a fixed sized footer
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			// Padding defines how much space to put around the outside of the grid.
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			// Spacing defines how much space to put between each column and row
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(res.background),
	)

	// window title
	titleBar := titleBarContainer(menu)

	// settings pages
	settings := settingsContainer(menu)
	root.AddChild(settings)

	// footer
	footer := footerContainer(menu)
	root.AddChild(footer)

	ww, wh := ebiten.WindowSize()
	window := widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(root),
		widget.WindowOpts.TitleBar(titleBar, 30),
		widget.WindowOpts.MinSize(500, 200),
		widget.WindowOpts.MaxSize(ww, wh),
		widget.WindowOpts.Resizeable(),
		widget.WindowOpts.ResizeHandler(func(args *widget.WindowChangedEventArgs) {
			fmt.Println("Resize: ", args.Rect)
		}),
		widget.WindowOpts.Draggable(),
		widget.WindowOpts.MoveHandler(func(args *widget.WindowChangedEventArgs) {
			fmt.Println("Move: ", args.Rect)
		}),
	)

	r := image.Rect(0, 0, 2*ww/3, 2*wh/3)
	r = r.Add(image.Point{ww / 8, wh / 16})
	window.SetLocation(r)
	ui.AddWindow(window)

	return menu
}

func (g *Game) openMenu() {
	g.paused = true
	g.menu.active = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	// setup initial values for held vars that should not get updated in real-time or need type conversion
	g.menu.newRenderWidth = int32(g.screenWidth)
	g.menu.newRenderHeight = int32(g.screenHeight)
	g.menu.newRenderScale = float32(g.renderScale)
	g.menu.newFovDegrees = float32(g.fovDegrees)
	g.menu.newRenderDistance = float32(g.renderDistance)

	g.menu.newLightFalloff = float32(g.lightFalloff)
	g.menu.newGlobalIllumination = float32(g.globalIllumination)

	// for color menu items [1, 1, 1] represents NRGBA{255, 255, 255}
	g.menu.newMinLightRGB = [3]float32{
		float32(g.minLightRGB.R) * 1 / 255, float32(g.minLightRGB.G) * 1 / 255, float32(g.minLightRGB.B) * 1 / 255,
	}
	g.menu.newMaxLightRGB = [3]float32{
		float32(g.maxLightRGB.R) * 1 / 255, float32(g.maxLightRGB.G) * 1 / 255, float32(g.maxLightRGB.B) * 1 / 255,
	}
}

func (g *Game) closeMenu() {
	g.paused = false
	g.menu.active = false
}

func (m *DemoMenu) layout(w, h int) {
	// TODO: react to game window layout size/scale changes
	//m.mgr.SetDisplaySize(float32(w), float32(h))
}

func (m *DemoMenu) update() {
	if !m.active {
		return
	}

	m.ui.Update()
}

func (m *DemoMenu) draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.ui.Draw(screen)
}
