# Lambdalytics

Replicate the Google Analytics platform on top of AWS.

## Status

Work in progress

- [x] Highly available `/collect` endpoint.
- [x] Stream, validate, and data from Kinesis into DynamoDB.
- [x] Archive data from DynamoDB into S3 access in Athena or Presto.
- [ ] Replicate S3 objects cross region for availability.
- [ ] Infra setup for Athena
- [ ] Default dashboards powered by ....

## Presto

I found this [blog post][] to be useful.
There, they talk about how data should be stored in S3 to make it compatible with the Hive connector.
Since we won't be using Presto to write the data for us, we need to be aware of the format.

[blog post]: https://joshua-robinson.medium.com/a-presto-data-pipeline-with-s3-b04009aec3d9

### S3 Key Form

```
[path_prefix/][school=west[/...]]/segment
```

Using this, we can augment our thoughts and align our implementation.
This will keep fields that are sent together, stored together.
It'll also make defining hit-type specific tables easy to do.

```
hit_type=event/date=2006-01-02/01
hit_type=event/date=2006-01-02/02
hit_type=event/date=2006-01-02/03
```

### Examples

Here's a definition for a table only for events. 

```sql
CREATE TABLE event_analytics (
    event_category varchar,
    event_action varchar,
    event_label varchar,
    event_value int,
    date varchar,
) WITH (
    format = 'json', 
    external_location = 's3a://bucket_name/hit_type=event/',
    partitioned_by = ARRAY['date'],
);
```

Here's a generalized definition for a table that stores all.

```sql
CREATE TABLE event_analytics (
    -- ...
    hit_type varchar,
    date varchar,
) WITH (
    format = 'json', 
    external_location = 's3a://bucket_name/',
    partitioned_by = ARRAY['hit_type', 'date'],
);
```
