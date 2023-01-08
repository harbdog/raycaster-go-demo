package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"

	"image/color"
	_ "image/png"

	"github.com/harbdog/raycaster-go-demo/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/spf13/viper"
)

const (
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	menu   DemoMenu
	paused bool

	//--create slicer and declare slices--//
	tex                *TextureHandler
	initRenderFloorTex bool

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64
	fullscreen   bool
	vsync        bool
	fovDegrees   float64
	fovDepth     float64

	//--viewport width / height--//
	width  int
	height int

	player *model.Player

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode      MouseMode
	mouseX, mouseY int

	crosshairs *model.Crosshairs

	// zoom settings
	zoomFovDepth float64

	renderDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        color.NRGBA
	maxLightRGB        color.NRGBA

	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *model.Map
	collisionMap []geom.Line

	sprites     map[*model.Sprite]struct{}
	projectiles map[*model.Projectile]struct{}
	effects     map[*model.Effect]struct{}

	mapWidth, mapHeight int

	showSpriteBoxes bool
	debug           bool
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	fmt.Printf("Initializing Game\n")

	// initialize Game object
	g := new(Game)

	g.initConfig()

	ebiten.SetWindowTitle("Raycaster-Go Demo")

	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(g.fullscreen)
	g.setVsyncEnabled(g.vsync)

	// load map
	g.mapObj = model.NewMap()

	// load texture handler
	g.tex = NewTextureHandler(g.mapObj, 32)
	g.tex.renderFloorTex = g.initRenderFloorTex

	g.collisionMap = g.mapObj.GetCollisionLines(clipDistance)
	worldMap := g.mapObj.Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// load content once when first run
	g.loadContent()

	// create crosshairs and weapon
	g.crosshairs = model.NewCrosshairs(1, 1, 2.0, g.tex.textures[16], 8, 8, 55, 57)

	// init player model
	angleDegrees := 60.0
	g.player = model.NewPlayer(8.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance
	g.player.CollisionHeight = 0.5

	// init the sprites
	g.loadSprites()

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mapObj, g.tex)
	g.camera.SetRenderDistance(g.renderDistance)

	g.camera.SetFloorTexture(getTextureFromFile("floor.png"))
	g.camera.SetSkyTexture(getTextureFromFile("sky.png"))

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.fovDegrees = g.camera.FovAngle() // TODO: store and load from config file
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0

	// set demo non-default lighting settings
	g.lightFalloff = -200
	g.globalIllumination = 500
	g.minLightRGB = color.NRGBA{R: 76, G: 76, B: 76}
	g.maxLightRGB = color.NRGBA{R: 255, G: 255, B: 255}

	g.camera.SetLightFalloff(g.lightFalloff)
	g.camera.SetGlobalIllumination(g.globalIllumination)
	g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)

	// init menu system
	g.menu = mainMenu()

	return g
}

