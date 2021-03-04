CREATE {APP_PREFIX}_STREAM MINER_INFOS_STREAM (
  "miner_info_key" VARCHAR,
  "miner" VARCHAR,
  "owner" VARCHAR,
  "worker" VARCHAR,
  "control_addresses" STRING,
  "new_worker_address" VARCHAR,
  "new_worker_effective_at" STRING,
  "peer_id" VARCHAR,
  "multiaddrs" STRING,
  "seal_proof_type" INT,
  "sector_size" STRING,
  "window_post_partition_sectors" STRING,
  "miner_raw_byte_power" STRING,
  "miner_quality_adj_power" STRING,
  "total_raw_byte_power" STRING,
  "total_quality_adj_power" STRING,
  "height" STRING
) WITH (kafka_topic='{APP_PREFIX}_MINER_INFOS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_MINER_INFOS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM {APP_PREFIX}_MINER_INFOS_STREAM EMIT CHANGES;