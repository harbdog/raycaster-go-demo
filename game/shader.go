package game

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	fsr0Src = []byte(
		`
//kage:unit pixels
package main
/* EASU stage
*
* This takes a reduced resolution source, and scales it up while preserving detail.
*
* Updates:
*   stretch definition fixed. Thanks nehon for the bug report!
 */
func FsrEasuCF(p vec2) vec3 {
	origin, size := imageSrcRegionOnTexture()
	return imageSrc0UnsafeAt(p*size + origin).rgb
}
/**** EASU ****/
func FsrEasuCon() (vec4, vec4, vec4, vec4) {
	srcSize := imageSrc0Size()
	dstSize := imageDstSize()
	// Output integer position to a pixel position in viewport.
	con0 := vec4(
		srcSize.x/dstSize.x,
		srcSize.y/dstSize.y,
		.5*srcSize.x/dstSize.x-.5,
		.5*srcSize.y/dstSize.y-.5,
	)
	// Viewport pixel position to normalized image space.
	// This is used to get upper-left of 'F' tap.
	con1 := vec4(1, 1, 1, -1) / srcSize.xyxy
	// Centers of gather4, first offset from upper-left of 'F'.
	//      +---+---+
	//      |   |   |
	//      +--(0)--+
	//      | b | c |
	//  +---F---+---+---+
	//  | e | f | g | h |
	//  +--(1)--+--(2)--+
	//  | i | j | k | l |
	//  +---+---+---+---+
	//      | n | o |
	//      +--(3)--+
	//      |   |   |
	//      +---+---+
	// These are from (0) instead of 'F'.
	con2 := vec4(-1, 2, 1, 2) / srcSize.xyxy
	con3 := vec4(0, 4, 0, 0) / srcSize.xyxy
	return con0, con1, con2, con3
}
// Filtering for a given tap for the scalar.
func FsrEasuTapF(aC vec3, aW float, off, dir, len vec2, lob, clp float, c vec3) (vec3, float) {
	// Tap color.
	// Rotate offset by direction.
	v := vec2(dot(off, dir), dot(off, vec2(-dir.y, dir.x)))
	// Anisotropy.
	v *= len
	// Compute distance^2.
	d2 := min(dot(v, v), clp)
	// Limit to the window as at corner, 2 taps can easily be outside.
	// Approximation of lancos2 without sin() or rcp(), or sqrt() to get x.
	//  (25/16 * (2/5 * x^2 - 1)^2 - (25/16 - 1)) * (1/4 * x^2 - 1)^2
	//  |_______________________________________|   |_______________|
	//                   base                             window
	// The general form of the 'base' is,
	//  (a*(b*x^2-1)^2-(a-1))
	// Where 'a=1/(2*b-b^2)' and 'b' moves around the negative lobe.
	wB := .4*d2 - 1.
	wA := lob*d2 - 1.
	wB *= wB
	wA *= wA
	wB = 1.5625*wB - .5625
	w := wB * wA
	// Do weighted average.
	aC += c * w
	aW += w
	return aC, aW
}
// ------------------------------------------------------------------------------------------------------------------------------
// Accumulate direction and length.
func FsrEasuSetF(dir vec2, len, w, lA, lB, lC, lD, lE float) (vec2, float) {
	// Direction is the '+' diff.
	//    a
	//  b c d
	//    e
	// Then takes magnitude from abs average of both sides of 'c'.
	// Length converts gradient reversal to 0, smoothly to non-reversal at 1, shaped, then adding horz and vert terms.
	lenX := max(abs(lD-lC), abs(lC-lB))
	dirX := lD - lB
	dir.x += dirX * w
	lenX = clamp(abs(dirX)/lenX, 0., 1.)
	lenX *= lenX
	len += lenX * w
	// Repeat for the y axis.
	lenY := max(abs(lE-lC), abs(lC-lA))
	dirY := lE - lA
	dir.y += dirY * w
	lenY = clamp(abs(dirY)/lenY, 0., 1.)
	lenY *= lenY
	len += lenY * w
	return dir, len
}
// ------------------------------------------------------------------------------------------------------------------------------
func FsrEasuF(ip vec2, con0, con1, con2, con3 vec4) vec3 {
	//------------------------------------------------------------------------------------------------------------------------------
	// Get position of 'f'.
	pp := ip*con0.xy + con0.zw // Corresponding input pixel/subpixel
	fp := floor(pp)            // fp = source nearest pixel
	pp -= fp                   // pp = source subpixel
	//------------------------------------------------------------------------------------------------------------------------------
	// 12-tap kernel.
	//    b c
	//  e f g h
	//  i j k l
	//    n o
	// Gather 4 ordering.
	//  a b
	//  r g
	p0 := fp*con1.xy + con1.zw
	// These are from p0 to avoid pulling two constants on pre-Navi hardware.
	p1 := p0 + con2.xy
	p2 := p0 + con2.zw
	p3 := p0 + con3.xy
	// TextureGather is not available on WebGL2
	off := vec4(-.5, .5, -.5, .5) * con1.xxyy
	// textureGather to texture offsets
	// x=west y=east z=north w=south
	bC := FsrEasuCF(p0 + off.xw)
	bL := bC.g + 0.5*(bC.r+bC.b)
	cC := FsrEasuCF(p0 + off.yw)
	cL := cC.g + 0.5*(cC.r+cC.b)
	iC := FsrEasuCF(p1 + off.xw)
	iL := iC.g + 0.5*(iC.r+iC.b)
	jC := FsrEasuCF(p1 + off.yw)
	jL := jC.g + 0.5*(jC.r+jC.b)
	fC := FsrEasuCF(p1 + off.yz)
	fL := fC.g + 0.5*(fC.r+fC.b)
	eC := FsrEasuCF(p1 + off.xz)
	eL := eC.g + 0.5*(eC.r+eC.b)
	kC := FsrEasuCF(p2 + off.xw)
	kL := kC.g + 0.5*(kC.r+kC.b)
	lC := FsrEasuCF(p2 + off.yw)
	lL := lC.g + 0.5*(lC.r+lC.b)
	hC := FsrEasuCF(p2 + off.yz)
	hL := hC.g + 0.5*(hC.r+hC.b)
	gC := FsrEasuCF(p2 + off.xz)
	gL := gC.g + 0.5*(gC.r+gC.b)
	oC := FsrEasuCF(p3 + off.yz)
	oL := oC.g + 0.5*(oC.r+oC.b)
	nC := FsrEasuCF(p3 + off.xz)
	nL := nC.g + 0.5*(nC.r+nC.b)
	//------------------------------------------------------------------------------------------------------------------------------
	// Simplest multi-channel approximate luma possible (luma times 2, in 2 FMA/MAD).
	// Accumulate for bilinear interpolation.
	dir := vec2(0)
	len := 0.
	dir, len = FsrEasuSetF(dir, len, (1.-pp.x)*(1.-pp.y), bL, eL, fL, gL, jL)
	dir, len = FsrEasuSetF(dir, len, pp.x*(1.-pp.y), cL, fL, gL, hL, kL)
	dir, len = FsrEasuSetF(dir, len, (1.-pp.x)*pp.y, fL, iL, jL, kL, nL)
	dir, len = FsrEasuSetF(dir, len, pp.x*pp.y, gL, jL, kL, lL, oL)
	//------------------------------------------------------------------------------------------------------------------------------
	// Normalize with approximation, and cleanup close to zero.
	dir2 := dir * dir
	dirR := dir2.x + dir2.y
	zro := dirR < (1.0 / 32768.0)
	dirR = inversesqrt(dirR)
	//dirR = zro ? 1.0 : dirR
	//dir.x = zro ? 1.0 : dir.x
	if zro {
		dirR = 1.
		dir.x = 1
	}
	dir *= vec2(dirR)
	// Transform from {0 to 2} to {0 to 1} range, and shape with square.
	len = len * 0.5
	len *= len
	// Stretch kernel {1.0 vert|horz, to sqrt(2.0) on diagonal}.
	stretch := dot(dir, dir) / (max(abs(dir.x), abs(dir.y)))
	// Anisotropic length after rotation,
	//  x := 1.0 lerp to 'stretch' on edges
	//  y := 1.0 lerp to 2x on edges
	len2 := vec2(1.+(stretch-1.0)*len, 1.-.5*len)
	// Based on the amount of 'edge',
	// the window shifts from +/-{sqrt(2.0) to slightly beyond 2.0}.
	lob := .5 - .29*len
	// Set distance^2 clipping point to the end of the adjustable window.
	clp := 1. / lob
	//------------------------------------------------------------------------------------------------------------------------------
	// Accumulation mixed with min/max of 4 nearest.
	//    b c
	//  e f g h
	//  i j k l
	//    n o
	min4 := min(min(fC, gC), min(jC, kC))
	max4 := max(max(fC, gC), max(jC, kC))
	// Accumulation.
	aC := vec3(0)
	aW := 0.
	aC, aW = FsrEasuTapF(aC, aW, vec2(0, -1)-pp, dir, len2, lob, clp, bC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(1, -1)-pp, dir, len2, lob, clp, cC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(-1, 1)-pp, dir, len2, lob, clp, iC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(0, 1)-pp, dir, len2, lob, clp, jC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(0, 0)-pp, dir, len2, lob, clp, fC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(-1, 0)-pp, dir, len2, lob, clp, eC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(1, 1)-pp, dir, len2, lob, clp, kC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(2, 1)-pp, dir, len2, lob, clp, lC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(2, 0)-pp, dir, len2, lob, clp, hC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(1, 0)-pp, dir, len2, lob, clp, gC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(1, 2)-pp, dir, len2, lob, clp, oC)
	aC, aW = FsrEasuTapF(aC, aW, vec2(0, 2)-pp, dir, len2, lob, clp, nC)
	//------------------------------------------------------------------------------------------------------------------------------
	// Normalize and dering.
	return min(max4, max(min4, aC/aW))
}

var Scale vec2

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	origin, _ := imageSrcRegionOnTexture()
	uv := src - origin
	con0, con1, con2, con3 := FsrEasuCon()
	c := FsrEasuF(uv*Scale, con0, con1, con2, con3)
	return vec4(c.xyz, 1)
}
`)
	fsr1Src = []byte(`
//kage:unit pixels
package main
	
const FSR_RCAS_LIMIT = (0.25 - (1.0 / 16.0))
	
//#define FSR_RCAS_DENOISE
	
// Input callback prototypes that need to be implemented by calling shader
// vec4 FsrRcasLoadF(vec2 p);
// ------------------------------------------------------------------------------------------------------------------------------
func FsrRcasCon(sharpness float) float {
	// Transform from stops to linear value.
	return exp2(-sharpness)
}
func FsrRcasF(ip vec2, con float) vec3 {
	// Constant generated by RcasSetup().
	// Algorithm uses minimal 3x3 pixel neighborhood.
	//    b
	//  d e f
	//    h
	sp := vec2(ip)
	b := FsrRcasLoadF(sp + vec2(0, -1)).rgb
	d := FsrRcasLoadF(sp + vec2(-1, 0)).rgb
	e := FsrRcasLoadF(sp).rgb
	f := FsrRcasLoadF(sp + vec2(1, 0)).rgb
	h := FsrRcasLoadF(sp + vec2(0, 1)).rgb
	// Luma times 2.
	bL := b.g + .5*(b.b+b.r)
	dL := d.g + .5*(d.b+d.r)
	eL := e.g + .5*(e.b+e.r)
	fL := f.g + .5*(f.b+f.r)
	hL := h.g + .5*(h.b+h.r)
	// Noise detection.
	nz := .25*(bL+dL+fL+hL) - eL
	nz = clamp(
		abs(nz)/(max(max(bL, dL), max(eL, max(fL, hL)))-min(min(bL, dL), min(eL, min(fL, hL)))),
		0., 1.,
	)
	nz = 1. - .5*nz
	// Min and max of ring.
	mn4 := min(b, min(f, h))
	mx4 := max(b, max(f, h))
	// Immediate constants for peak range.
	peakC := vec2(1., -4.)
	// Limiters, these need to be high precision RCPs.
	hitMin := mn4 / (4. * mx4)
	hitMax := (peakC.x - mx4) / (4.*mn4 + peakC.y)
	lobeRGB := max(-hitMin, hitMax)
	lobe := max(
		-FSR_RCAS_LIMIT,
		min(max(lobeRGB.r, max(lobeRGB.g, lobeRGB.b)), 0.),
	) * con
	// Apply noise removal.
	//#ifdef FSR_RCAS_DENOISE
	  lobe *= nz
	  //#endif
	// Resolve, which needs the medium precision rcp approximation to avoid visible tonality changes.
	return (lobe*(b+d+h+f) + e) / (4.*lobe + 1.)
}
func FsrRcasLoadF(p vec2) vec4 {
	origin, _ := imageSrcRegionOnTexture()
	return imageSrc0UnsafeAt(p + origin)
}

var Sharpness float

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	origin, size := imageSrcRegionOnTexture()
	// Set up constants
	//division := 0.5 + .3*sin(iTime*.3)
	con := FsrRcasCon(Sharpness)
	// Perform RCAS pass
	col := FsrRcasF(src-origin, con)
	// Source image
	// Bilinear interpolation
	// Nearest pixel
	//uv1 = (floor(vec2(textureSize(iChannel1,0))*fragCoord/iResolution.xy)+.5)/vec2(textureSize(iChannel1,0));
	_ = size
	//uv1 := (src - origin)
	//col_orig := imageSrc1UnsafeAt(origin + uv1).rgb
	//return vec4(col_orig, 1.)
	// Comparison tool
	/*if (fragCoord.x/iResolution.x > division) col = col_orig;
	  if (abs(fragCoord.x/iResolution.x - division)<.005) col = vec3(0);*/
	return vec4(col, 1)
}
`)

	fsr0Shader *ebiten.Shader
	fsr1Shader *ebiten.Shader
)

