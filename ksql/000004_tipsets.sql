CREATE STREAM {APP_PREFIX}_TIPSETS_STREAM (
  "height" STRING,
  "parents" STRING,
  "parent_weight" STRING,
  "parent_state" VARCHAR,
  "blocks" STRING,
  "min_timestamp" BIGINT,
  "state" INT
) WITH (kafka_topic='{APP_PREFIX}_TIPSETS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_TIPSETS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM {APP_PREFIX}_TIPSETS_STREAM EMIT CHANGES;