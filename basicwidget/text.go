// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"log/slog"
	"math"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/internal/clipboard"
)

type TextEventType int

const (
	TextEventTypeEnterPressed TextEventType = iota
)

type TextEvent struct {
	Type TextEventType
	Text string
}

func isKeyRepeating(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	// In the current implementation of text, d == 1 might be skipped especially for backspace key.
	// TODO: Fix this.
	if d == 2 {
		return true
	}
	if d < 24 {
		return false
	}
	return (d-24)%4 == 0
}

type TextFilter func(text string, start, end int) (string, int, int)

type Text struct {
	field textinput.Field

	hAlign      HorizontalAlign
	vAlign      VerticalAlign
	color       color.Color
	transparent float64
	lang        language.Tag
	scaleMinus1 float64
	bold        bool

	selectable           bool
	editable             bool
	multiline            bool
	selectionDragStart   int
	selectionShiftIndex  int
	dragging             bool
	toAdjustScrollOffset bool
	prevFocused          bool

	filter TextFilter

	cursor              textCursor
	cursorWidget        *guigui.Widget
	scrollOverlayWidget *guigui.Widget

	needsRedraw bool
}

func (t *Text) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if t.cursorWidget == nil {
		t.cursorWidget = guigui.NewPopupWidget(&t.cursor)
	}
	appender.AppendChildWidget(t.cursorWidget, t.cursorBounds(context, widget))

	if t.scrollOverlayWidget == nil {
		t.scrollOverlayWidget = guigui.NewWidget(&ScrollOverlay{})
		t.scrollOverlayWidget.Hide()
	}
	appender.AppendChildWidget(t.scrollOverlayWidget, widget.Bounds())
}

func (t *Text) SetSelectable(selectable bool) {
	if t.selectable == selectable {
		return
	}
	t.selectable = selectable
	t.selectionDragStart = -1
	t.selectionShiftIndex = -1
	t.needsRedraw = true
}

func (t *Text) Text() string {
	return t.field.Text()
}

func (t *Text) SetText(text string) {
	start, end := t.field.Selection()
	start = min(start, len(text))
	end = min(end, len(text))
	t.setTextAndSelection(text, start, end, -1)
}

func (t *Text) SetFilter(filter TextFilter) {
	t.filter = filter
	t.applyFilter()
}

func (t *Text) selectAll() {
	t.setTextAndSelection(t.field.Text(), 0, len(t.field.Text()), -1)
}

func (t *Text) setTextAndSelection(text string, start, end int, shiftIndex int) {
	t.selectionShiftIndex = shiftIndex
	if start > end {
		start, end = end, start
	}

	if s, e := t.field.Selection(); t.field.Text() == text && s == start && e == end {
		return
	}
	t.field.SetTextAndSelection(text, start, end)
	t.toAdjustScrollOffset = true
	t.needsRedraw = true
}

func (t *Text) SetLanguage(lang language.Tag) {
	if t.lang == lang {
		return
	}

	t.lang = lang
	t.needsRedraw = true
}

func (t *Text) SetBold(bold bool) {
	if t.bold == bold {
		return
	}

	t.bold = bold
	t.needsRedraw = true
}

func (t *Text) SetScale(scale float64) {
	if t.scaleMinus1 == scale-1 {
		return
	}

	t.scaleMinus1 = scale - 1
	t.needsRedraw = true
}

func (t *Text) SetHorizontalAlign(align HorizontalAlign) {
	if t.hAlign == align {
		return
	}

	t.hAlign = align
	t.needsRedraw = true
}

func (t *Text) SetVerticalAlign(align VerticalAlign) {
	if t.vAlign == align {
		return
	}

	t.vAlign = align
	t.needsRedraw = true
}

func (t *Text) SetColor(color color.Color) {
	if equalColor(t.color, color) {
		return
	}

	t.color = color
	t.needsRedraw = true
}

func (t *Text) SetOpacity(opacity float64) {
	if 1-t.transparent == opacity {
		return
	}

	t.transparent = 1 - opacity
	t.needsRedraw = true
}

func (t *Text) SetEditable(editable bool) {
	if t.editable == editable {
		return
	}

	if editable {
		t.selectionDragStart = -1
		t.selectionShiftIndex = -1
	}
	t.editable = editable
	t.needsRedraw = true
}

func (t *Text) SetScrollable(scrollable bool) {
	if t.scrollOverlayWidget == nil {
		t.scrollOverlayWidget = guigui.NewWidget(&ScrollOverlay{})
		t.scrollOverlayWidget.Hide()
	}
	if scrollable {
		t.scrollOverlayWidget.Show()
	} else {
		t.scrollOverlayWidget.Hide()
	}
}

