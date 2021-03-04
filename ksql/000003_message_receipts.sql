CREATE STREAM {APP_PREFIX}_MESSAGE_RECEIPTS_STREAM (
  "cid" VARCHAR,
  "gas_used" BIGINT,
  "exit_code" INTEGER,
  "return" STRING
) WITH (kafka_topic='{APP_PREFIX}_MESSAGE_RECEIPTS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_MESSAGE_RECEIPTS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM {APP_PREFIX}_MESSAGE_RECEIPTS_STREAM EMIT CHANGES;
