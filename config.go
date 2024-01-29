package otel

import (
	"fmt"
	"os"
	"strconv"
)

type ExporterType int

const (
	stdout ExporterType = iota
	otlp
)

type Config struct {
	Enabled          bool         `mapstructure:"enabled"`
	Exporter         ExporterType `mapstructure:"exporter"`
	OtlpCollectorUrl string       `mapstructure:"collector_url"`
}

func (C *Config) Default() {

	C.Enabled = false
	C.Exporter = stdout
}

func (C *Config) LoadVarsFromEnv() {

	enabled, ok := os.LookupEnv("OTEL_ENABLED")
	if ok {

		var boolEn bool
		boolEn, err := strconv.ParseBool(enabled)
		if err != nil {
			fmt.Println("Error:", err)
			boolEn = false
		}

		C.Enabled = boolEn
	}

	exporter, ok := os.LookupEnv("OTEL_EXPORTER")
	if ok {

		var u int
		u, err := strconv.Atoi(exporter)
		if err != nil {
			fmt.Println("Error:", err)
			u = 0
		}

		C.Exporter = ExporterType(u)
	}

	url, ok := os.LookupEnv("OTEL_COLLECTOR_URL")
	if ok {
		C.OtlpCollectorUrl = url
	}
}