func (t *Text) IsMultiline() bool {
	return t.multiline
}

func (t *Text) SetMultiline(multiline bool) {
	if t.multiline == multiline {
		return
	}

	t.multiline = multiline
	t.needsRedraw = true
}

func (t *Text) textBounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	offsetX, offsetY := t.scrollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
	b := widget.Bounds()

	switch t.hAlign {
	case HorizontalAlignStart:
	case HorizontalAlignCenter:
		// TODO: What is the correct value?
	case HorizontalAlignEnd:
		b.Max.X += int(offsetX)
		w, _ := text.Measure(t.field.Text(), t.face(context), t.lineHeight(context))
		b.Max.X += max(int(w)-b.Dx(), 0)
	}

	txt := t.textToDraw()
	if txt == "" {
		txt = " "
	}
	switch t.vAlign {
	case VerticalAlignTop:
	case VerticalAlignMiddle:
		h := b.Dy()
		th := t.textHeight(context, txt)
		b.Min.Y += (h - th) / 2
		b.Max.Y = b.Min.Y + th
	case VerticalAlignBottom:
		h := b.Dy()
		th := t.textHeight(context, txt)
		b.Min.Y = h - th
	}

	b.Min.X += int(offsetX)
	b.Min.Y += int(offsetY)
	return b
}

func (t *Text) face(context *guigui.Context) text.Face {
	size := FontSize(context) * (t.scaleMinus1 + 1)
	weight := text.WeightMedium
	if t.bold {
		weight = text.WeightBold
	}
	return FontFace(size, weight, t.lang)
}

func (t *Text) lineHeight(context *guigui.Context) float64 {
	return LineHeight(context) * (t.scaleMinus1 + 1)
}

