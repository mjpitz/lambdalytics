package data

import "os"

const (
	RequestTimestampHeader = "x-request-unixtime"
)

var (
	CollectionStream = os.Getenv("COLLECTION_STREAM_NAME")
	RecordTable = os.Getenv("RECORD_TABLE_NAME")
	DataBucket = os.Getenv("DATA_BUCKET_NAME")
)