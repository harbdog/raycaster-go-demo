package game

import (
	"image/color"
	"strconv"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

const (
	backgroundColor = "131a22"

	textIdleColor     = "dff4ff"
	textDisabledColor = "5a7a91"

	labelIdleColor     = textIdleColor
	labelDisabledColor = textDisabledColor

	buttonIdleColor     = textIdleColor
	buttonDisabledColor = labelDisabledColor

	listSelectedBackground         = "4b687a"
	listDisabledSelectedBackground = "2a3944"

	listFocusedBackground = "2a3944"

	headerColor = textIdleColor

	textInputCaretColor         = "e7c34b"
	textInputDisabledCaretColor = "766326"

	toolTipColor = backgroundColor

	separatorColor = listDisabledSelectedBackground
)

const (
	fontFaceRegular = "resources/menu/fonts/NotoSans-Regular.ttf"
	fontFaceBold    = "resources/menu/fonts/NotoSans-Bold.ttf"
)

type uiResources struct {
	fonts          *fonts
	background     *image.NineSlice
	separatorColor color.Color

	text        *textResources
	button      *buttonResources
	label       *labelResources
	checkbox    *checkboxResources
	comboButton *comboButtonResources
	list        *listResources
	slider      *sliderResources
	progressBar *progressBarResources
	panel       *panelResources
	tabBook     *tabBookResources
	header      *headerResources
	textInput   *textInputResources
	textArea    *textAreaResources
	toolTip     *toolTipResources
}

type textResources struct {
	idleColor     color.Color
	disabledColor color.Color
	face          font.Face
	titleFace     font.Face
	bigTitleFace  font.Face
	smallFace     font.Face
}

type buttonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    font.Face
	padding widget.Insets
}

type checkboxResources struct {
	image   *widget.ButtonImage
	graphic *widget.CheckboxGraphicImage
	spacing int
}

type labelResources struct {
	text *widget.LabelColor
	face font.Face
}

type comboButtonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    font.Face
	graphic *widget.ButtonImageImage
	padding widget.Insets
}

type listResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding widget.Insets
	handle       *widget.ButtonImage
	handleSize   int
	face         font.Face
	entry        *widget.ListEntryColor
	entryPadding widget.Insets
}

type sliderResources struct {
	trackImage *widget.SliderTrackImage
	handle     *widget.ButtonImage
	handleSize int
}

type progressBarResources struct {
	trackImage *widget.ProgressBarImage
	fillImage  *widget.ProgressBarImage
}

type panelResources struct {
	image    *image.NineSlice
	titleBar *image.NineSlice
	padding  widget.Insets
}

type tabBookResources struct {
	buttonFace    font.Face
	buttonText    *widget.ButtonTextColor
	buttonPadding widget.Insets
}

type headerResources struct {
	background *image.NineSlice
	padding    widget.Insets
	face       font.Face
	color      color.Color
}

type textInputResources struct {
	image   *widget.TextInputImage
	padding widget.Insets
	face    font.Face
	color   *widget.TextInputColor
}

type textAreaResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding widget.Insets
	handle       *widget.ButtonImage
	handleSize   int
	face         font.Face
	entryPadding widget.Insets
}

type toolTipResources struct {
	background *image.NineSlice
	padding    widget.Insets
	face       font.Face
	color      color.Color
}

type fonts struct {
	face         font.Face
	titleFace    font.Face
	bigTitleFace font.Face
	toolTipFace  font.Face
}