func (t *Text) HandleInput(context *guigui.Context, widget *guigui.Widget) guigui.HandleInputResult {
	if !t.selectable && !t.editable {
		return guigui.HandleInputResult{}
	}

	textBounds := t.textBounds(context, widget)

	face := t.face(context)
	x, y := ebiten.CursorPosition()
	if t.dragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			idx := textIndexFromPosition(textBounds, x, y, t.field.Text(), face, t.lineHeight(context), t.hAlign, t.vAlign)
			if idx < t.selectionDragStart {
				t.setTextAndSelection(t.field.Text(), idx, t.selectionDragStart, -1)
			} else {
				t.setTextAndSelection(t.field.Text(), t.selectionDragStart, idx, -1)
			}
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			t.dragging = false
			t.selectionDragStart = -1
		}
		return guigui.HandleInputByWidget(widget)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if image.Pt(x, y).In(widget.VisibleBounds()) {
			t.dragging = true
			idx := textIndexFromPosition(textBounds, x, y, t.field.Text(), face, t.lineHeight(context), t.hAlign, t.vAlign)
			t.selectionDragStart = idx
			widget.Focus()
			if start, end := t.field.Selection(); start != idx || end != idx {
				t.setTextAndSelection(t.field.Text(), idx, idx, -1)
			}
			return guigui.HandleInputByWidget(widget)
		}
		widget.Blur()
	}

	if !widget.IsFocused() {
		if t.field.IsFocused() {
			t.field.Blur()
			widget.RequestRedraw()
		}
		return guigui.HandleInputResult{}
	}
	t.field.Focus()

	if !t.editable && !t.selectable {
		return guigui.HandleInputResult{}
	}

	start, _ := t.field.Selection()
	var processed bool
	if x, _, bottom, ok := textPosition(textBounds, t.field.Text(), start, face, t.lineHeight(context), t.hAlign, t.vAlign); ok {
		var err error
		processed, err = t.field.HandleInput(int(x), int(bottom))
		if err != nil {
			slog.Error(err.Error())
			processed = false
		}
	}
	if processed {
		widget.RequestRedraw()
		t.adjustScrollOffset(context, widget)
		return guigui.HandleInputByWidget(widget)
	}

	// Do not accept key inputs when compositing.
	if _, _, ok := t.field.CompositionSelection(); ok {
		return guigui.HandleInputResult{}
	}

	// TODO: Use WebAPI to detect OS is runtime.GOOS == "js"
	isWindows := runtime.GOOS == "windows"
	isDarwin := runtime.GOOS == "darwin"

	if t.editable {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
			if t.multiline {
				start, end := t.field.Selection()
				text := t.field.Text()[:start] + "\n" + t.field.Text()[end:]
				t.setTextAndSelection(text, start+len("\n"), start+len("\n"), -1)
			}
			t.applyFilter()
			widget.EnqueueEvent(TextEvent{
				Type: TextEventTypeEnterPressed,
				Text: t.field.Text(),
			})
		case isKeyRepeating(ebiten.KeyBackspace) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyH):
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if start > 0 {
				text, pos := backspaceOnClusters(t.field.Text(), face, start)
				t.setTextAndSelection(text, pos, pos, -1)
			}
		case isWindows && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD):
			// Delete
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if isDarwin && end < len(t.field.Text()) {
				text, pos := deleteOnClusters(t.field.Text(), face, end)
				t.setTextAndSelection(text, pos, pos, -1)
			}

		case isWindows && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyX) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyX):
			// Cut
			start, end := t.field.Selection()
			if start != end {
				if err := clipboard.WriteAll(t.field.Text()[start:end]); err != nil {
					slog.Error(err.Error())
					return guigui.HandleInputResult{}
				}
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			}
		case isWindows && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyV) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyV):
			// Paste
			start, end := t.field.Selection()
			ct, err := clipboard.ReadAll()
			if err != nil {
				slog.Error(err.Error())
				return guigui.HandleInputResult{}
			}
			text := t.field.Text()[:start] + ct + t.field.Text()[end:]
			t.setTextAndSelection(text, start+len(ct), start, -1)
		}
	}

	switch {
	case isKeyRepeating(ebiten.KeyLeft) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyB):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndex == end {
				pos := prevPositionOnClusters(t.field.Text(), face, end)
				t.setTextAndSelection(t.field.Text(), start, pos, pos)
			} else {
				pos := prevPositionOnClusters(t.field.Text(), face, start)
				t.setTextAndSelection(t.field.Text(), pos, end, pos)
			}
		} else {
			if start != end {
				t.setTextAndSelection(t.field.Text(), start, start, -1)
			} else if start > 0 {
				pos := prevPositionOnClusters(t.field.Text(), face, start)
				t.setTextAndSelection(t.field.Text(), pos, pos, -1)
			}
		}
	case isKeyRepeating(ebiten.KeyRight) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyF):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndex == start {
				pos := nextPositionOnClusters(t.field.Text(), face, start)
				t.setTextAndSelection(t.field.Text(), pos, end, pos)
			} else {
				pos := nextPositionOnClusters(t.field.Text(), face, end)
				t.setTextAndSelection(t.field.Text(), start, pos, pos)
			}
		} else {
			if start != end {
				t.setTextAndSelection(t.field.Text(), end, end, -1)
			} else if start < len(t.field.Text()) {
				pos := nextPositionOnClusters(t.field.Text(), face, start)
				t.setTextAndSelection(t.field.Text(), pos, pos, -1)
			}
		}
	case isKeyRepeating(ebiten.KeyUp) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyP):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveEnd bool
		start, end := t.field.Selection()
		idx := start
		if shift && t.selectionShiftIndex == end {
			idx = end
			moveEnd = true
		}
		if x, y0, y1, ok := textPosition(textBounds, t.field.Text(), idx, face, lh, t.hAlign, t.vAlign); ok {
			y := (y0+y1)/2 - lh
			idx := textIndexFromPosition(textBounds, int(x), int(y), t.field.Text(), face, lh, t.hAlign, t.vAlign)
			if shift {
				if moveEnd {
					t.setTextAndSelection(t.field.Text(), start, idx, idx)
				} else {
					t.setTextAndSelection(t.field.Text(), idx, end, idx)
				}
			} else {
				t.setTextAndSelection(t.field.Text(), idx, idx, -1)
			}
		}
	case isKeyRepeating(ebiten.KeyDown) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyN):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveStart bool
		start, end := t.field.Selection()
		idx := end
		if shift && t.selectionShiftIndex == start {
			idx = start
			moveStart = true
		}
		if x, y0, y1, ok := textPosition(textBounds, t.field.Text(), idx, face, lh, t.hAlign, t.vAlign); ok {
			y := (y0+y1)/2 + lh
			idx := textIndexFromPosition(textBounds, int(x), int(y), t.field.Text(), face, lh, t.hAlign, t.vAlign)
			if shift {
				if moveStart {
					t.setTextAndSelection(t.field.Text(), idx, end, idx)
				} else {
					t.setTextAndSelection(t.field.Text(), start, idx, idx)
				}
			} else {
				t.setTextAndSelection(t.field.Text(), idx, idx, -1)
			}
		}
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA):
		idx := 0
		start, end := t.field.Selection()
		if i := strings.LastIndex(t.field.Text()[:start], "\n"); i >= 0 {
			idx = i + 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			t.setTextAndSelection(t.field.Text(), idx, end, idx)
		} else {
			t.setTextAndSelection(t.field.Text(), idx, idx, -1)
		}
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyE):
		idx := len(t.field.Text())
		start, end := t.field.Selection()
		if i := strings.Index(t.field.Text()[end:], "\n"); i >= 0 {
			idx = end + i
		}
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			t.setTextAndSelection(t.field.Text(), start, idx, idx)
		} else {
			t.setTextAndSelection(t.field.Text(), idx, idx, -1)
		}
	case isWindows && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyA):
		t.selectAll()
	case isWindows && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyC) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyC):
		// Copy
		start, end := t.field.Selection()
		if start != end {
			if err := clipboard.WriteAll(t.field.Text()[start:end]); err != nil {
				slog.Error(err.Error())
				return guigui.HandleInputResult{}
			}
		}
	}

	return guigui.HandleInputResult{}
}

