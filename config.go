package otel

import (
	"fmt"
	"go.opentelemetry.io/otel/sdk/trace"
	"os"
	"strconv"
)

type ExporterType int

const (
	stdout ExporterType = iota
	file
	otlp
)

type Config struct {
	tp               *trace.TracerProvider
	Enabled          bool         `mapstructure:"enabled"`
	Exporter         ExporterType `mapstructure:"exporter"`
	OtlpCollectorUrl string       `mapstructure:"collector_url"`
}

func Default() Config {
	return Config{
		Enabled:  false,
		Exporter: stdout,
	}
}

func (c *Config) LoadVarsFromEnv() {

	enabled, ok := os.LookupEnv("OTEL_ENABLED")
	if ok {

		var boolEn bool
		boolEn, err := strconv.ParseBool(enabled)
		if err != nil {
			fmt.Println("Error:", err)
			boolEn = false
		}

		c.Enabled = boolEn
	}

	exporter, ok := os.LookupEnv("OTEL_EXPORTER")
	if ok {

		var u int
		u, err := strconv.Atoi(exporter)
		if err != nil {
			fmt.Println("Error:", err)
			u = 0
		}

		c.Exporter = ExporterType(u)
	}

	url, ok := os.LookupEnv("OTEL_COLLECTOR_URL")
	if ok {
		c.OtlpCollectorUrl = url
	}
}
