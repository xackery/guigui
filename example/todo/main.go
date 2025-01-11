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
	doneButton basicwidget.TextButton
	text       basicwidget.Text
}

type Root struct {
	guigui.RootWidgetBehavior

	createButton basicwidget.TextButton
	textField    basicwidget.TextField
	taskWidgets  map[uuid.UUID]*TaskWidgets
	tasksPanel   basicwidget.ScrollablePanel

	tasks []Task
}

func (r *Root) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	width, _ := r.Size(context)
	w := width - int(6.5*u)
	r.textField.SetSize(context, w, int(u))
	{
		p := guigui.Position(r).Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(&r.textField, p)
	}

	r.createButton.SetText("Create")
	r.createButton.SetWidth(int(5 * u))
	{
		p := guigui.Position(r)
		w, _ := r.Size(context)
		p.X += w - int(0.5*u) - int(5*u)
		p.Y += int(0.5 * u)
		appender.AppendChildWidget(&r.createButton, p)
	}

	w, h := r.Size(context)
	r.tasksPanel.SetSize(context, w, h-int(2*u))
	r.tasksPanel.SetContent(func(context *guigui.Context, childAppender *basicwidget.ScrollablePanelChildWidgetAppender) {
		p := guigui.Position(&r.tasksPanel)
		minX := p.X + int(0.5*u)
		y := p.Y
		for i, t := range r.tasks {
			if _, ok := r.taskWidgets[t.ID]; !ok {
				var tw TaskWidgets
				tw.doneButton.SetText("Done")
				tw.doneButton.SetWidth(int(3 * u))
				tw.text.SetText(t.Text)
				tw.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
				if r.taskWidgets == nil {
					r.taskWidgets = map[uuid.UUID]*TaskWidgets{}
				}
				r.taskWidgets[t.ID] = &tw
			}
			if i > 0 {
				y += int(u / 4)
			}
			childAppender.AppendChildWidget(&r.taskWidgets[t.ID].doneButton, image.Pt(minX, y))
			r.taskWidgets[t.ID].text.SetSize(w-int(4.5*u), int(u))
			childAppender.AppendChildWidget(&r.taskWidgets[t.ID].text, image.Pt(minX+int(3.5*u), y))
			y += int(u)
		}
	})
	r.tasksPanel.SetPadding(0, int(0.5*u))
	appender.AppendChildWidget(&r.tasksPanel, image.Pt(0, int(2*u)))

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

func (r *Root) Update(context *guigui.Context) error {
	for event := range guigui.DequeueEvents(&r.createButton) {
		switch e := event.(type) {
		case basicwidget.ButtonEvent:
			if e.Type == basicwidget.ButtonEventTypeUp {
				r.tryCreateTask()
			}
		}
	}
	for event := range guigui.DequeueEvents(&r.textField) {
		switch e := event.(type) {
		case basicwidget.TextEvent:
			if e.Type == basicwidget.TextEventTypeEnterPressed {
				r.tryCreateTask()
			}
		}
	}
	for id, t := range r.taskWidgets {
		for event := range guigui.DequeueEvents(&t.doneButton) {
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
		guigui.Enable(&r.createButton)
	} else {
		guigui.Disable(&r.createButton)
	}

	return nil
}

func (r *Root) canCreateTask() bool {
	str := r.textField.Text()
	str = strings.TrimSpace(str)
	return str != ""
}

func (r *Root) tryCreateTask() {
	str := r.textField.Text()
	str = strings.TrimSpace(str)
	if str != "" {
		r.tasks = slices.Insert(r.tasks, 0, NewTask(str))
		r.textField.SetText("")
	}
}

func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:           "TODO",
		WindowMinWidth:  320,
		WindowMinHeight: 240,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
