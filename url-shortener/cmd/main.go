// Package cmd...
package cmd

import (
	"context"
	"log"
	"os"

	"github.com/cmmasaba/prototypes/telemetry"
)

func StartApplication(ctx context.Context) {
	version := os.Getenv("VERSION")
	environment := os.Getenv("ENVIRONMENT")
	serviceName := "url-shortener"

	otelClient := telemetry.Client{
		OTLPBaseURL: os.Getenv("OTLP_BASE_URL"),
		ServiceName: serviceName,
		Environment: environment,
		Version:     version,
	}

	otelShutdownFuncs, err := telemetry.NewOtelSDK(ctx, &otelClient)
	if err != nil {
		log.Printf("failed to setup opentelemetry sdk: %v", err)
	}

	defer func() {
		if otelShutdownFuncs == nil {
			return
		}

		otelErr := otelShutdownFuncs(ctx)

		log.Println(otelErr)
	}()
}