func (g *Game) initConfig() {
	viper.SetConfigName("demo-config")
	viper.SetConfigType("json")

	// setup environment variable with DEMO as prefix (e.g. "export DEMO_SCREEN_VSYNC=false")
	viper.SetEnvPrefix("demo")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.raycaster-go-demo"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)
	viper.SetDefault("showSpriteBoxes", false)
	viper.SetDefault("screen.width", 1024)
	viper.SetDefault("screen.height", 768)
	viper.SetDefault("screen.renderScale", 1.0)
	viper.SetDefault("screen.fullscreen", false)
	viper.SetDefault("screen.vsync", true)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.renderFloor", true)

	err := viper.ReadInConfig()
	if err != nil && g.debug {
		log.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.fullscreen = viper.GetBool("screen.fullscreen")
	g.vsync = viper.GetBool("screen.vsync")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.initRenderFloorTex = viper.GetBool("screen.renderFloor")
	g.showSpriteBoxes = viper.GetBool("showSpriteBoxes")
	g.debug = viper.GetBool("debug")
}

func (g *Game) SaveConfig() error {
	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.raycaster-go-demo"

	userConfig := userConfigPath + "/demo-config.json"
	log.Print("Saving config file ", userConfig)

	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err := viper.WriteConfigAs(userConfig)
	if err != nil {
		log.Print(err)
	}

	return err
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	g.paused = false

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.screenWidth)*g.renderScale), int(float64(g.screenHeight)*g.renderScale)
	g.menu.layout(w, h)
	return int(w), int(h)
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// handle input (when paused making sure only to allow input for closing menu so it can be unpaused)
	g.handleInput()

	if !g.paused {
		// Perform logical updates
		w := g.player.Weapon
		if w != nil {
			w.Update()
		}
		g.updateProjectiles()
		g.updateSprites()

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.update(g)

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	numSprites, numProjectiles, numEffects := len(g.sprites), len(g.projectiles), len(g.effects)
	raycastSprites := make([]raycaster.Sprite, numSprites+numProjectiles+numEffects)
	index := 0
	for sprite := range g.sprites {
		raycastSprites[index] = sprite
		index += 1
	}
	for projectile := range g.projectiles {
		raycastSprites[index] = projectile.Sprite
		index += 1
	}
	for effect := range g.effects {
		raycastSprites[index] = effect.Sprite
		index += 1
	}

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast to screen
	g.camera.Draw(screen)

	// draw equipped weapon
	if g.player.Weapon != nil {
		w := g.player.Weapon
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		weaponScale := w.Scale() * g.renderScale
		op.GeoM.Scale(weaponScale, weaponScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(w.W)*weaponScale/2,
			float64(g.height)-float64(w.H)*weaponScale+1,
		)

		// apply lighting setting
		op.ColorM.Scale(float64(g.maxLightRGB.R)/255, float64(g.maxLightRGB.G)/255, float64(g.maxLightRGB.B)/255, 1)

		screen.DrawImage(w.Texture(), op)
	}

	if g.showSpriteBoxes {
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		for sprite := range g.sprites {
			drawSpriteBox(screen, sprite)
		}

		for sprite := range g.projectiles {
			drawSpriteBox(screen, sprite.Sprite)
		}

		for sprite := range g.effects {
			drawSpriteBox(screen, sprite.Sprite)
		}
	}

	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0*g.renderScale, 5.0*g.renderScale)
		op.GeoM.Translate(0, 50)
		screen.DrawImage(mmImg, op)
	}

	// draw crosshairs
	if g.crosshairs != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		crosshairScale := g.crosshairs.Scale() * g.renderScale
		op.GeoM.Scale(crosshairScale, crosshairScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(g.crosshairs.W)*crosshairScale/2,
			float64(g.height)/2-float64(g.crosshairs.H)*crosshairScale/2,
		)
		screen.DrawImage(g.crosshairs.Texture(), op)

		if g.crosshairs.IsHitIndicatorActive() {
			screen.DrawImage(g.crosshairs.HitIndicator.Texture(), op)
			g.crosshairs.Update()
		}
	}

	// draw menu (if active)
	g.menu.draw(screen)

	// draw FPS/TPS counter debug display
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	ebitenutil.DebugPrint(screen, fps)
}

func drawSpriteBox(screen *ebiten.Image, sprite *model.Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	minX, minY := float64(r.Min.X), float64(r.Min.Y)
	maxX, maxY := float64(r.Max.X), float64(r.Max.Y)

	ebitenutil.DrawLine(screen, minX, minY, minX, maxY, color.RGBA{255, 0, 0, 255})
	ebitenutil.DrawLine(screen, minX, maxY, maxX, maxY, color.RGBA{255, 0, 0, 255})
	ebitenutil.DrawLine(screen, maxX, maxY, maxX, minY, color.RGBA{255, 0, 0, 255})
	ebitenutil.DrawLine(screen, maxX, minY, minX, minY, color.RGBA{255, 0, 0, 255})
}

func (g *Game) setFullscreen(fullscreen bool) {
	g.fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

func (g *Game) setResolution(screenWidth, screenHeight int) {
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.renderScale)
}

func (g *Game) setRenderScale(renderScale float64) {
	g.renderScale = renderScale
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
}

