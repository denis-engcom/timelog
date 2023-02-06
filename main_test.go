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

func TestProcessAndPrintTimeLogs(t *testing.T) {
	timeLogContent, err := os.Open("testdata/input.md")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	expectedAggregationOutput, err := os.ReadFile("testdata/output.md")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	timeLogSections, err := processTimeLog(timeLogContent)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	aggregationOutput := bytes.Buffer{}
	err = outputTimeLogAggregation(timeLogSections, &aggregationOutput)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Equal(t, string(expectedAggregationOutput), aggregationOutput.String())
}
