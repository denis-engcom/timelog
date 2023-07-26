package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"time"
)

type Summary struct {
	Day string        `json:"day"`
	Sum time.Duration `json:"sum"`
}

func (s *Summary) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Day string `json:"day"`
		Sum string `json:"sum"`
	}{
		Day: s.Day,
		Sum: dropTrailingZeros(s.Sum),
	})
}

func getSummaryWithoutHLedger(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	daysMap := map[string]time.Duration{}
	for _, ts := range timeLogSections {
		// daysMap[ts.Day] = daysMap[ts.Day] + ts.EventStarts
		for i := 0; i < len(ts.EventStarts)-1; i++ {
			start := ts.EventStarts[i].Start
			end := ts.EventStarts[i+1].Start
			// >= (instead of >) means we also ignore 0m events.
			if start >= end {
				return errors.New("start time cannot be after end time")
			}
			daysMap[ts.Day] = daysMap[ts.Day] + (end - start)
		}
	}
	days := make([]Summary, 0)
	for day, sum := range daysMap {
		days = append(days, Summary{Day: day, Sum: sum})
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Day < days[j].Day
	})
	for _, summary := range days {
		fmt.Printf("%s: %s\n", summary.Day, dropTrailingZeros(summary.Sum))
	}

	// summaryBytes, err := json.Marshal(&days)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("%s\n", summaryBytes)

	return nil
}
