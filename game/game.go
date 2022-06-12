package game

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	_ "image/png"

	"github.com/harbdog/raycaster-go-demo/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
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

type MouseMode int

const (
	MouseModeLook MouseMode = iota
	MouseModeMove
	MouseModeCursor
)

// Game - This is the main type for your game.
type Game struct {
	//--create slicer and declare slices--//
	tex *TextureHandler

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64

	//--viewport width / height--//
	width  int
	height int

	player *model.Player

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode       MouseMode
	mouseModeToggle bool
	mouseX, mouseY  int

	// keep track of certain key presses for debounce protection
	debouncedKeys map[ebiten.Key]int

	crosshairs *model.Crosshairs

	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *model.Map
	collisionMap []geom.Line

	sprites     map[*model.Sprite]struct{}
	projectiles map[*model.Projectile]struct{}
	effects     map[*model.Effect]struct{}

	mapWidth, mapHeight int

	debug bool
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

	ebiten.SetWindowSize(g.screenWidth, g.screenHeight)
	ebiten.SetWindowTitle("Raycaster-Go Demo")

	// set target TPS
	targetTPS := 60
	ebiten.SetMaxTPS(targetTPS)

	// use scale to keep the desired window width and height
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))

	// load map
	g.mapObj = model.NewMap()

	// load texture handler
	g.tex = NewTextureHandler(g.mapObj, 32)

	g.collisionMap = g.mapObj.GetCollisionLines(clipDistance)
	worldMap := g.mapObj.Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// load content once when first run
	g.loadContent()

	// create crosshairs and weapon
	g.crosshairs = model.NewCrosshairs(1, 1, 2.0, g.tex.textures[16], 8, 8, 55, 57)

	// init player model
	angleDegrees := 90.0
	g.player = model.NewPlayer(9.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance

	// init the sprites
	g.loadSprites()

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	g.debouncedKeys = make(map[ebiten.Key]int, 8)

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mapObj, g.tex)
	g.camera.SetFloorTexture(getTextureFromFile("floor.png"))
	g.camera.SetSkyTexture(getTextureFromFile("sky.png"))

	// initialize camera to player position
	g.updatePlayerCamera(true)

	return g
}

