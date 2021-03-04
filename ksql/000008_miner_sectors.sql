CREATE STREAM {APP_PREFIX}_MINER_SECTORS_STREAM (
  "miner_sector_key" VARCHAR,
  "sector_number" STRING,
  "seal_proof" INT,
  "sealed_cid" VARCHAR,
  "deal_ids" STRING,
  "activation" STRING,
  "expiration" STRING,
  "deal_weight" STRING,
  "verified_deal_weight" STRING,
  "initial_pledge" STRING,
  "expected_day_reward" STRING,
  "expected_storage_pledge" STRING,
  "miner" VARCHAR,
  "height" STRING
) WITH (kafka_topic='{APP_PREFIX}_MINER_SECTORS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_MINER_SECTORS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM {APP_PREFIX}_MINER_SECTORS_STREAM EMIT CHANGES;