func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	if enableVsync {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	} else {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	}
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	pi2 := geom.Pi2
	if g.player.Angle >= pi2 {
		g.player.Angle = pi2 - g.player.Angle
	} else if g.player.Angle <= -pi2 {
		g.player.Angle = g.player.Angle + pi2
	}

	g.player.Moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	// current raycasting method can only allow up to 22.5 degrees down, 45 degrees up
	g.player.Pitch = geom.Clamp(pSpeed+g.player.Pitch, -math.Pi/8, math.Pi/4)
	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.CameraZ = 0.5
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.PositionZ == 0 && g.player.CameraZ == 0.5
}

func (g *Game) Jump() {
	g.player.CameraZ = 0.9
	g.player.PositionZ = 0.4
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.CameraZ = 0.3
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.CameraZ = 0.1
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) fireWeapon() {
	w := g.player.Weapon
	if w == nil {
		g.player.NextWeapon(false)
		return
	}
	if w.OnCooldown() {
		return
	}

	// set weapon firing for animation to run
	w.Fire()

	// spawning projectile at player position just slightly below player's center point of view
	pX, pY, pZ := g.player.Position.X, g.player.Position.Y, geom.Clamp(g.player.CameraZ-0.1, 0.05, 0.95)
	// pitch, angle based on raycasted point at crosshairs
	var pAngle, pPitch float64
	convergenceDistance := g.camera.GetConvergenceDistance()
	convergencePoint := g.camera.GetConvergencePoint()
	if convergenceDistance <= 0 || convergencePoint == nil {
		pAngle, pPitch = g.player.Angle, g.player.Pitch
	} else {
		convergenceLine3d := &geom3d.Line3d{
			X1: pX, Y1: pY, Z1: pZ,
			X2: convergencePoint.X, Y2: convergencePoint.Y, Z2: convergencePoint.Z,
		}
		pAngle, pPitch = convergenceLine3d.Heading(), convergenceLine3d.Pitch()
	}

	projectile := w.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, g.player.Entity)
	if projectile != nil {
		g.addProjectile(projectile)
	}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	g.camera.SetPosition(g.player.Position.Copy())
	g.camera.SetPositionZ(g.player.CameraZ)
	g.camera.SetHeadingAngle(g.player.Angle)
	g.camera.SetPitchAngle(g.player.Pitch)
}

func (g *Game) updateProjectiles() {
	// Testing animated projectile movement
	for p := range g.projectiles {
		if p.Velocity != 0 {

			trajectory := geom3d.Line3dFromAngle(p.Position.X, p.Position.Y, p.PositionZ, p.Angle, p.Pitch, p.Velocity)

			xCheck := trajectory.X2
			yCheck := trajectory.Y2
			zCheck := trajectory.Z2

			newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
			if isCollision || p.PositionZ <= 0 {
				// for testing purposes, projectiles instantly get deleted when collision occurs
				g.deleteProjectile(p)

				// make a sprite/wall getting hit by projectile cause some visual effect
				if p.ImpactEffect.Sprite != nil {
					if len(collisions) >= 1 {
						// use the first collision point to place effect at
						newPos = collisions[0].collision
					}

					// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
					effect := p.SpawnEffect(newPos.X, newPos.Y, p.PositionZ, p.Angle, p.Pitch)

					g.addEffect(effect)
				}

				for _, collisionEntity := range collisions {
					if collisionEntity.entity == g.player.Entity {
						println("ouch!")
					} else {
						// show crosshair hit effect
						g.crosshairs.ActivateHitIndicator(30)
					}
				}
			} else {
				p.Position = newPos
				p.PositionZ = zCheck
			}
		}
		p.Update(g.player.Position)
	}

	// Testing animated effects (explosions)
	for e := range g.effects {
		e.Update(g.player.Position)
		if e.LoopCounter() >= e.LoopCount {
			g.deleteEffect(e)
		}
	}
}

func (g *Game) updateSprites() {
	// Testing animated sprite movement
	for s := range g.sprites {
		if s.Velocity != 0 {
			vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2
			zCheck := s.PositionZ

			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
			if isCollision {
				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
				s.Angle = randFloat(-math.Pi, math.Pi)
				s.Velocity = randFloat(0.01, 0.03)
			} else {
				s.Position = newPos
			}
		}
		s.Update(g.player.Position)
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	// TODO: any cleanup?
	os.Exit(rc)
}
