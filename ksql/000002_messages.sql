CREATE STREAM {APP_PREFIX}_MESSAGES_STREAM (
  "cid" VARCHAR,
  "height" STRING,
  "block_cid" VARCHAR,
  "method" INTEGER,
  "method_name" VARCHAR,
  "from" VARCHAR,
  "from_id" VARCHAR,
  "from_type" VARCHAR,
  "to" VARCHAR,
  "to_id" VARCHAR,
  "to_type" VARCHAR,
  "value" STRING,
  "gas_limit" BIGINT,
  "gas_premium" STRING,
  "gas_fee_cap" STRING,
  "base_fee" STRING,
  "params" STRING,
  "data" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='{APP_PREFIX}_MESSAGES_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM {APP_PREFIX}_MESSAGES_STREAM EMIT CHANGES;