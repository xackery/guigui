// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"slices"
)

type Event any

type EventQueue struct {
	events []Event
}

func (e *EventQueue) Enqueue(arg Event) {
	e.events = append(e.events, arg)
}

func (e *EventQueue) Dequeue() (Event, bool) {
	if len(e.events) == 0 {
		return nil, false
	}
	event := e.events[0]
	e.events = slices.Delete(e.events, 0, 1)
	return event, true
}

func (e *EventQueue) Clear() {
	e.events = slices.Delete(e.events, 0, len(e.events))
}
