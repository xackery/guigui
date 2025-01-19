// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
)

func easeOutQuad(t float64) float64 {
	// https://greweb.me/2012/02/bezier-curve-based-easing-functions-from-concept-to-implementation
	// easeOutQuad
	return t * (2 - t)
}

func popupMaxOpacity() int {
	return ebiten.TPS() / 6
}

type Popup struct {
	guigui.DefaultWidget

	//callback *PopupCallback

	content    popupContent
	background popupBackground

	opacity           int
	showing           bool
	hiding            bool
	backgroundBlurred bool

	initOnce sync.Once
}

func (p *Popup) IsPopup() bool {
	return true
}

func (p *Popup) SetContent(f func(context *guigui.Context, childAppender *ContainerChildWidgetAppender)) {
	p.content.setContent(f)
}

func (p *Popup) SetContentBounds(bounds image.Rectangle) {
	guigui.SetPosition(&p.content, bounds.Min)
	p.content.setSize(bounds.Dx(), bounds.Dy())
}

func (p *Popup) SetBackgroundBlurred(blurBackground bool) {
	p.backgroundBlurred = blurBackground
}

func (p *Popup) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	p.initOnce.Do(func() {
		guigui.Hide(p)
	})

	if p.backgroundBlurred {
		p.background.popup = p
		appender.AppendChildWidget(&p.background)
	}
	appender.AppendChildWidget(&p.content)

}

func (p *Popup) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	// As this editor is a modal dialog, do not let other widgets to handle inputs.
	if image.Pt(ebiten.CursorPosition()).In(guigui.VisibleBounds(p)) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			guigui.Hide(p)
		}
		return guigui.AbortHandlingInput()
	}
	return guigui.HandleInputResult{}
}

func (p *Popup) Open() {
	guigui.Show(p)
	p.showing = true
	p.hiding = false
}

func (p *Popup) Close() {
	p.showing = false
	p.hiding = true
}

func (p *Popup) Update(context *guigui.Context) error {
	if p.showing {
		if p.opacity < popupMaxOpacity() {
			p.opacity++
		}
		guigui.SetOpacity(&p.content, easeOutQuad(float64(p.opacity)/float64(popupMaxOpacity())))
		guigui.RequestRedraw(&p.background)
		if p.opacity == popupMaxOpacity() {
			p.showing = false
		}
	}
	if p.hiding {
		if 0 < p.opacity {
			p.opacity--
		}
		guigui.SetOpacity(&p.content, easeOutQuad(float64(p.opacity)/float64(popupMaxOpacity())))
		guigui.RequestRedraw(&p.background)
		if p.opacity == 0 {
			p.hiding = false
			guigui.Hide(p)
		}
	}
	return nil
}

func (p *Popup) Size(context *guigui.Context) (int, int) {
	return context.AppSize()
}

type popupContent struct {
	guigui.DefaultWidget

	setContentFunc func(context *guigui.Context, childAppender *ContainerChildWidgetAppender)
	childWidgets   ContainerChildWidgetAppender

	width  int
	height int
}

func (p *popupContent) setContent(f func(context *guigui.Context, childAppender *ContainerChildWidgetAppender)) {
	p.setContentFunc = f
}

func (p *popupContent) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	p.childWidgets.reset()
	if p.setContentFunc != nil {
		p.setContentFunc(context, &p.childWidgets)
	}
	for _, childWidget := range p.childWidgets.iter() {
		appender.AppendChildWidget(childWidget)
	}
}

func (p *popupContent) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	return guigui.AbortHandlingInput()
}

func (p *popupContent) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := guigui.Bounds(p)
	DrawRoundedRect(context, dst, bounds, Color2(context.ColorMode(), ColorTypeBase, 1, 0.05), RoundedCornerRadius(context))
	DrawRoundedRectBorder(context, dst, bounds, Color2(context.ColorMode(), ColorTypeBase, 0.7, 0), RoundedCornerRadius(context), float32(1*context.Scale()), RoundedRectBorderTypeOutset)
}

func (p *popupContent) setSize(width, height int) {
	p.width = width
	p.height = height
}

func (p *popupContent) Size(context *guigui.Context) (int, int) {
	return p.width, p.height
}

type popupBackground struct {
	guigui.DefaultWidget

	popup *Popup

	backgroundCache *ebiten.Image
}

func (p *popupBackground) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := guigui.Bounds(p)
	if p.backgroundCache != nil && !bounds.In(p.backgroundCache.Bounds()) {
		p.backgroundCache.Deallocate()
		p.backgroundCache = nil
	}
	if p.backgroundCache == nil {
		p.backgroundCache = ebiten.NewImageWithOptions(bounds, nil)
	}

	rate := easeOutQuad(float64(p.popup.opacity) / float64(popupMaxOpacity()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dst.Bounds().Min.X), float64(dst.Bounds().Min.Y))
	p.backgroundCache.DrawImage(dst, op)

	DrawBlurredImage(dst, p.backgroundCache, rate)
}

func (p *popupBackground) Size(context *guigui.Context) (int, int) {
	return guigui.Parent(p).Size(context)
}
