package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	dynamoattr "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	v1 "github.com/mjpitz/go-ga/client/v1"
	"github.com/mjpitz/go-ga/client/v1/gatypes"
	"github.com/mjpitz/simple-analytics/internal/data"
	"github.com/mjpitz/simple-analytics/internal/models"
)

var (
	HandlerErr = fmt.Errorf("handler failed")
)

type Server struct {
	dynamoClient *dynamodb.DynamoDB
}

func (s *Server) Convert(request *events.APIGatewayProxyRequest) *models.Record {
	values := make(url.Values)
	var err error

	if request.HTTPMethod == http.MethodPost {
		values, err = url.ParseQuery(request.Body)
		if err != nil {
			// log
			return nil
		}
	} else {
		for key, value := range request.QueryStringParameters {
			values.Set(key, value)
		}
	}

	payload := &gatypes.Payload{}
	err = v1.Decode(values, payload)
	if err != nil {
		// log
		return nil
	}

	_, err = v1.Values(payload)
	if err != nil {
		// log, payload missing required bit of information
		return nil
	}

	partitionKey, err := strconv.ParseInt(request.Headers[data.RequestTimestampHeader], 10, 64)
	if err != nil {
		return nil
	}

	rangeKey := fmt.Sprintf("%s/%s/%s",
		payload.TrackingID, payload.HitType, request.RequestContext.RequestID)

	return &models.Record{
		PartitionKey: partitionKey,
		RangeKey:     rangeKey,
		Payload:      *payload,
	}
}

func (s *Server) Handler(ctx context.Context, event events.KinesisEvent) error {
	requests := make([]*dynamodb.WriteRequest, 0, len(event.Records))

	for _, record := range event.Records {
		request := &events.APIGatewayProxyRequest{}
		err := json.Unmarshal(record.Kinesis.Data, &request)
		if err != nil {
			// log
			continue
		}

		record := s.Convert(request)
		if record == nil {
			continue
		}

		item, err := dynamoattr.MarshalMap(record)
		if err != nil {
			// log
			continue
		}

		requests = append(requests, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: item,
			},
		})
	}

	_, err := s.dynamoClient.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			data.RecordTable: requests,
		},
	})
	if err != nil {
		// log
		return HandlerErr
	}

	return nil
}

func main() {
	sess := session.Must(session.NewSession())

	server := &Server{
		dynamoClient: dynamodb.New(sess),
	}

	lambda.Start(server.Handler)
}
