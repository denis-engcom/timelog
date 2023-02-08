package main

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"sort"
	"time"
)

func printTimelogFormat(timeLogSections []TimeLogSection, contentOutput io.Writer) error {
	for _, section := range timeLogSections {
		event, err := section.GetEventTree()
		if err != nil {
			return errors.WithStack(err)
		}
		fmt.Fprintf(contentOutput, "# %s\n\n", section.Day)
		event.PrintTreeSorted(contentOutput)
		fmt.Fprint(contentOutput, "\n")
	}
	return nil
}

type Event struct {
	Description string
	Sum         time.Duration
	Indent      uint
	SubEvents   map[string]*Event
}

func NewEvent(description string, indent uint) *Event {
	return &Event{
		Description: description,
		Sum:         0,
		Indent:      indent,
		SubEvents:   make(map[string]*Event, 0),
	}
}

func (ts TimeLogSection) GetEventTree() (*Event, error) {
	mainEvent := NewEvent("", 0)
	for i := 0; i < len(ts.EventStarts)-1; i++ {
		start := ts.EventStarts[i].Start
		end := ts.EventStarts[i+1].Start
		// >= (instead of >) means we also ignore 0m events.
		if start >= end {
			return nil, errors.New("start time cannot be after end time")
		}
		mainEvent.Aggregate(ts.EventStarts[i].EventsSplit, end-start)
	}
	return mainEvent, nil
}

func (e *Event) Aggregate(eventsSplit []string, duration time.Duration) {
	e.Sum = e.Sum + duration
	if len(eventsSplit) == 0 {
		return
	}

	description := eventsSplit[0]
	_, ok := e.SubEvents[description]
	if !ok {
		e.SubEvents[description] = NewEvent(description, e.Indent+1)
	}
	e.SubEvents[description].Aggregate(eventsSplit[1:], duration)
}

func (e *Event) PrintTreeSorted(contentOutput io.Writer) {
	events := make([]*Event, 0, len(e.SubEvents))
	for _, ev := range e.SubEvents {
		events = append(events, ev)
	}
	sort.Slice(events, largerSumFirst(events))
	for i := uint(0); i < e.Indent; i++ {
		fmt.Fprintf(contentOutput, "\t")
	}
	fmt.Fprintf(contentOutput, "- %s", dropTrailingZeros(e.Sum))
	if e.Description != "" {
		fmt.Fprintf(contentOutput, ": %s", e.Description)
	}
	fmt.Fprint(contentOutput, "\n")
	for _, ev := range events {
		ev.PrintTreeSorted(contentOutput)
	}
}

func largerSumFirst(events []*Event) func(i, j int) bool {
	return func(i, j int) bool {
		if events[i].Sum == events[j].Sum {
			return events[i].Description < events[j].Description
		}
		return events[i].Sum > events[j].Sum
	}
}

// Drop "0s" and "0m" from the end of default duration formatting.
func dropTrailingZeros(d time.Duration) string {
	dStr := d.String()
	// Small exception: the zero value "0s" should be returned as-is.
	if dStr == "0s" {
		return dStr
	}
	if len(dStr) >= 3 && dStr[len(dStr)-3:] == "m0s" {
		dStr = dStr[:len(dStr)-2]
	}
	if len(dStr) >= 3 && dStr[len(dStr)-3:] == "h0m" {
		dStr = dStr[:len(dStr)-2]
	}
	return dStr
}
