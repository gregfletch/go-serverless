package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
	"strings"
	"time"

	"github.com/akamensky/base58"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type UserInput struct {
	Address     string `json:"address"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phone"`
}

type User struct {
	Address     string `json:"address"`
	CreatedAt   string `json:"createdAt"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	Id          string `json:"Id"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phone"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

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
		Logger()
}

func WriteToS3(req events.APIGatewayProxyRequest, user UserInput) (*s3.PutObjectOutput, error) {
	svc := s3.New(session.New())
	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(req.Body)),
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(user.LastName + "," + user.FirstName + ".json"),
	}
	return svc.PutObject(input)
}

func RespondWithError(err error) (Response, error) {
	return Response{StatusCode: 400}, err
}

func GenerateId(randBytes []byte) string {
	userid := base58.Encode(randBytes)
	if len(userid) > 16 {
		userid = userid[0:16]
	}
	return "u_" + userid
}

func WriteToDynamoDB(userInput UserInput, userid string) error {
	svc := dynamodb.New(session.New())
	user := User{
		Address:     userInput.Address,
		CreatedAt:   time.Now().String(),
		Email:       userInput.Email,
		FirstName:   userInput.FirstName,
		Id:          userid,
		LastName:    userInput.LastName,
		PhoneNumber: userInput.PhoneNumber,
		UpdatedAt:   time.Now().String(),
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return err
	}

	ddbInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
	}

	_, err = svc.PutItem(ddbInput)
	if err != nil {
		return err
	}

	return nil
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (Response, error) {
	start := time.Now()

	lambdaLogger := GetLogger(ctx, req)
	lambdaLogger.Info().Msg("Starting Hello Lambda.")

	var userInput UserInput
	err := json.Unmarshal([]byte(req.Body), &userInput)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error unmarshaling data from request.")
		return RespondWithError(err)
	}

	result, err := WriteToS3(req, userInput)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error writing to S3.")
		return RespondWithError(err)
	}
	lambdaLogger.Info().
		Dur("elapsed", time.Since(start)).
		Str("result", result.String()).
		Str("operation", "write").
		Str("source", "s3").
		Msg("Raw payload written to S3.")

	randBytes := make([]byte, 12)
	_, err = rand.Read(randBytes)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error generating random bytes.")
		return RespondWithError(err)
	}

	userid := GenerateId(randBytes)
	lambdaLogger.Info().Str("userId", userid).Msg("Created user ID.")

	err = WriteToDynamoDB(userInput, userid)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error writing user to DynamoDB.")
		return RespondWithError(err)
	}

	lambdaLogger.Info().
		Dur("elapsed", time.Since(start)).
		Str("result", result.String()).
		Str("operation", "write").
		Str("source", "DynamoDB").
		Msg("User added to DynamoDB successfully.")

	var buf bytes.Buffer

	body, err := json.Marshal(CreateResponse{
		Id:      userid,
		Message: "User created successfully!",
	})
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error creating user.")
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	lambdaLogger.Info().Dur("elapsed", time.Since(start)).Msg("Completed create user API.")

	resp := Response{
		StatusCode:      201,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "users-create-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