func NewUIResources() (*uiResources, error) {
	background := image.NewNineSliceColor(hexToColorAlpha(backgroundColor, 155))

	fonts, err := loadFonts()
	if err != nil {
		return nil, err
	}

	button, err := newButtonResources(fonts)
	if err != nil {
		return nil, err
	}

	checkbox, err := newCheckboxResources()
	if err != nil {
		return nil, err
	}

	comboButton, err := newComboButtonResources(fonts)
	if err != nil {
		return nil, err
	}

	list, err := newListResources(fonts)
	if err != nil {
		return nil, err
	}

	slider, err := newSliderResources()
	if err != nil {
		return nil, err
	}

	progressBar, err := newProgressBarResources()
	if err != nil {
		return nil, err
	}

	panel, err := newPanelResources()
	if err != nil {
		return nil, err
	}

	tabBook, err := newTabBookResources(fonts)
	if err != nil {
		return nil, err
	}

	header, err := newHeaderResources(fonts)
	if err != nil {
		return nil, err
	}

	textInput, err := newTextInputResources(fonts)
	if err != nil {
		return nil, err
	}
	textArea, err := newTextAreaResources(fonts)
	if err != nil {
		return nil, err
	}
	toolTip, err := newToolTipResources(fonts)
	if err != nil {
		return nil, err
	}

	return &uiResources{
		fonts:          fonts,
		background:     background,
		separatorColor: hexToColor(separatorColor),

		text: &textResources{
			idleColor:     hexToColor(textIdleColor),
			disabledColor: hexToColor(textDisabledColor),
			face:          fonts.face,
			titleFace:     fonts.titleFace,
			bigTitleFace:  fonts.bigTitleFace,
			smallFace:     fonts.toolTipFace,
		},

		button:      button,
		label:       newLabelResources(fonts),
		checkbox:    checkbox,
		comboButton: comboButton,
		list:        list,
		slider:      slider,
		panel:       panel,
		tabBook:     tabBook,
		header:      header,
		textInput:   textInput,
		toolTip:     toolTip,
		textArea:    textArea,
		progressBar: progressBar,
	}, nil
}

func loadFonts() (*fonts, error) {
	fontFace, err := loadFont(fontFaceRegular, 20)
	if err != nil {
		return nil, err
	}

	titleFontFace, err := loadFont(fontFaceBold, 24)
	if err != nil {
		return nil, err
	}

	bigTitleFontFace, err := loadFont(fontFaceBold, 28)
	if err != nil {
		return nil, err
	}

	toolTipFace, err := loadFont(fontFaceRegular, 15)
	if err != nil {
		return nil, err
	}

	return &fonts{
		face:         fontFace,
		titleFace:    titleFontFace,
		bigTitleFace: bigTitleFontFace,
		toolTipFace:  toolTipFace,
	}, nil
}

func loadFont(path string, size float64) (font.Face, error) {
	fontData, err := embedded.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(ttfFont, &truetype.Options{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}), nil
}

