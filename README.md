# Lambdalytics

Replicate the Google Analytics platform on top of AWS.

## Status

Proof of concept that I'm mostly doing for fun.
With the way Google as an organization has been going, I've thought about all the analytics trackers out there and how we might replace them.
I've also been thinking a bit about the analytics system we had at Indeed and the similarities between the two.
I might generalize this more later on, but for now it's mostly an idea.

- [x] Highly available `/collect` endpoint.
- [x] Stream and filter data from Kinesis into DynamoDB.
- [ ] Issue and validate tracking ids.
- [ ] Enrich data with additional metadata about the caller.
- [x] Archive data from DynamoDB into S3 access in Athena or Presto.
- [ ] Replicate S3 objects cross region.
- [ ] Infra setup for Athena.
- [ ] Default dashboards powered by ....

## Presto

I found this [blog post][] to be useful.
There, they talk about how data should be stored in S3 to make it compatible with the Hive connector.
Since we won't be using Presto to write the data for us, we need to be aware of the format.

[blog post]: https://joshua-robinson.medium.com/a-presto-data-pipeline-with-s3-b04009aec3d9

### S3 Key Form

```
[path_prefix/][partitioned_by=value[/...]]/segment
```

Using this, we can augment our thoughts and align our implementation.
This will keep fields that are sent together, stored together.
It'll also make defining type specific tables easy to do.

```
type=event/date=2006-01-02/01
type=event/date=2006-01-02/02
type=event/date=2006-01-02/03
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
    external_location = 's3a://bucket_name/type=event/',
    partitioned_by = ARRAY['date'],
);
```

Here's a generalized definition for a table that stores all.

```sql
CREATE TABLE event_analytics (
    -- ...
    type varchar,
    date varchar,
) WITH (
    format = 'json', 
    external_location = 's3a://bucket_name/',
    partitioned_by = ARRAY['type', 'date'],
);
```
