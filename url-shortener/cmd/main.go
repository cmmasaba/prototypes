// Package cmd...
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation"
)

// StartApplication starts the server.
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
		otelShutdownFuncs = nil

		log.Printf("error setting up otel sdk: %v", err)
	}

	defer func() {
		if otelShutdownFuncs == nil {
			return
		}

		otelErr := otelShutdownFuncs(ctx)

		log.Printf("error shutting down otel: %v", otelErr)
	}()

	pyroscopeEndpoint := os.Getenv("PYROSCOPE_ENDPOINT")

	profiler, err := telemetry.StartProfiler(pyroscopeEndpoint, serviceName, environment)
	if err != nil {
		profiler = nil

		log.Printf("error starting pyroscope profiling: %v", err)
	}

	defer func() {
		if profiler == nil {
			return
		}

		err := profiler.Stop()
		if err != nil {
			log.Printf("error stopping pyroscope profiler: %v", err)
		}
	}()

	handler, err := presentation.PrepareServer()
	if err != nil {
		log.Printf("error starting server: %v", err)

		return
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Printf("error reading port: %v", err)

		port = 8080
	}

	httpServer := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		log.Printf("server listening on %s", httpServer.Addr)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("error listening for requests: %v", err)
		}
	})

	wg.Go(func() {
		<-ctx.Done()

		ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
		defer cancelFunc()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("error shutting down server: %v", err)
		}
	})

	wg.Wait()
}
