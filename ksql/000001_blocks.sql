CREATE STREAM {APP_PREFIX}_BLOCKS_STREAM (
  "cid" VARCHAR,
  "height" STRING,
  "parents" VARCHAR,
  "win_count" INTEGER,
  "miner" VARCHAR,
  "messages_cid" VARCHAR,
  "validated" BOOLEAN,
  "blocksig" VARCHAR,
  "bls_aggregate" VARCHAR,
  "block" VARCHAR,
  "parent_base_fee" STRING,
  "block_time" BIGINT
) WITH (kafka_topic='{APP_PREFIX}_BLOCKS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_BLOCKS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM {APP_PREFIX}_BLOCKS_STREAM EMIT CHANGES;
