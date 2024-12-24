// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
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
	guigui.DefaultWidgetBehavior

	createButton       basicwidget.TextButton
	createButtonWidget *guigui.Widget
	textField          basicwidget.TextField
	textFieldWidget    *guigui.Widget
	taskWidgets        map[uuid.UUID]*TaskWidgets
	tasksPanel         basicwidget.ScrollablePanel
	tasksPanelWidget   *guigui.Widget

	tasks []Task
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if r.createButtonWidget == nil {
		r.createButton.SetText("Create")
		r.createButtonWidget = guigui.NewWidget(&r.createButton)
	}
	if r.textFieldWidget == nil {
		r.textFieldWidget = guigui.NewWidget(&r.textField)
	}

	taskItemsLinearGrid := &basicwidget.LinearGrid{
		Direction: basicwidget.LinearGridDirectionVertical,
	}
	for _, t := range r.tasks {
		if _, ok := r.taskWidgets[t.ID]; !ok {
			var b basicwidget.TextButton
			b.SetText("Done")
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
		if len(taskItemsLinearGrid.Items) > 0 {
			taskItemsLinearGrid.Items = append(taskItemsLinearGrid.Items, basicwidget.LinearGridItem{
				Size: 0.25,
			})
		}
		taskItemsLinearGrid.Items = append(taskItemsLinearGrid.Items, basicwidget.LinearGridItem{
			Widget: guigui.NewWidget(&basicwidget.LinearGrid{
				Direction: basicwidget.LinearGridDirectionHorizontal,
				Items: []basicwidget.LinearGridItem{
					{
						Widget: r.taskWidgets[t.ID].doneButtonWidget,
						Size:   3,
					},
					{
						Size: 0.5,
					},
					{
						Widget:   r.taskWidgets[t.ID].textWidget,
						Size:     1,
						SizeUnit: basicwidget.SizeUnitFraction,
					},
				},
			}),
			Size: 1,
		})
	}

	if r.tasksPanelWidget == nil {
		r.tasksPanelWidget = guigui.NewWidget(&r.tasksPanel)
	}
	r.tasksPanel.SetContent(guigui.NewWidget(&basicwidget.LinearGrid{
		Items: []basicwidget.LinearGridItem{
			{
				Widget:        guigui.NewWidget(taskItemsLinearGrid),
				PaddingLeft:   0.5,
				PaddingRight:  0.5,
				PaddingTop:    0.5,
				PaddingBottom: 0.5,
				Size:          1,
				SizeUnit:      basicwidget.SizeUnitFraction,
			},
		},
	}))
	r.tasksPanel.SetContentSize(widget.Bounds().Dx(), taskItemsLinearGrid.MinimumSize(context)+1*basicwidget.UnitSize(context))

	c := guigui.NewWidget(&basicwidget.LinearGrid{
		Direction: basicwidget.LinearGridDirectionVertical,
		Items: []basicwidget.LinearGridItem{
			{
				Widget: guigui.NewWidget(&basicwidget.LinearGrid{
					Direction: basicwidget.LinearGridDirectionHorizontal,
					Items: []basicwidget.LinearGridItem{
						{
							Widget:   r.textFieldWidget,
							Size:     1,
							SizeUnit: basicwidget.SizeUnitFraction,
						},
						{
							Size: 0.5,
						},
						{
							Widget: r.createButtonWidget,
							Size:   4,
						},
					},
				}),
				Size:          2,
				PaddingLeft:   0.5,
				PaddingRight:  0.5,
				PaddingTop:    0.5,
				PaddingBottom: 0.5,
			},
			{
				Widget:   r.tasksPanelWidget,
				Size:     1,
				SizeUnit: basicwidget.SizeUnitFraction,
			},
		},
	})
	appender.AppendChildWidget(c, widget.Bounds())

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

func (r *Root) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title: "TODO",
	}
	// scheme.ColorMode = basicwidget.ColorModeDark
	if err := guigui.Run(guigui.NewWidget(&Root{}), op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