func (g *Game) initConfig() {
	viper.SetConfigName("demo-config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("demo")

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.raycaster-go-demo"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)
	viper.SetDefault("screen.width", 1024)
	viper.SetDefault("screen.height", 768)
	viper.SetDefault("screen.renderScale", 0.75)

	err := viper.ReadInConfig()
	if err != nil {
		log.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.renderScale = viper.GetFloat64("screen.renderScale")
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

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {

	// TODO: make resource management better

	// load wall textures
	g.tex.textures[0] = getTextureFromFile("stone.png")
	g.tex.textures[1] = getTextureFromFile("left_bot_house.png")
	g.tex.textures[2] = getTextureFromFile("right_bot_house.png")
	g.tex.textures[3] = getTextureFromFile("left_top_house.png")
	g.tex.textures[4] = getTextureFromFile("right_top_house.png")

	// separating sprites out a bit from wall textures
	g.tex.textures[9] = getSpriteFromFile("tree_09.png")
	g.tex.textures[10] = getSpriteFromFile("tree_10.png")
	g.tex.textures[14] = getSpriteFromFile("tree_14.png")

	// load texture sheets
	g.tex.textures[15] = getSpriteFromFile("sorcerer_sheet.png")
	g.tex.textures[16] = getSpriteFromFile("crosshairs_sheet.png")
	g.tex.textures[17] = getSpriteFromFile("charged_bolt_sheet.png")
	g.tex.textures[18] = getSpriteFromFile("blue_explosion_sheet.png")
	g.tex.textures[19] = getSpriteFromFile("outleader_walking_sheet.png")
	g.tex.textures[20] = getSpriteFromFile("hand_spell.png")
	g.tex.textures[21] = getSpriteFromFile("hand_staff.png")
	g.tex.textures[22] = getSpriteFromFile("red_bolt.png")
	g.tex.textures[23] = getSpriteFromFile("red_explosion_sheet.png")

	// just setting the grass texture apart from the rest since it gets special handling
	if g.debug {
		g.tex.floorTex = getRGBAFromFile("grass_debug.png")
	} else {
		g.tex.floorTex = getRGBAFromFile("grass.png")
	}
}

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := filepath.Join("game", "resources", "textures")
	_, tex, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texWidth, texWidth))
		// convert into RGBA format
		for x := 0; x < texWidth; x++ {
			for y := 0; y < texWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "textures")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "sprites")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, sFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func (g *Game) loadSprites() {
	g.projectiles = make(map[*model.Projectile]struct{}, 1024)
	g.effects = make(map[*model.Effect]struct{}, 1024)
	g.sprites = make(map[*model.Sprite]struct{}, 128)

	// colors for minimap representation
	blueish := color.RGBA{62, 62, 100, 96}
	reddish := color.RGBA{180, 62, 62, 96}
	brown := color.RGBA{47, 40, 30, 196}
	green := color.RGBA{27, 37, 7, 196}
	orange := color.RGBA{69, 30, 5, 196}
	yellow := color.RGBA{255, 200, 0, 196}

	// preload projectile sprites
	chargedBoltCollisionRadius := 20.0 / texWidth
	chargedBoltProjectile := model.NewAnimatedProjectile(
		0, 0, 0.75, 1, g.tex.textures[17], blueish,
		12, 1, texWidth, 32, chargedBoltCollisionRadius,
	)

	redBoltCollisionRadius := 5.0 / texWidth
	redBoltProjectile := model.NewProjectile(
		0, 0, 0.25, g.tex.textures[22], reddish,
		texWidth, 32, redBoltCollisionRadius,
	)

	// preload effect sprites
	blueExplosionEffect := model.NewAnimatedEffect(
		0, 0, 0.75, 3, g.tex.textures[18], 5, 3, texWidth, 32, 1,
	)
	chargedBoltProjectile.ImpactEffect = *blueExplosionEffect

	redExplosionEffect := model.NewAnimatedEffect(
		0, 0, 0.20, 1, g.tex.textures[23], 8, 3, texWidth, -32, 1,
	)
	redBoltProjectile.ImpactEffect = *redExplosionEffect

	// create weapons
	chargedBoltRoF := 2.5      // Rate of Fire (as RoF/second)
	chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	chargedBoltWeapon := model.NewAnimatedWeapon(1, 1, 1.0, 7, g.tex.textures[20], 3, 1, *chargedBoltProjectile, chargedBoltVelocity, chargedBoltRoF)
	g.player.AddWeapon(chargedBoltWeapon)

	staffBoltRoF := 6.0
	staffBoltVelocity := 24.0
	staffBoltWeapon := model.NewAnimatedWeapon(1, 1, 1.0, 7, g.tex.textures[21], 3, 1, *redBoltProjectile, staffBoltVelocity, staffBoltRoF)
	g.player.AddWeapon(staffBoltWeapon)

	// animated single facing sorcerer
	sorcScale := 1.25
	sorcVoffset := -76.0
	sorcCollisionRadius := 25.0 / texWidth
	sorc := model.NewAnimatedSprite(5.5, 8.0, sorcScale, 5, g.tex.textures[15], yellow, 10, 1, texWidth, sorcVoffset, sorcCollisionRadius)
	// give sprite a sample velocity for movement
	sorc.Angle = geom.Radians(270)
	sorc.Velocity = 0.02
	g.addSprite(sorc)

	// animated walking 8-directional leader
	// [walkerTexFacingMap] player facing angle : texture row index
	var walkerTexFacingMap = map[float64]int{
		geom.Radians(315): 0,
		geom.Radians(270): 1,
		geom.Radians(225): 2,
		geom.Radians(180): 3,
		geom.Radians(135): 4,
		geom.Radians(90):  5,
		geom.Radians(45):  6,
		geom.Radians(0):   7,
	}
	walkerScale := 0.75
	walkerVoffset := 76.0
	walkerCollisionRadius := 30.0 / texWidth
	walker := model.NewAnimatedSprite(9.5, 6.0, walkerScale, 10, g.tex.textures[19], yellow, 4, 8, texWidth, walkerVoffset, walkerCollisionRadius)
	walker.SetAnimationReversed(true) // this sprite sheet has reversed animation frame order
	walker.SetTextureFacingMap(walkerTexFacingMap)
	// give sprite a sample velocity for movement
	walker.Angle = geom.Radians(0)
	walker.Velocity = 0.02
	g.addSprite(walker)

	if g.debug {
		// just some debugging stuff
		sorc.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
		walker.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
		chargedBoltProjectile.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	}

	// testing sprite scaling
	testScale := 0.5
	g.addSprite(model.NewSprite(10.5, 2.5, testScale, g.tex.textures[9], green, texWidth, 128, 0))

	// // line of trees for testing in front of initial view
	// Setting CollisionRadius=0 to disable collision against small trees
	g.addSprite(model.NewSprite(19.5, 11.5, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(17.5, 11.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(15.5, 11.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	// // // render a forest!
	g.addSprite(model.NewSprite(11.5, 1.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 1.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(132.5, 1.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 2, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 2, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 2, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 2.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.25, 2.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 2.25, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 3, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 3, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.25, 3, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(10.5, 3.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 3.25, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 3.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.25, 3.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(10.5, 4, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 4, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 4, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 4, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(10.5, 4.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.25, 4.5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 4.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 4.5, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.5, 4.25, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(10.5, 5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 5, 1.0, g.tex.textures[9], green, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.25, 5, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.5, 5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 5.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 5.25, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 5.25, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.5, 5.5, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(15.5, 5.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(11.5, 6, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 6, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.25, 6, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.25, 6, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(15.5, 6, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 6.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 6.25, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.5, 6.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(12.5, 7, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 7, 1.0, g.tex.textures[10], brown, texWidth, 0, 0))
	g.addSprite(model.NewSprite(14.5, 7, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 7.5, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
	g.addSprite(model.NewSprite(13.5, 8, 1.0, g.tex.textures[14], orange, texWidth, 0, 0))
}

func (g *Game) addSprite(sprite *model.Sprite) {
	g.sprites[sprite] = struct{}{}
}

// func (g *Game) deleteSprite(sprite *model.Sprite) {
// 	delete(g.sprites, sprite)
// }

func (g *Game) addProjectile(projectile *model.Projectile) {
	g.projectiles[projectile] = struct{}{}
}

func (g *Game) deleteProjectile(projectile *model.Projectile) {
	delete(g.projectiles, projectile)
}

func (g *Game) addEffect(effect *model.Effect) {
	g.effects[effect] = struct{}{}
}

func (g *Game) deleteEffect(effect *model.Effect) {
	delete(g.effects, effect)
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(float64(g.screenWidth) * g.renderScale), int(float64(g.screenHeight) * g.renderScale)
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

	// Render to screen
	g.camera.Draw(screen)

	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0, 5.0)
		op.GeoM.Translate(0, 50)
		screen.DrawImage(mmImg, op)
	}

	// draw equipped weapon
	if g.player.Weapon != nil {
		w := g.player.Weapon
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		weaponScale := w.Scale()
		op.GeoM.Scale(weaponScale, weaponScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(w.W)*weaponScale/2,
			float64(g.height)-float64(w.H)*weaponScale+1,
		)
		screen.DrawImage(w.Texture(), op)

		w.Update()
	}

	// draw crosshairs
	if g.crosshairs != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		crosshairScale := g.crosshairs.Scale()
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
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Perform logical updates
	g.updateProjectiles()
	g.updateSprites()

	// handle input
	g.handleInput()

	// handle player camera movement
	g.updatePlayerCamera(false)

	return nil
}

func (g *Game) debounceInput(key ebiten.Key, duration int) {
	g.debouncedKeys[key] = duration
}

func (g *Game) updatedDebounces() {
	for key, duration := range g.debouncedKeys {
		duration--
		g.debouncedKeys[key] = duration

		if duration <= 0 {
			delete(g.debouncedKeys, key)
		}
	}
}

func (g *Game) isDebouncedInput(key ebiten.Key) bool {
	if value, ok := g.debouncedKeys[key]; ok {
		if value > 0 {
			return true
		}
	}
	return false
}

func (g *Game) handleInput() {
	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	// update any currently debounced inputs
	g.updatedDebounces()

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyEscape):
		if g.mouseMode != MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			g.mouseMode = MouseModeCursor
			g.mouseModeToggle = true
			g.debounceInput(ebiten.KeyEscape, 5)
		} else if g.isDebouncedInput(ebiten.KeyEscape) {
			// continue to debounce key since it is still being held
			g.debounceInput(ebiten.KeyEscape, 5)
		} else {
			// debounce period over, it has been pressed again after some pause
			g.mouseModeToggle = false
		}

	case ebiten.IsKeyPressed(ebiten.KeyControl):
		if g.mouseMode != MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			g.mouseMode = MouseModeCursor
		}

	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		if g.mouseMode != MouseModeMove {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			g.mouseMode = MouseModeMove
			g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
		}

	case g.mouseMode != MouseModeLook && !g.mouseModeToggle:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.mouseMode = MouseModeLook
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
				g.Move(0.01 * float64(dy) * moveModifier)
			}
		}
	case MouseModeLook:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
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

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, true)
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

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, true)
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
	g.player.Pitch += pSpeed

	// current raycasting method can only allow up to 45 degree pitch in either direction
	g.player.Pitch = geom.Clamp(g.player.Pitch, -math.Pi/4, math.Pi/4)

	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.PositionZ = 0.5
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.PosZ() == 0.5
}

func (g *Game) Jump() {
	g.player.PositionZ = 0.9
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.PositionZ = 0.3
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.PositionZ = 0.1
	g.player.Moved = true
}

type EntityCollision struct {
	entity    *model.Entity
	collision *geom.Vector2
}

// checks for valid move from current position, returns valid (x, y) position, whether a collision
// was encountered, and a list of entity collisions that may have been encountered
func (g *Game) getValidMove(entity *model.Entity, moveX, moveY float64, checkAlternate bool) (*geom.Vector2, bool, []*EntityCollision) {
	newX, newY := moveX, moveY

	posX, posY := entity.Position.X, entity.Position.Y
	if posX == newX && posY == moveY {
		return &geom.Vector2{X: posX, Y: posY}, false, []*EntityCollision{}
	}

	moveLine := geom.Line{X1: posX, Y1: posY, X2: newX, Y2: newY}

	intersectPoints := []geom.Vector2{}
	collisionEntities := []*EntityCollision{}

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := geom.LineIntersection(moveLine, borderLine); ok {
			intersectPoints = append(intersectPoints, geom.Vector2{X: px, Y: py})
		}
	}

	// check sprite against player collision
	if entity != g.player.Entity && entity.Parent != g.player.Entity && entity.CollisionRadius > 0 {
		// TODO: only check for collision if player is somewhat nearby

		// check if movement line intersects with combined collision radii
		combinedCircle := geom.Circle{X: g.player.Position.X, Y: g.player.Position.Y, Radius: g.player.CollisionRadius + entity.CollisionRadius}
		combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

		if len(combinedIntersects) > 0 {
			playerCircle := geom.Circle{X: g.player.Position.X, Y: g.player.Position.Y, Radius: g.player.CollisionRadius}
			for _, chkPoint := range combinedIntersects {
				// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
				chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: g.player.Position.X, Y2: g.player.Position.Y}
				intersectPoints = append(intersectPoints, geom.LineCircleIntersection(chkLine, playerCircle, true)...)

				for _, intersect := range intersectPoints {
					collisionEntities = append(collisionEntities, &EntityCollision{entity: g.player.Entity, collision: &intersect})
				}
			}
		}
	}

	// check sprite collisions
	for sprite := range g.sprites {
		// TODO: only check intersection of nearby sprites instead of all of them
		if entity == sprite.Entity || entity.Parent == sprite.Entity || entity.CollisionRadius <= 0 || sprite.CollisionRadius <= 0 {
			continue
		}

		// check if movement line intersects with combined collision radii
		combinedCircle := geom.Circle{X: sprite.Position.X, Y: sprite.Position.Y, Radius: sprite.CollisionRadius + entity.CollisionRadius}
		combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

		if len(combinedIntersects) > 0 {
			spriteCircle := geom.Circle{X: sprite.Position.X, Y: sprite.Position.Y, Radius: sprite.CollisionRadius}
			for _, chkPoint := range combinedIntersects {
				// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
				chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: sprite.Position.X, Y2: sprite.Position.Y}
				intersectPoints = append(intersectPoints, geom.LineCircleIntersection(chkLine, spriteCircle, true)...)

				for _, intersect := range intersectPoints {
					collisionEntities = append(collisionEntities, &EntityCollision{entity: sprite.Entity, collision: &intersect})
				}
			}
		}
	}

	// sort collisions by distance to current entity position
	sort.Slice(collisionEntities, func(i, j int) bool {
		distI := geom.Distance2(posX, posY, collisionEntities[i].collision.X, collisionEntities[i].collision.Y)
		distJ := geom.Distance2(posX, posY, collisionEntities[j].collision.X, collisionEntities[j].collision.Y)
		return distI < distJ
	})

	isCollision := len(intersectPoints) > 0

	if isCollision {
		if checkAlternate {
			// find the point closest to the start position
			min := math.Inf(1)
			minI := -1
			for i, p := range intersectPoints {
				d2 := geom.Distance2(posX, posY, p.X, p.Y)
				if d2 < min {
					min = d2
					minI = i
				}
			}

			// use the closest intersecting point to determine a safe distance to make the move
			moveLine = geom.Line{X1: posX, Y1: posY, X2: intersectPoints[minI].X, Y2: intersectPoints[minI].Y}
			dist := math.Sqrt(min)
			angle := moveLine.Angle()

			// generate new move line using calculated angle and safe distance from intersecting point
			moveLine = geom.LineFromAngle(posX, posY, angle, dist-0.01)

			newX, newY = moveLine.X2, moveLine.Y2

			// if either X or Y direction was already intersecting, attempt move only in the adjacent direction
			xDiff := math.Abs(newX - posX)
			yDiff := math.Abs(newY - posY)
			if xDiff > 0.001 || yDiff > 0.001 {
				switch {
				case xDiff <= 0.001:
					// no more room to move in X, try to move only Y
					// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
					// 	c.pos.X, c.pos.Y, moveX, moveY, posX, moveY)
					return g.getValidMove(entity, posX, moveY, false)
				case yDiff <= 0.001:
					// no more room to move in Y, try to move only X
					// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
					// 	c.pos.X, c.pos.Y, moveX, moveY, moveX, posY)
					return g.getValidMove(entity, moveX, posY, false)
				default:
					// try the new position
					// TODO: need some way to try a potentially valid shorter move without checkAlternate while also avoiding infinite loop
					return g.getValidMove(entity, newX, newY, false)
				}
			} else {
				// looks like it cannot move
				return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
			}
		} else {
			// looks like it cannot move
			return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
		}
	}

	// prevent index out of bounds errors
	ix := int(newX)
	iy := int(newY)

	switch {
	case ix < 0 || newX < 0:
		newX = clipDistance
		ix = 0
	case ix >= g.mapWidth:
		newX = float64(g.mapWidth) - clipDistance
		ix = int(newX)
	}

	switch {
	case iy < 0 || newY < 0:
		newY = clipDistance
		iy = 0
	case iy >= g.mapHeight:
		newY = float64(g.mapHeight) - clipDistance
		iy = int(newY)
	}

	worldMap := g.mapObj.Level(0)
	if worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	} else {
		isCollision = true
	}

	return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
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
	pX, pY, pZ := g.player.Position.X, g.player.Position.Y, geom.Clamp(g.player.PositionZ-0.15, 0.05, g.player.PositionZ+0.5)
	// TODO: pitch angle should be based on raycasted angle toward crosshairs, for now just simplified as player pitch angle
	pAngle, pPitch := g.player.Angle, g.player.Pitch

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

	playerPos := g.player.Position.Copy()
	playerPosZ := (g.player.PositionZ - 0.5) * float64(g.height)

	g.camera.SetPosition(playerPos)
	g.camera.SetPositionZ(playerPosZ)
	g.camera.SetHeadingAngle(g.player.Angle)
	g.camera.SetPitchAngle(g.player.Pitch)
}

func (g *Game) updateProjectiles() {
	// Testing animated projectile movement
	for p := range g.projectiles {
		if p.Velocity != 0 {

			realVelocity := p.Velocity
			zVelocity := 0.0
			if p.Pitch != 0 {
				// would be better to use proper 3D geometry math here, but trying to avoid matrix math library
				// for this one simple use (but if becomes desired: https://github.com/ungerik/go3d)
				realVelocity = geom.GetAdjacentHypotenuseTriangleLeg(p.Pitch, p.Velocity)
				zVelocity = geom.LineFromAngle(0, 0, p.Pitch, realVelocity).Y2
			}

			vLine := geom.LineFromAngle(p.Position.X, p.Position.Y, p.Angle, realVelocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2

			// TODO: getValidMove needs to be able to take PosZ into account for wall/sprite collisions
			newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, false)
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

				if zVelocity != 0 {
					p.PositionZ += zVelocity
				}
			}
		}
		p.Update(g.player.Position)
	}

	// Testing animated effects (explosions)
	for e := range g.effects {
		e.Update(g.player.Position)
		if e.GetLoopCounter() >= e.LoopCount {
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

			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, false)
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
