package common

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func GetLogger(ctx context.Context, req events.APIGatewayProxyRequest) zerolog.Logger {
	if os.Getenv("PRETTY_LOGS") == "true" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	lambdactx, _ := lambdacontext.FromContext(ctx)
	return log.With().
		Str("requestId", lambdactx.AwsRequestID).
		Str("path", req.Path).
		Str("method", req.HTTPMethod).
		Str("id", req.PathParameters["id"]).
		Logger()
}

func RespondWithError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: 400}, err
}
