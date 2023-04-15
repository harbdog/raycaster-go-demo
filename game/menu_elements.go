package game

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func titleBarContainer(menu *DemoMenu) *widget.Container {
	res := menu.res

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Stretch([]bool{true, false, false}, []bool{true}), widget.GridLayoutOpts.Padding(widget.Insets{
			Left:   30,
			Right:  5,
			Top:    6,
			Bottom: 5,
		}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Raycaster Go Demo Settings", res.text.titleFace, res.textInput.color.Idle),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c.AddChild(widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("X", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			menu.game.closeMenu()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	return c
}

func footerContainer(menu *DemoMenu) *widget.Container {
	res := menu.res

	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.Insets{
			Left:  25,
			Right: 25,
		}),
	)))
	c.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/harbdog/raycaster-go-demo", res.text.smallFace, res.text.disabledColor)))
	return c
}

func settingsContainer(menu *DemoMenu) widget.PreferredSizeLocateableWidget {
	res := menu.res

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0),
		)))

	pages := []interface{}{
		gamePage(menu),
		displayPage(menu),
		renderPage(menu),
		lightingPage(menu),
	}

	pageContainer := newPageContainer(res)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*page).title
		}),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.MinHandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			pageContainer.setPage(args.Entry.(*page))
			menu.root.RequestRelayout()
		}))
	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
}

func newCheckbox(label string, checked bool, changedHandler widget.CheckboxChangedHandlerFunc, res *uiResources) *widget.LabeledCheckbox {
	c := widget.NewLabeledCheckbox(
		widget.LabeledCheckboxOpts.Spacing(res.checkbox.spacing),
		widget.LabeledCheckboxOpts.CheckboxOpts(
			widget.CheckboxOpts.ButtonOpts(widget.ButtonOpts.Image(res.checkbox.image)),
			widget.CheckboxOpts.Image(res.checkbox.graphic),
			widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				if changedHandler != nil {
					changedHandler(args)
				}
			})),
		widget.LabeledCheckboxOpts.LabelOpts(widget.LabelOpts.Text(label, res.label.face, res.label.text)))

	if checked {
		c.SetState(widget.WidgetChecked)
	}

	return c
}

func newListComboButton(entries []interface{}, selectedEntry interface{}, buttonLabel widget.SelectComboButtonEntryLabelFunc, entryLabel widget.ListEntryLabelFunc,
	entrySelectedHandler widget.ListComboButtonEntrySelectedHandlerFunc, res *uiResources) *widget.ListComboButton {

	c := widget.NewListComboButton(
		widget.ListComboButtonOpts.SelectComboButtonOpts(
			widget.SelectComboButtonOpts.ComboButtonOpts(
				widget.ComboButtonOpts.ButtonOpts(
					widget.ButtonOpts.Image(res.comboButton.image),
					widget.ButtonOpts.TextPadding(res.comboButton.padding),
				),
			),
		),
		widget.ListComboButtonOpts.Text(res.comboButton.face, res.comboButton.graphic, res.comboButton.text),
		widget.ListComboButtonOpts.ListOpts(
			widget.ListOpts.Entries(entries),
			widget.ListOpts.ScrollContainerOpts(
				widget.ScrollContainerOpts.Image(res.list.image),
			),
			widget.ListOpts.SliderOpts(
				widget.SliderOpts.Images(res.list.track, res.list.handle),
				widget.SliderOpts.MinHandleSize(res.list.handleSize),
				widget.SliderOpts.TrackPadding(res.list.trackPadding)),
			widget.ListOpts.EntryFontFace(res.list.face),
			widget.ListOpts.EntryColor(res.list.entry),
			widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		),
		widget.ListComboButtonOpts.EntryLabelFunc(buttonLabel, entryLabel),
		widget.ListComboButtonOpts.EntrySelectedHandler(entrySelectedHandler))

	if selectedEntry != nil {
		c.SetSelectedEntry(selectedEntry)
	}

	return c
}

func newSeparator(res *uiResources, ld interface{}) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.separatorColor)),
	))

	return c
}