func (t *Text) adjustScrollOffset(context *guigui.Context, widget *guigui.Widget) {
	t.updateContentSize(context)

	s, e := t.selectionToDraw(widget)
	text := t.textToDraw()

	tb := t.textBounds(context, widget)
	face := t.face(context)
	if x, _, y, ok := textPosition(tb, text, e, face, t.lineHeight(context), t.hAlign, t.vAlign); ok {
		var dx, dy float64
		if max := float64(widget.Bounds().Max.X); x > max {
			dx = max - x
		}
		if max := float64(widget.Bounds().Max.Y); y > max {
			dy = max - y
		}
		t.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetOffsetByDelta(dx, dy)
	}
	if x, y, _, ok := textPosition(tb, text, s, face, t.lineHeight(context), t.hAlign, t.vAlign); ok {
		var dx, dy float64
		if min := float64(widget.Bounds().Min.X); x < min {
			dx = min - x
		}
		if min := float64(widget.Bounds().Min.Y); y < min {
			dy = min - y
		}
		t.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetOffsetByDelta(dx, dy)
	}
}

func (t *Text) textToDraw() string {
	return t.field.TextForRendering()
}

func (t *Text) selectionToDraw(widget *guigui.Widget) (int, int) {
	s, e := t.field.Selection()
	if t.editable && widget.IsFocused() {
		if cs, _, ok := t.field.CompositionSelection(); ok {
			s += cs
			e = s
		}
	}
	return s, e
}

func (t *Text) Update(context *guigui.Context, widget *guigui.Widget) error {
	if !t.prevFocused && widget.IsFocused() {
		t.field.Focus()
		t.cursor.resetCounter()
		start, end := t.field.Selection()
		if start < 0 || end < 0 {
			t.selectAll()
		}
	} else if t.prevFocused && !widget.IsFocused() {
		t.applyFilter()
	}

	if t.needsRedraw {
		widget.RequestRedraw()
		t.needsRedraw = false
	}

	if t.toAdjustScrollOffset && !widget.VisibleBounds().Empty() {
		t.adjustScrollOffset(context, widget)
		t.toAdjustScrollOffset = false
	}

	t.prevFocused = widget.IsFocused()

	return nil
}

func (t *Text) applyFilter() {
	if t.filter != nil {
		start, end := t.field.Selection()
		text, start, end := t.filter(t.field.Text(), start, end)
		t.setTextAndSelection(text, start, end, -1)
	}
}

func (t *Text) updateContentSize(context *guigui.Context) {
	face := t.face(context)
	w, h := text.Measure(t.textToDraw(), face, t.lineHeight(context))
	t.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetContentSize(int(math.Floor(w)), int(math.Floor(h)))
}

func (t *Text) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	textBounds := t.textBounds(context, widget)
	if !textBounds.Overlaps(widget.VisibleBounds()) {
		return
	}

	s, e := t.selectionToDraw(widget)
	text := t.textToDraw()
	face := t.face(context)

	// TODO: Change the color to detect focus.
	if start, end := t.field.Selection(); widget.IsFocused() && start >= 0 && end >= 0 {
		if s != e {
			var tailIndices []int
			for i, r := range text[s:e] {
				if r != '\n' {
					continue
				}
				tailIndices = append(tailIndices, s+i)
			}
			tailIndices = append(tailIndices, e)

			headIdx := s
			for _, idx := range tailIndices {
				x0, top0, bottom0, ok0 := textPosition(textBounds, text, headIdx, face, t.lineHeight(context), t.hAlign, t.vAlign)
				x1, _, _, ok1 := textPosition(textBounds, text, idx, face, t.lineHeight(context), t.hAlign, t.vAlign)
				if ok0 && ok1 {
					x := float32(x0)
					y := float32(top0)
					width := float32(x1 - x0)
					height := float32(bottom0 - top0)
					vector.DrawFilledRect(dst, x, y, width, height, Color(context.ColorMode(), ColorTypeAccent, 0.8), false)
				}
				headIdx = idx + 1
			}
		}
	}

	var clr color.RGBA
	if t.color != nil {
		clr = color.RGBAModel.Convert(t.color).(color.RGBA)
	} else {
		clr = color.RGBAModel.Convert(Color(context.ColorMode(), ColorTypeBase, 0.1)).(color.RGBA)
	}
	if t.transparent > 0 {
		opacity := 1 - t.transparent
		clr = color.RGBA{
			R: byte(float64(clr.R) * opacity),
			G: byte(float64(clr.G) * opacity),
			B: byte(float64(clr.B) * opacity),
			A: byte(float64(clr.A) * opacity),
		}
	}
	drawText(textBounds, dst, text, face, t.lineHeight(context), t.hAlign, t.vAlign, clr)
}

