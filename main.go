package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// TODO accept parsing configuration via env vars to override " - " for splitting lines.
// Use caarlos0/env or urfave/cli env capabilities.
//"https://github.com/caarlos0/env"
//type Config struct {
//	TimelogLineSeparator string `env:"TIMELOG_LINE_SEPARATOR"`
//}

var (
	// Example line:
	// # 2006-01-02
	regexTimeLogSectionStart = regexp.MustCompile("^# ([[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2})$")
	// Example line:
	// 11:00 Daily meeting
	regexTimeLogSectionLine = regexp.MustCompile("^([[:digit:]:]+)[[:blank:]]+(.+)$")
	// 11:00 or 9:00
	regexHHColonMM = regexp.MustCompile("^([[:digit:]]{1,2}):([[:digit:]]{2})$")
	// 1100
	regexHHMM = regexp.MustCompile("^([[:digit:]]{2})([[:digit:]]{2})$")
	// 900
	regexHMM = regexp.MustCompile("^([[:digit:]])([[:digit:]]{2})$")
)

var logger *zap.SugaredLogger

func main() {
	devLoggerConfig := zap.NewDevelopmentConfig()
	devLoggerConfig.Level.SetLevel(zap.InfoLevel)
	devLogger := zap.Must(devLoggerConfig.Build())
	defer devLogger.Sync()
	logger = devLogger.Sugar()

	app := &cli.App{
		Name:  "timelog",
		Usage: "Processes timelog input and generates aggregated event information.",
		Description: `The command processes timelog input from stdin and outputs data in the following formats:
- Timelog output format: our own per-day tree-like event aggregation
- Timeclock (hledger) format: for subsequent processing with hledger (https://hledger.org/)

The timelog output format is ideal for minimal configuration. The bullet-point tree is easy to reason about. However, for the time being, the output cannot be further processed.

The timeclock format is verbose and intended to be processed by hledger into reports with many types of aggregations and filtering.

TODO provide hledger examples.`,
		UsageText: `timelog [global options] (stdin)
timelog < 2023-01_timelog.md > 2023-01_timelog_aggregated.md
timelog -O timeclock < 2023-01_timelog.md > 2023-01_timelog.timeclock
timelog -O timeclock < 2023-01_timelog.md | hledger -ftimeclock:- register --daily > 2023-01_timelog_reports_register.txt`,
		Version:              "0.3.0",
		HideHelpCommand:      true,
		ArgsUsage:            "(stdin)",
		Action:               timelog,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output-format",
				Aliases: []string{"O"},
				Usage:   "Choose output format, choices: (timelog|timeclock)",
				Value:   "timeclock",
			},
		},
		Commands: cli.Commands{
			{
				Name:    "summary",
				Aliases: []string{"sum"},
				Usage:   "Print timeclock entries aggregated by day",
				Action:  summary,
			},
			{
				Name:    "register",
				Aliases: []string{"reg"},
				Usage:   "Print timeclock entries in modified \"hledger register\" format",
				Action:  register,
			},
			{
				Name:    "print",
				Aliases: []string{"p"},
				Usage:   "Print timeclock entries in collapsed \"hledger print\" format",
				Action:  print,
			},
			{
				Name:    "balance",
				Aliases: []string{"b"},
				Usage:   "Print timeclock entries in \"hledger balance\" format",
				Action:  balance,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatalf("%+v", err)
	}
}

var (
	TIMECLOCK_FORMAT = "-ftimeclock:-"
	NO_BREAK         = "not:acct:^Break"
	NO_LUNCH         = "not:acct:^Lunch"
)

