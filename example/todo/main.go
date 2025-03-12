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
	guigui.RootWidget

	createButton basicwidget.TextButton
	textField    basicwidget.TextField
	taskWidgets  map[uuid.UUID]*TaskWidgets
	tasksPanel   basicwidget.ScrollablePanel

	tasks []Task
}

func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	width, _ := r.Size(context)
	w := width - int(6.5*u)
	r.textField.SetSize(context, w, int(u))
	r.textField.SetOnEnterPressed(func(text string) {
		r.tryCreateTask()
	})
	{
		guigui.SetPosition(&r.textField, guigui.Position(r).Add(image.Pt(int(0.5*u), int(0.5*u))))
		appender.AppendChildWidget(&r.textField)
	}

	r.createButton.SetText("Create")
	r.createButton.SetWidth(int(5 * u))
	r.createButton.SetOnUp(func() {
		r.tryCreateTask()
	})
	{
		p := guigui.Position(r)
		w, _ := r.Size(context)
		p.X += w - int(0.5*u) - int(5*u)
		p.Y += int(0.5 * u)
		guigui.SetPosition(&r.createButton, p)
		appender.AppendChildWidget(&r.createButton)
	}

	w, h := r.Size(context)
	r.tasksPanel.SetSize(context, w, h-int(2*u))
	r.tasksPanel.SetContent(func(context *guigui.Context, childAppender *basicwidget.ContainerChildWidgetAppender, offsetX, offsetY float64) {
		p := guigui.Position(&r.tasksPanel).Add(image.Pt(int(offsetX), int(offsetY)))
		minX := p.X + int(0.5*u)
		y := p.Y
		for i, t := range r.tasks {
			if _, ok := r.taskWidgets[t.ID]; !ok {
				var tw TaskWidgets
				tw.doneButton.SetText("Done")
				tw.doneButton.SetWidth(int(3 * u))
				tw.doneButton.SetOnUp(func() {
					r.tasks = slices.DeleteFunc(r.tasks, func(task Task) bool {
						return task.ID == t.ID
					})
				})
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
			guigui.SetPosition(&r.taskWidgets[t.ID].doneButton, image.Pt(minX, y))
			childAppender.AppendChildWidget(&r.taskWidgets[t.ID].doneButton)
			r.taskWidgets[t.ID].text.SetSize(w-int(4.5*u), int(u))
			guigui.SetPosition(&r.taskWidgets[t.ID].text, image.Pt(minX+int(3.5*u), y))
			childAppender.AppendChildWidget(&r.taskWidgets[t.ID].text)
			y += int(u)
		}
	})
	r.tasksPanel.SetPadding(0, int(0.5*u))
	guigui.SetPosition(&r.tasksPanel, guigui.Position(r).Add(image.Pt(0, int(2*u))))
	appender.AppendChildWidget(&r.tasksPanel)

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
	//basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:             "TODO",
		WindowMinWidth:    320,
		WindowMinHeight:   240,
		ScreenTransparent: true,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
