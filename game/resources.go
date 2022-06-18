package game

import (
	"image"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go-demo/game/model"
	"github.com/harbdog/raycaster-go/geom"
)

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
	g.tex.textures[5] = getTextureFromFile("ebitengine_splash.png")

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
	walker := model.NewAnimatedSprite(7.5, 6.0, walkerScale, 10, g.tex.textures[19], yellow, 4, 8, texWidth, walkerVoffset, walkerCollisionRadius)
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
