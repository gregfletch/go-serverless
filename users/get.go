package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gregfletch/go-serverless/common"
	"github.com/gregfletch/go-serverless/models"
	"os"
	"time"
)

type UserQuery struct {
	Id string `json:"Id"`
}

type UserResponse struct {
	User models.User `json:"user"`
}

func GetUserFromDynamoDb(id string) (*dynamodb.GetItemOutput, error, time.Duration) {
	start := time.Now()
	svc := dynamodb.New(session.New())

	userQuery := UserQuery{
		Id: id,
	}

	av, err := dynamodbattribute.MarshalMap(userQuery)
	if err != nil {
		return nil, err, time.Since(start)
	}

	ddbInput := &dynamodb.GetItemInput{
		Key:       av,
		TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
	}
	result, err := svc.GetItem(ddbInput)

	if err != nil {
		return nil, err, time.Since(start)
	}

	return result, err, time.Since(start)
}

func GetHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	start := time.Now()

	lambdaLogger := common.GetLogger(ctx, req)
	lambdaLogger.Info().Msg("Starting get user API.")

	id := req.PathParameters["id"]

	lambdaLogger.Info().Msg("Looking up user.")

	result, err, elapsed := GetUserFromDynamoDb(id)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error querying for user.")
		return common.RespondWithError(err)
	}
	lambdaLogger.Info().
		Dur("elapsed", elapsed).
		Str("operation", "read").
		Str("source", "DynamoDB").
		Msg("Successfully queried for user.")

	if len(result.Item) == 0 {
		lambdaLogger.Warn().Msg("No users found with the given ID.")
	}

	user := models.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		lambdaLogger.Error().Err(err).Msg("Error unmarshalling user object")
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	var buf bytes.Buffer
	body, err := json.Marshal(UserResponse{
		User: user,
	})

	json.HTMLEscape(&buf, body)

	lambdaLogger.Info().Dur("elapsed", time.Since(start)).Msg("Completed get user API.")

	resp := events.APIGatewayProxyResponse{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "users-get-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(GetHandler)
}
