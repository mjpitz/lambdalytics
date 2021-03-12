package models

import "github.com/mjpitz/go-ga/client/v1/gatypes"

type Record struct {
	PartitionKey int64  `json:"partition_key"`
	RangeKey     string `json:"range_key"`
	gatypes.Payload
}
