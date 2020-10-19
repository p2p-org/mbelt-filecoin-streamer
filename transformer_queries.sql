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


CREATE STREAM TIPSETS_STREAM (
  "height" BIGINT,
  "parents" STRING,
  "parent_weight" BIGINT,
  "parent_state" VARCHAR,
  "blocks" STRING,
  "min_timestamp" BIGINT
) WITH (kafka_topic='tipsets_stream', value_format='JSON');

CREATE STREAM TIPSETS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM TIPSETS_STREAM EMIT CHANGES;

CREATE STREAM TIPSETS_TO_REVERT_STREAM (
  "height" BIGINT,
  "parents" STRING,
  "parent_weight" BIGINT,
  "parent_state" VARCHAR,
  "blocks" STRING,
  "min_timestamp" BIGINT
) WITH (kafka_topic='tipsets_to_revert_stream', value_format='JSON');

CREATE STREAM TIPSETS_TO_REVERT_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM TIPSETS_TO_REVERT_STREAM EMIT CHANGES;