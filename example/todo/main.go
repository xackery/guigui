// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
)

type Task struct {
	ID        uuid.UUID
	Text      string
	CreatedAt time.Time
}

func NewTask(text string) Task {
	return Task{
		ID:        uuid.New(),
		Text:      text,
		CreatedAt: time.Now(),
	}
}

type TaskWidgets struct {
	doneButtonWidget *guigui.Widget
	textWidget       *guigui.Widget
}

type Root struct {
	guigui.RootWidgetBehavior

	createButtonWidget *guigui.Widget
	textFieldWidget    *guigui.Widget
	taskWidgets        map[uuid.UUID]*TaskWidgets
	tasksPanelWidget   *guigui.Widget

	tasks []Task
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	if r.textFieldWidget == nil {
		var t basicwidget.TextField
		r.textFieldWidget = guigui.NewWidget(&t)
	}
	width, _ := widget.Size(context)
	w := width - int(6.5*u)
	r.textFieldWidget.Behavior().(*basicwidget.TextField).SetSize(context, w, int(u))
	{
		p := widget.Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(r.textFieldWidget, p)
	}

	if r.createButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Create")
		b.SetWidth(int(5 * u))
		r.createButtonWidget = guigui.NewWidget(&b)
	}
	{
		p := widget.Position()
		w, _ := widget.Size(context)
		p.X += w - int(0.5*u) - int(5*u)
		p.Y += int(0.5 * u)
		appender.AppendChildWidget(r.createButtonWidget, p)
	}

	if r.tasksPanelWidget == nil {
		var sp basicwidget.ScrollablePanel
		r.tasksPanelWidget = guigui.NewWidget(&sp)
	}
	tasksSP := r.tasksPanelWidget.Behavior().(*basicwidget.ScrollablePanel)
	w, h := widget.Size(context)
	tasksSP.SetSize(context, w, h-int(2*u))
	tasksSP.SetContent(func(context *guigui.Context, widget *guigui.Widget, childAppender *basicwidget.ScrollablePanelChildWidgetAppender) {
		p := widget.Position()
		minX := p.X + int(0.5*u)
		y := p.Y
		for i, t := range r.tasks {
			if _, ok := r.taskWidgets[t.ID]; !ok {
				var b basicwidget.TextButton
				b.SetText("Done")
				b.SetWidth(int(3 * u))
				var text basicwidget.Text
				text.SetText(t.Text)
				text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
				if r.taskWidgets == nil {
					r.taskWidgets = map[uuid.UUID]*TaskWidgets{}
				}
				r.taskWidgets[t.ID] = &TaskWidgets{
					doneButtonWidget: guigui.NewWidget(&b),
					textWidget:       guigui.NewWidget(&text),
				}
			}
			if i > 0 {
				y += int(u / 4)
			}
			childAppender.AppendChildWidget(r.taskWidgets[t.ID].doneButtonWidget, image.Pt(minX, y))
			r.taskWidgets[t.ID].textWidget.Behavior().(*basicwidget.Text).SetSize(w-int(4.5*u), int(u))
			childAppender.AppendChildWidget(r.taskWidgets[t.ID].textWidget, image.Pt(minX+int(3.5*u), y))
			y += int(u)
		}
	})
	tasksSP.SetPadding(0, int(0.5*u))
	appender.AppendChildWidget(r.tasksPanelWidget, image.Pt(0, int(2*u)))

	// GC widgets
	for id := range r.taskWidgets {
		if slices.IndexFunc(r.tasks, func(t Task) bool {
			return t.ID == id
		}) >= 0 {
			continue
		}
		delete(r.taskWidgets, id)
	}
}

func (r *Root) Update(context *guigui.Context, widget *guigui.Widget) error {
	for event := range r.createButtonWidget.DequeueEvents() {
		switch e := event.(type) {
		case basicwidget.ButtonEvent:
			if e.Type == basicwidget.ButtonEventTypeUp {
				r.tryCreateTask()
			}
		}
	}
	for event := range r.textFieldWidget.DequeueEvents() {
		switch e := event.(type) {
		case basicwidget.TextEvent:
			if e.Type == basicwidget.TextEventTypeEnterPressed {
				r.tryCreateTask()
			}
		}
	}
	for id, t := range r.taskWidgets {
		for event := range t.doneButtonWidget.DequeueEvents() {
			switch e := event.(type) {
			case basicwidget.ButtonEvent:
				if e.Type == basicwidget.ButtonEventTypeUp {
					r.tasks = slices.DeleteFunc(r.tasks, func(task Task) bool {
						return task.ID == id
					})
				}
			}
		}
	}

	if r.canCreateTask() {
		r.createButtonWidget.Enable()
	} else {
		r.createButtonWidget.Disable()
	}

	return nil
}

func (r *Root) canCreateTask() bool {
	t := r.textFieldWidget.Behavior().(*basicwidget.TextField)
	str := t.Text()
	str = strings.TrimSpace(str)
	return str != ""
}

func (r *Root) tryCreateTask() {
	t := r.textFieldWidget.Behavior().(*basicwidget.TextField)
	str := t.Text()
	str = strings.TrimSpace(str)
	if str != "" {
		r.tasks = slices.Insert(r.tasks, 0, NewTask(str))
		t.SetText("")
	}
}

func (r *Root) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:           "TODO",
		WindowMinWidth:  320,
		WindowMinHeight: 240,
	}
	if err := guigui.Run(guigui.NewWidget(&Root{}), op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
