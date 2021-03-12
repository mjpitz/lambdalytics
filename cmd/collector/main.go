package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/go-uuid"
	"github.com/mjpitz/simple-analytics/internal/data"
)

var (
	OK = events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/html",
			"Content-Length": "0",
		},
		Body: "",
	}
	InternalError = events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Content-Type": "text/html",
			"Content-Length": "0",
		},
		Body: "",
	}
)

type Server struct {
	kinesisClient *kinesis.Kinesis
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (s *Server) Handler(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	request.Headers[data.RequestTimestampHeader] = fmt.Sprintf("%d", time.Now().Unix())

	payload, err := json.Marshal(request)
	if err != nil {
		return InternalError
	}

	partitionKey, _ := uuid.GenerateUUID()

	_, err = s.kinesisClient.PutRecordWithContext(ctx, &kinesis.PutRecordInput{
		StreamName: aws.String(data.CollectionStream),
		PartitionKey: aws.String(partitionKey),
		Data: payload,
	})
	if err != nil {
		return InternalError
	}

	return OK
}

func main() {
	// load configuration from environment variables
	sess := session.Must(session.NewSession())

	server := &Server{
		kinesisClient: kinesis.New(sess),
	}

	lambda.Start(server.Handler)
}
