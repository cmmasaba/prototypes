// Package telemetry provides OTLP functionalities.
package telemetry

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

//nolint:gochecknoglobals
var validate = validator.New()

type Client struct {
	OTLPBaseURL string `json:"otlpBaseURL" validate:"required"`
	ServiceName string `json:"serviceName" validate:"required"`
	Environment string `json:"environment" validate:"required"`
	Version     string `json:"version"     validate:"required"`
}

// NewOtelSDK initializes the OpenTelemetry SDK and returns a shutdown
// function that the caller MUST invoke on application exit.
//
//nolint:nonamedreturns
func NewOtelSDK(
	ctx context.Context,
	client *Client,
) (shutdown func(context.Context) error, err error) {
	if client == nil {
		return nil, fmt.Errorf("[telemetry] client must not be nil") //nolint: err113
	}

	err = validate.Struct(client)
	if err != nil {
		return nil, fmt.Errorf("[telemetry] error validating client struct: %w", err)
	}

	return client.setupOtelSDK(ctx)
}
