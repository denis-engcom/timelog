package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	devLoggerConfig := zap.NewDevelopmentConfig()
	devLoggerConfig.Level.SetLevel(zap.DebugLevel)
	devLogger := zap.Must(devLoggerConfig.Build())
	defer devLogger.Sync()
	logger = devLogger.Sugar()

	m.Run()
}

func TestProcessAndPrintTimelog(t *testing.T) {
	timeLogContent, err := os.Open("testdata/input.md")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	expectedAggregationOutput, err := os.ReadFile("testdata/output-timelog.md")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	timeLogSections, err := processTimeLog(timeLogContent)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	aggregationOutput := bytes.Buffer{}
	err = printTimelogFormat(timeLogSections, &aggregationOutput)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Equal(t, string(expectedAggregationOutput), aggregationOutput.String())
}

func TestProcessAndPrintTimeclock(t *testing.T) {
	timeLogContent, err := os.Open("testdata/input.md")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	expectedAggregationOutput, err := os.ReadFile("testdata/output-hledger.timeclock")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	timeLogSections, err := processTimeLog(timeLogContent)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	aggregationOutput := bytes.Buffer{}
	err = printTimeclockFormat(timeLogSections, &aggregationOutput)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Equal(t, string(expectedAggregationOutput), aggregationOutput.String())
}
