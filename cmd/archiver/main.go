package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	dynamoattr "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mjpitz/simple-analytics/internal/data"
	"github.com/mjpitz/simple-analytics/internal/models"
)

var (
	HandlerErr = fmt.Errorf("handler failed")
)

type ScheduleEvent struct {
	Time time.Time `json:"time,omitempty"`
}

type Server struct {
	dynamoClient *dynamodb.DynamoDB
	s3Client     *s3.S3
}

// runs every hour @30min, archiving data from the hour before.
func (s *Server) Handler(ctx context.Context, request ScheduleEvent) error {
	l := request.Time.Location()
	year, month, day := request.Time.Date()
	end := request.Time.Hour()
	start := end - 1

	startTime := time.Date(year, month, day, start, 0, 0, 0, l)
	endTime := time.Date(year, month, day, end, 0, 0, 0, l)

	// subtract one since query is exclusive on the start key
	startPartition := startTime.Unix() - 1
	endPartition := endTime.Unix()

	startRecord := &models.Record{ PartitionKey: startPartition, RangeKey: "" }
	startKey, err := dynamoattr.MarshalMap(startRecord)
	if err != nil {
		return HandlerErr
	}

	date := startTime.Format("2006-01-02")
	segment := startTime.Format("15")

	idx := make(map[string][]*models.Record, endPartition - startPartition)

	for ptr := startRecord; ptr.PartitionKey < endPartition; {
		q, err := s.dynamoClient.Query(&dynamodb.QueryInput{
			ExclusiveStartKey: startKey,
		})
		if err != nil {
			return HandlerErr
		}

		for _, item := range q.Items {
			record := &models.Record{}
			err = dynamoattr.UnmarshalMap(item, record)
			if err != nil {
				return HandlerErr
			}

			// partitioned by date
			k := fmt.Sprintf("hit_type=%s/date=%s/%s",
				record.HitType, date, segment)

			idx[k] = append(idx[k], record)
		}

		startKey = q.LastEvaluatedKey
		err = dynamoattr.UnmarshalMap(q.LastEvaluatedKey, ptr)
		if err != nil {
			return HandlerErr
		}
	}

	for key, records := range idx {
		out := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(out)

		for _, record := range records {
			_ = encoder.Encode(record)
		}

		_, err = s.s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(data.DataBucket),
			Key: aws.String(key),
			Body: bytes.NewReader(out.Bytes()),
		})
		if err != nil {
			return HandlerErr
		}
	}

	return nil
}

func main() {
	sess := session.Must(session.NewSession())

	server := &Server{
		dynamoClient: dynamodb.New(sess),
		s3Client:     s3.New(sess),
	}

	lambda.Start(server.Handler)
}