func (t *Text) TextWidth(context *guigui.Context) int {
	w, _ := text.Measure(t.textToDraw(), t.face(context), t.lineHeight(context))
	w *= t.scaleMinus1 + 1
	return int(w)
}

func (t *Text) TextHeight(context *guigui.Context) int {
	return t.textHeight(context, t.textToDraw())
}

func (t *Text) textHeight(context *guigui.Context, str string) int {
	if str == "" {
		return 0
	}
	// The text is already shifted by (height - (m.HAscent + m.Descent)) / 2.
	return int(t.lineHeight(context) * float64(strings.Count(str, "\n")+1))
}

func (t *Text) CursorShape(context *guigui.Context, widget *guigui.Widget) (ebiten.CursorShapeType, bool) {
	if t.selectable || t.editable {
		return ebiten.CursorShapeText, true
	}
	return 0, false
}

func (t *Text) cursorPosition(context *guigui.Context, widget *guigui.Widget) (x, top, bottom float64, ok bool) {
	if !widget.IsFocused() {
		return 0, 0, 0, false
	}
	if !t.editable {
		return 0, 0, 0, false
	}
	start, end := t.field.Selection()
	if start < 0 {
		return 0, 0, 0, false
	}
	if end < 0 {
		return 0, 0, 0, false
	}

	textBounds := t.textBounds(context, widget)
	if !textBounds.Overlaps(widget.VisibleBounds()) {
		return 0, 0, 0, false
	}

	s, e := t.selectionToDraw(widget)
	if s != e {
		return 0, 0, 0, false
	}

	text := t.textToDraw()
	face := t.face(context)
	return textPosition(textBounds, text, e, face, t.lineHeight(context), t.hAlign, t.vAlign)
}

func (t *Text) cursorBounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	x, top, bottom, ok := t.cursorPosition(context, widget)
	if !ok {
		return image.Rectangle{}
	}
	w := int(2 * context.Scale())
	return image.Rect(int(x)-w/2, int(top), int(x)+w/2, int(bottom))
}

type textCursor struct {
	guigui.DefaultWidgetBehavior

	counter    int
	prevShown  bool
	prevX      float64
	prevTop    float64
	prevBottom float64
	prevOK     bool
}

func (t *textCursor) resetCounter() {
	t.counter = 0
}

func (t *textCursor) Update(context *guigui.Context, widget *guigui.Widget) error {
	textWidget := widget.Parent()
	text := textWidget.Behavior().(*Text)
	x, top, bottom, ok := text.cursorPosition(context, textWidget)
	if t.prevX != x || t.prevTop != top || t.prevBottom != bottom || t.prevOK != ok {
		t.resetCounter()
	}
	t.prevX = x
	t.prevTop = top
	t.prevBottom = bottom
	t.prevOK = ok

	t.counter++
	if r := t.shouldRenderCursor(context, textWidget); t.prevShown != r {
		t.prevShown = r
		// TODO: This is not efficient. Improve this.
		widget.RequestRedraw()
	}
	return nil
}

func (t *textCursor) shouldRenderCursor(context *guigui.Context, textWidget *guigui.Widget) bool {
	offset := ebiten.TPS() / 2
	if t.counter > offset && (t.counter-offset)%ebiten.TPS() >= ebiten.TPS()/2 {
		return false
	}
	text := textWidget.Behavior().(*Text)
	_, _, _, ok := text.cursorPosition(context, textWidget)
	return ok
}

func (t *textCursor) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	textWidget := widget.Parent()
	if !t.shouldRenderCursor(context, textWidget) {
		return
	}
	dst.Fill(Color(context.ColorMode(), ColorTypeAccent, 0.4))
}