func loadGraphicImages(idle string, disabled string) (*widget.ButtonImageImage, error) {
	idleImage, _, err := newImageFromFile(idle)
	if err != nil {
		return nil, err
	}

	var disabledImage *ebiten.Image
	if disabled != "" {
		disabledImage, _, err = newImageFromFile(disabled)
		if err != nil {
			return nil, err
		}
	}

	return &widget.ButtonImageImage{
		Idle:     idleImage,
		Disabled: disabledImage,
	}, nil
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int) (*image.NineSlice, error) {
	i, _, err := newImageFromFile(path)
	if err != nil {
		return nil, err
	}
	w, h := i.Bounds().Dx(), i.Bounds().Dy()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func newButtonResources(fonts *fonts) (*buttonResources, error) {
	idle, err := loadImageNineSlice("resources/menu/ui/button-idle.png", 12, 0)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("resources/menu/ui/button-hover.png", 12, 0)
	if err != nil {
		return nil, err
	}
	pressed_hover, err := loadImageNineSlice("resources/menu/ui/button-selected-hover.png", 12, 0)
	if err != nil {
		return nil, err
	}
	pressed, err := loadImageNineSlice("resources/menu/ui/button-pressed.png", 12, 0)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("resources/menu/ui/button-disabled.png", 12, 0)
	if err != nil {
		return nil, err
	}

	i := &widget.ButtonImage{
		Idle:         idle,
		Hover:        hover,
		Pressed:      pressed,
		PressedHover: pressed_hover,
		Disabled:     disabled,
	}

	return &buttonResources{
		image: i,

		text: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		face: fonts.face,

		padding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newCheckboxResources() (*checkboxResources, error) {
	idle, err := loadImageNineSlice("resources/menu/ui/checkbox-idle.png", 20, 0)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("resources/menu/ui/checkbox-hover.png", 20, 0)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("resources/menu/ui/checkbox-disabled.png", 20, 0)
	if err != nil {
		return nil, err
	}

	checked, err := loadGraphicImages("resources/menu/ui/checkbox-checked-idle.png", "resources/menu/ui/checkbox-checked-disabled.png")
	if err != nil {
		return nil, err
	}

	unchecked, err := loadGraphicImages("resources/menu/ui/checkbox-unchecked-idle.png", "resources/menu/ui/checkbox-unchecked-disabled.png")
	if err != nil {
		return nil, err
	}

	greyed, err := loadGraphicImages("resources/menu/ui/checkbox-greyed-idle.png", "resources/menu/ui/checkbox-greyed-disabled.png")
	if err != nil {
		return nil, err
	}

	return &checkboxResources{
		image: &widget.ButtonImage{
			Idle:     idle,
			Hover:    hover,
			Pressed:  hover,
			Disabled: disabled,
		},

		graphic: &widget.CheckboxGraphicImage{
			Checked:   checked,
			Unchecked: unchecked,
			Greyed:    greyed,
		},

		spacing: 10,
	}, nil
}

func newLabelResources(fonts *fonts) *labelResources {
	return &labelResources{
		text: &widget.LabelColor{
			Idle:     hexToColor(labelIdleColor),
			Disabled: hexToColor(labelDisabledColor),
		},

		face: fonts.face,
	}
}

func newComboButtonResources(fonts *fonts) (*comboButtonResources, error) {
	idle, err := loadImageNineSlice("resources/menu/ui/combo-button-idle.png", 12, 0)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("resources/menu/ui/combo-button-hover.png", 12, 0)
	if err != nil {
		return nil, err
	}

	pressed, err := loadImageNineSlice("resources/menu/ui/combo-button-pressed.png", 12, 0)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("resources/menu/ui/combo-button-disabled.png", 12, 0)
	if err != nil {
		return nil, err
	}

	i := &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}

	arrowDown, err := loadGraphicImages("resources/menu/ui/arrow-down-idle.png", "resources/menu/ui/arrow-down-disabled.png")
	if err != nil {
		return nil, err
	}

	return &comboButtonResources{
		image: i,

		text: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		face:    fonts.face,
		graphic: arrowDown,

		padding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newListResources(fonts *fonts) (*listResources, error) {
	idle, _, err := newImageFromFile("resources/menu/ui/list-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := newImageFromFile("resources/menu/ui/list-disabled.png")
	if err != nil {
		return nil, err
	}

	mask, _, err := newImageFromFile("resources/menu/ui/list-mask.png")
	if err != nil {
		return nil, err
	}

	trackIdle, _, err := newImageFromFile("resources/menu/ui/list-track-idle.png")
	if err != nil {
		return nil, err
	}

	trackDisabled, _, err := newImageFromFile("resources/menu/ui/list-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := newImageFromFile("resources/menu/ui/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := newImageFromFile("resources/menu/ui/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	return &listResources{
		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},

		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},

		trackPadding: widget.Insets{
			Top:    5,
			Bottom: 24,
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},

		handleSize: 5,
		face:       fonts.face,

		entry: &widget.ListEntryColor{
			Unselected:         hexToColor(textIdleColor),
			DisabledUnselected: hexToColor(textDisabledColor),

			Selected:         hexToColor(textIdleColor),
			DisabledSelected: hexToColor(textDisabledColor),

			SelectedBackground:         hexToColor(listSelectedBackground),
			DisabledSelectedBackground: hexToColor(listDisabledSelectedBackground),

			FocusedBackground:         hexToColor(listFocusedBackground),
			SelectedFocusedBackground: hexToColor(listSelectedBackground),
		},

		entryPadding: widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
	}, nil
}

func newSliderResources() (*sliderResources, error) {
	idle, _, err := newImageFromFile("resources/menu/ui/slider-track-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := newImageFromFile("resources/menu/ui/slider-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := newImageFromFile("resources/menu/ui/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := newImageFromFile("resources/menu/ui/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	handleDisabled, _, err := newImageFromFile("resources/menu/ui/slider-handle-disabled.png")
	if err != nil {
		return nil, err
	}

	return &sliderResources{
		trackImage: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(idle, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
			Hover:    image.NewNineSlice(idle, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
			Disabled: image.NewNineSlice(disabled, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleDisabled, 0, 5),
		},

		handleSize: 6,
	}, nil
}

func newProgressBarResources() (*progressBarResources, error) {
	idle, _, err := newImageFromFile("resources/menu/ui/progressbar-track-idle.png")
	if err != nil {
		return nil, err
	}
	fill_idle, _, err := newImageFromFile("resources/menu/ui/progressbar-fill-idle.png")
	if err != nil {
		return nil, err
	}
	disabled, _, err := newImageFromFile("resources/menu/ui/slider-track-disabled.png")
	if err != nil {
		return nil, err
	}

	return &progressBarResources{
		trackImage: &widget.ProgressBarImage{
			Idle:     image.NewNineSlice(idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Hover:    image.NewNineSlice(idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Disabled: image.NewNineSlice(disabled, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
		},

		fillImage: &widget.ProgressBarImage{
			Idle:     image.NewNineSlice(fill_idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Hover:    image.NewNineSlice(fill_idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Disabled: image.NewNineSlice(fill_idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
		},
	}, nil
}
func newPanelResources() (*panelResources, error) {
	i, err := loadImageNineSlice("resources/menu/ui/panel-idle.png", 10, 10)
	if err != nil {
		return nil, err
	}
	t, err := loadImageNineSlice("resources/menu/ui/titlebar-idle.png", 10, 10)
	if err != nil {
		return nil, err
	}
	return &panelResources{
		image:    i,
		titleBar: t,
		padding: widget.Insets{
			Left:   30,
			Right:  30,
			Top:    20,
			Bottom: 20,
		},
	}, nil
}

func newTabBookResources(fonts *fonts) (*tabBookResources, error) {

	return &tabBookResources{
		buttonFace: fonts.face,

		buttonText: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		buttonPadding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newHeaderResources(fonts *fonts) (*headerResources, error) {
	bg, err := loadImageNineSlice("resources/menu/ui/header.png", 446, 9)
	if err != nil {
		return nil, err
	}

	return &headerResources{
		background: bg,

		padding: widget.Insets{
			Left:   25,
			Right:  25,
			Top:    4,
			Bottom: 4,
		},

		face:  fonts.bigTitleFace,
		color: hexToColor(headerColor),
	}, nil
}

func newTextInputResources(fonts *fonts) (*textInputResources, error) {
	idle, _, err := newImageFromFile("resources/menu/ui/text-input-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := newImageFromFile("resources/menu/ui/text-input-disabled.png")
	if err != nil {
		return nil, err
	}

	return &textInputResources{
		image: &widget.TextInputImage{
			Idle:     image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
			Disabled: image.NewNineSlice(disabled, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
		},

		padding: widget.Insets{
			Left:   8,
			Right:  8,
			Top:    4,
			Bottom: 4,
		},

		face: fonts.face,

		color: &widget.TextInputColor{
			Idle:          hexToColor(textIdleColor),
			Disabled:      hexToColor(textDisabledColor),
			Caret:         hexToColor(textInputCaretColor),
			DisabledCaret: hexToColor(textInputDisabledCaretColor),
		},
	}, nil
}

func newTextAreaResources(fonts *fonts) (*textAreaResources, error) {
	idle, _, err := newImageFromFile("resources/menu/ui/list-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := newImageFromFile("resources/menu/ui/list-disabled.png")
	if err != nil {
		return nil, err
	}

	mask, _, err := newImageFromFile("resources/menu/ui/list-mask.png")
	if err != nil {
		return nil, err
	}

	trackIdle, _, err := newImageFromFile("resources/menu/ui/list-track-idle.png")
	if err != nil {
		return nil, err
	}

	trackDisabled, _, err := newImageFromFile("resources/menu/ui/list-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := newImageFromFile("resources/menu/ui/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := newImageFromFile("resources/menu/ui/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	return &textAreaResources{
		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},

		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},

		trackPadding: widget.Insets{
			Top:    5,
			Bottom: 24,
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},

		handleSize: 5,
		face:       fonts.face,

		entryPadding: widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
	}, nil
}

func newToolTipResources(fonts *fonts) (*toolTipResources, error) {
	bg, _, err := newImageFromFile("resources/menu/ui/tool-tip.png")
	if err != nil {
		return nil, err
	}

	return &toolTipResources{
		background: image.NewNineSlice(bg, [3]int{19, 6, 13}, [3]int{19, 5, 13}),

		padding: widget.Insets{
			Left:   15,
			Right:  15,
			Top:    10,
			Bottom: 10,
		},

		face:  fonts.toolTipFace,
		color: hexToColor(toolTipColor),
	}, nil
}

func hexToColor(h string) color.Color {
	return hexToColorAlpha(h, 255)
}

func hexToColorAlpha(h string, alpha uint8) color.Color {
	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		panic(err)
	}

	return color.NRGBA{
		R: uint8(u & 0xff0000 >> 16),
		G: uint8(u & 0xff00 >> 8),
		B: uint8(u & 0xff),
		A: alpha,
	}
}