func summary(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	var timeclockEntries bytes.Buffer
	err = printTimeclockFormat(timeLogSections, &timeclockEntries)
	if err != nil {
		return err
	}
	cmd := exec.Command("hledger", TIMECLOCK_FORMAT, "register", NO_BREAK, NO_LUNCH, "--daily", "--pivot", "date")
	cmd.Stdin = &timeclockEntries
	var hledgerRegister bytes.Buffer
	cmd.Stdout = &hledgerRegister
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// TODO easy to covert to using go get github.com/cheynewallace/tabby?
	cmd = exec.Command("awk", "{ printf(\"%s: %s (%s)\\n\", $1, $2, $3) }")
	cmd.Stdin = &hledgerRegister
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func register(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	var timeclockEntries bytes.Buffer
	err = printTimeclockFormat(timeLogSections, &timeclockEntries)
	if err != nil {
		return err
	}
	args := []string{TIMECLOCK_FORMAT, "register", NO_BREAK, NO_LUNCH}
	period := cCtx.Args().First()
	if period != "" {
		args = append(args, "-p", period)
	}
	cmd := exec.Command("hledger", args...)
	cmd.Stdin = &timeclockEntries
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func print(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	var timeclockEntries bytes.Buffer
	err = printTimeclockFormat(timeLogSections, &timeclockEntries)
	if err != nil {
		return err
	}
	args := []string{TIMECLOCK_FORMAT, "print", NO_BREAK, NO_LUNCH}
	period := cCtx.Args().First()
	if period != "" {
		args = append(args, "-p", period)
	}
	cmd := exec.Command("hledger", args...)
	cmd.Stdin = &timeclockEntries
	var hledgerPrint bytes.Buffer
	cmd.Stdout = &hledgerPrint
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("paste", "-d", "\\0", "-", "-", "-")
	cmd.Stdin = &hledgerPrint
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func balance(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	var timeclockEntries bytes.Buffer
	err = printTimeclockFormat(timeLogSections, &timeclockEntries)
	if err != nil {
		return err
	}
	args := []string{TIMECLOCK_FORMAT, "balance", NO_BREAK, NO_LUNCH, "--tree", "--sort-amount"}
	period := cCtx.Args().First()
	if period != "" {
		args = append(args, "-p", period)
	}
	cmd := exec.Command("hledger", args...)
	cmd.Stdin = &timeclockEntries
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func timelog(cCtx *cli.Context) error {
	timeLogSections, err := processTimeLog(os.Stdin)
	if err != nil {
		return err
	}
	switch cCtx.String("output-format") {
	case "timelog":
		err = printTimelogFormat(timeLogSections, os.Stdout)
		if err != nil {
			return err
		}
	case "timeclock":
		err = printTimeclockFormat(timeLogSections, os.Stdout)
		if err != nil {
			return err
		}
	default:
		return cli.ShowAppHelp(cCtx)
	}
	return nil
}

type TimeLogSection struct {
	Day         string
	EventStarts []TimeLogSectionLine
}

type TimeLogSectionLine struct {
	Start       time.Duration
	EventsSplit []string
}

func processTimeLog(timeLogContent io.Reader) ([]TimeLogSection, error) {
	timeLogSections := make([]TimeLogSection, 0)
	scanner := bufio.NewScanner(timeLogContent)
	scanner.Split(bufio.ScanLines)
	var lineNumber uint
	for scanner.Scan() {
		lineNumber += 1
		line := scanner.Text()
		matches := regexTimeLogSectionStart.FindStringSubmatch(line)
		if len(matches) == 2 {
			logger.Debugf("line %d: matched start of day section with date: %s", lineNumber, matches[1])
			timeLogSections = append(timeLogSections, TimeLogSection{
				Day:         matches[1],
				EventStarts: make([]TimeLogSectionLine, 0),
			})
			continue
		}
		matches = regexTimeLogSectionLine.FindStringSubmatch(line)
		if len(matches) == 3 {
			logger.Debugf("line %d: matched line in day section with time: %s, event: %s", lineNumber, matches[1], matches[2])
			if len(timeLogSections) == 0 {
				return nil, errors.Errorf(
					"error on line %d: %s, encountered a section line before encountering the start of a section", lineNumber, matches[0])
			}
			currentDayEventStarts := timeLogSections[len(timeLogSections)-1].EventStarts
			start, err := toDuration(matches[1])
			if err != nil {
				return nil, errors.Wrapf(err, "error on line %d: %s", lineNumber, matches[0])
			}
			currentDayEventStarts = append(currentDayEventStarts, TimeLogSectionLine{
				Start:       start,
				EventsSplit: splitEvents(matches[2]),
			})
			timeLogSections[len(timeLogSections)-1].EventStarts = currentDayEventStarts
			continue
		}
		if line != "" {
			logger.Debugf("line %d: non-empty non-timelog line ignored: %s", lineNumber, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return timeLogSections, nil
}

func toDuration(t string) (time.Duration, error) {
	var durationToParse string
	matches := regexHHColonMM.FindStringSubmatch(t)
	if len(matches) == 3 {
		durationToParse = fmt.Sprintf("%sh%sm", matches[1], matches[2])
	}
	matches = regexHHMM.FindStringSubmatch(t)
	if len(matches) == 3 {
		durationToParse = fmt.Sprintf("%sh%sm", matches[1], matches[2])
	}
	matches = regexHMM.FindStringSubmatch(t)
	if len(matches) == 3 {
		durationToParse = fmt.Sprintf("%sh%sm", matches[1], matches[2])
	}
	duration, err := time.ParseDuration(durationToParse)
	if err != nil {
		return 0, errors.Errorf("error parsing %s: string must match 'h:mm' or 'hh:mm', 'hmm' or 'hhmm' format", t)
	}
	return duration, nil
}

func splitEvents(line string) []string {
	return strings.Split(line, " - ")
}
