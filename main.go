package main

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	// Example line:
	// # 2006-01-02
	timeLogSectionStart = regexp.MustCompile("^# ([[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2})$")
	// Example line:
	// 1100: Daily meeting
	timeLogSectionLine = regexp.MustCompile("^([[:digit:]]{3,4}): (.+)$")
)

type TimeLogSectionLine struct {
	Start       string
	EventsSplit []string
}

type TimeLogSection struct {
	Day         string
	EventStarts []TimeLogSectionLine
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

func toDuration(t string) (time.Duration, error) {
	if len(t) == 3 {
		// 'hmm' to 'hhmm' (ex: 950 -> 0950)
		t = "0" + t
	} else if len(t) != 4 {
		return 0, errors.New("string must match 'hmm' or 'hhmm' format")
	}
	duration, err := time.ParseDuration(t[0:2] + "h" + t[2:4] + "m")
	if err != nil {
		return 0, errors.New("string must match 'hmm' or 'hhmm' format")
	}
	return duration, nil
}

func largerSumFirst(events []*Event) func(i, j int) bool {
	return func(i, j int) bool {
		if events[i].Sum == events[j].Sum {
			return events[i].Description < events[j].Description
		}
		return events[i].Sum > events[j].Sum
	}
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

func (ts TimeLogSection) GetEventTree() (*Event, error) {
	mainEvent := NewEvent("", 0)
	for i := 0; i < len(ts.EventStarts)-1; i++ {
		start := ts.EventStarts[i].Start
		startTime, err := toDuration(start)
		if err != nil {
			return nil, err
		}
		end := ts.EventStarts[i+1].Start
		endTime, err := toDuration(end)
		if err != nil {
			return nil, err
		}
		// >= (instead of >) means we also ignore 0m events.
		if startTime >= endTime {
			return nil, errors.New("start time cannot be after end time")
		}
		mainEvent.Aggregate(ts.EventStarts[i].EventsSplit, endTime-startTime)
	}
	return mainEvent, nil
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

func splitEvents(line string) []string {
	return strings.Split(line, " - ")
}

func processTimeLog(timeLogContent io.Reader) ([]TimeLogSection, error) {
	timeLogSections := make([]TimeLogSection, 0)
	scanner := bufio.NewScanner(timeLogContent)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		matches := timeLogSectionStart.FindStringSubmatch(line)
		if len(matches) == 2 {
			logger.Debug("matched start of day section with date: " + matches[1])
			timeLogSections = append(timeLogSections, TimeLogSection{
				Day:         matches[1],
				EventStarts: make([]TimeLogSectionLine, 0),
			})
			continue
		}
		matches = timeLogSectionLine.FindStringSubmatch(line)
		if len(matches) == 3 {
			logger.Debugf("matched line in day section with time: %s, event: %s", matches[1], matches[2])
			if len(timeLogSections) == 0 {
				return nil, errors.New("encountered a section line before encountering the start of a section")
			}
			currentDayEventStarts := timeLogSections[len(timeLogSections)-1].EventStarts
			currentDayEventStarts = append(currentDayEventStarts, TimeLogSectionLine{
				Start:       matches[1],
				EventsSplit: splitEvents(matches[2]),
			})
			timeLogSections[len(timeLogSections)-1].EventStarts = currentDayEventStarts
			continue
		}
		if line != "" {
			logger.Debug("non-empty non-timelog line ignored: " + line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return timeLogSections, nil
}

func outputTimeLogAggregation(timeLogSections []TimeLogSection, contentOutput io.Writer) error {
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

var logger *zap.SugaredLogger

func main() {
	devLoggerConfig := zap.NewDevelopmentConfig()
	devLoggerConfig.Level.SetLevel(zap.InfoLevel)
	devLogger := zap.Must(devLoggerConfig.Build())
	defer devLogger.Sync()
	logger = devLogger.Sugar()

	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		logger.Fatalf("%+v", err)
	}
	err = outputTimeLogAggregation(timeLogSections, os.Stdout)
	if err != nil {
		logger.Fatalf("%+v", err)
	}
}