var (
	pass0Img *ebiten.Image
	finalImg *ebiten.Image
)

func init() {
	var err error

	fsr0Shader, err = ebiten.NewShader(fsr0Src)
	if err != nil {
		log.Fatal(err)
	}

	fsr1Shader, err = ebiten.NewShader(fsr1Src)
	if err != nil {
		log.Fatal(err)
	}
}

// Courtesy of Zyko!
// - https://gist.github.com/Zyko0/0b9244d6780eeb2337162c6dbdf9b787
func (g *Game) DrawFSR(screen, sourceImg *ebiten.Image) {
	scaleX := float64(screen.Bounds().Dx()) / float64(sourceImg.Bounds().Dx())
	scaleY := float64(screen.Bounds().Dy()) / float64(sourceImg.Bounds().Dy())

	initFsrImages := pass0Img == nil || finalImg == nil
	// check for screen size changes after first time initialization
	if !initFsrImages &&
		(pass0Img.Bounds().Dx() != screen.Bounds().Dx() || pass0Img.Bounds().Dy() != screen.Bounds().Dy()) {

		initFsrImages = true
	}

	if initFsrImages {
		pass0Img = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		finalImg = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	}

	// indices mapping vertices to form a 2D quad
	indices := []uint16{0, 1, 2, 1, 2, 3}
	// quad vertices
	vertices := []ebiten.Vertex{
		{
			DstX: 0,
			DstY: 0,
			SrcX: 0,
			SrcY: 0,
		},
		{
			DstX: float32(pass0Img.Bounds().Dx()),
			DstY: 0,
			SrcX: float32(sourceImg.Bounds().Dx()),
			SrcY: 0,
		},
		{
			DstX: 0,
			DstY: float32(pass0Img.Bounds().Dy()),
			SrcX: 0,
			SrcY: float32(sourceImg.Bounds().Dy()),
		},
		{
			DstX: float32(pass0Img.Bounds().Dx()),
			DstY: float32(pass0Img.Bounds().Dy()),
			SrcX: float32(sourceImg.Bounds().Dx()),
			SrcY: float32(sourceImg.Bounds().Dy()),
		},
	}
	// Pass 0
	pass0Img.Clear()
	pass0Img.DrawTrianglesShader(vertices, indices, fsr0Shader, &ebiten.DrawTrianglesShaderOptions{
		Images: [4]*ebiten.Image{
			sourceImg,
		},
		Uniforms: map[string]interface{}{
			"Scale": []float64{
				scaleX, scaleY,
			},
		},
	})
	// Final pass
	// Redefine vertices for second pass
	vertices = []ebiten.Vertex{
		{
			DstX: 0,
			DstY: 0,
			SrcX: 0,
			SrcY: 0,
		},
		{
			DstX: float32(finalImg.Bounds().Dx()),
			DstY: 0,
			SrcX: float32(pass0Img.Bounds().Dx()),
			SrcY: 0,
		},
		{
			DstX: 0,
			DstY: float32(finalImg.Bounds().Dy()),
			SrcX: 0,
			SrcY: float32(pass0Img.Bounds().Dy()),
		},
		{
			DstX: float32(finalImg.Bounds().Dx()),
			DstY: float32(finalImg.Bounds().Dy()),
			SrcX: float32(pass0Img.Bounds().Dx()),
			SrcY: float32(pass0Img.Bounds().Dy()),
		},
	}
	finalImg.Clear()
	finalImg.DrawTrianglesShader(vertices, indices, fsr1Shader, &ebiten.DrawTrianglesShaderOptions{
		Images: [4]*ebiten.Image{
			pass0Img,
		},
		Uniforms: map[string]interface{}{
			"Sharpness": 0.05,
		},
	})

	screen.DrawImage(finalImg, nil)
}
