package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func printTimeclockFormat(timeLogSections []TimeLogSection, contentOutput io.Writer) error {
	for _, section := range timeLogSections {
		for i := 0; i < len(section.EventStarts)-1; i++ {
			start := toHHMM(section.EventStarts[i].Start)
			end := toHHMM(section.EventStarts[i+1].Start)
			accounts := strings.Join(section.EventStarts[i].EventsSplit, ":")
			fmt.Fprintf(contentOutput, "i %s %s %s\n", section.Day, start, accounts)
			fmt.Fprintf(contentOutput, "o %s %s\n", section.Day, end)
		}
	}
	return nil
}

func toHHMM(dur time.Duration) string {
	// %02d ensures we're padding with a leading zero when the number is less than 10.
	return fmt.Sprintf("%02d:%02d", dur/time.Hour, dur/time.Minute%60)
}
