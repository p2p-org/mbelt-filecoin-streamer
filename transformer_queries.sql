CREATE STREAM BLOCKS_STREAM (
  "cid" VARCHAR,
  "height" BIGINT,
  "parents" VARCHAR,
  "win_count" INTEGER,
  "miner" VARCHAR,
  "messages_cid" VARCHAR,
  "validated" BOOLEAN,
  "blocksig" VARCHAR,
  "bls_aggregate" VARCHAR,
  "block" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='blocks_stream', value_format='JSON');

CREATE STREAM BLOCKS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM BLOCKS_STREAM EMIT CHANGES;

CREATE STREAM BLOCKS_TO_REVERT_STREAM (
  "cid" VARCHAR,
  "height" BIGINT,
  "parents" VARCHAR,
  "win_count" INTEGER,
  "miner" VARCHAR,
  "messages_cid" VARCHAR,
  "validated" BOOLEAN,
  "blocksig" VARCHAR,
  "bls_aggregate" VARCHAR,
  "block" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='blocks_to_revert_stream', value_format='JSON');

CREATE STREAM BLOCKS_TO_REVERT_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM BLOCKS_TO_REVERT_STREAM EMIT CHANGES;

CREATE STREAM MESSAGES_STREAM (
  "cid" VARCHAR,
  "block_cid" VARCHAR,
  "method" INTEGER,
  "from" VARCHAR,
  "to" VARCHAR,
  "value" BIGINT,
  "gas" VARCHAR,
  "params" STRING,
  "data" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='messages_stream', value_format='JSON');

CREATE STREAM MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM MESSAGES_STREAM EMIT CHANGES;
