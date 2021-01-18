CREATE STREAM BLOCKS_STREAM (
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
) WITH (kafka_topic='blocks_stream', value_format='JSON');

CREATE STREAM BLOCKS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM BLOCKS_STREAM EMIT CHANGES;

CREATE STREAM MESSAGES_STREAM (
  "cid" VARCHAR,
  "block_cid" VARCHAR,
  "method" INTEGER,
  "from" VARCHAR,
  "to" VARCHAR,
  "value" STRING,
  "gas_limit" BIGINT,
  "gas_premium" STRING,
  "gas_fee_cap" STRING,
  "params" STRING,
  "data" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='messages_stream', value_format='JSON');

CREATE STREAM MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM MESSAGES_STREAM EMIT CHANGES;

CREATE STREAM MESSAGE_RECEIPTS_STREAM (
  "cid" VARCHAR,
  "gas_used" BIGINT,
  "exit_code" INTEGER,
  "return" STRING
) WITH (kafka_topic='message_receipts_stream', value_format='JSON');

CREATE STREAM MESSAGE_RECEIPTS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM MESSAGE_RECEIPTS_STREAM EMIT CHANGES;

CREATE STREAM TIPSETS_STREAM (
  "height" STRING,
  "parents" STRING,
  "parent_weight" STRING,
  "parent_state" VARCHAR,
  "blocks" STRING,
  "min_timestamp" BIGINT,
  "state" INT
) WITH (kafka_topic='tipsets_stream', value_format='JSON');

CREATE STREAM TIPSETS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM TIPSETS_STREAM EMIT CHANGES;

CREATE STREAM TIPSETS_TO_REVERT_STREAM (
  "height" STRING,
  "parents" STRING,
  "parent_weight" STRING,
  "parent_state" VARCHAR,
  "blocks" STRING,
  "min_timestamp" BIGINT,
  "state" INT
) WITH (kafka_topic='tipsets_to_revert_stream', value_format='JSON');

CREATE STREAM TIPSETS_TO_REVERT_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM TIPSETS_TO_REVERT_STREAM EMIT CHANGES;

CREATE STREAM ACTOR_STATES_STREAM (
  "actor_state_key" VARCHAR,
  "actor_code" VARCHAR,
  "actor_head" VARCHAR,
  "nonce" STRING,
  "balance" STRING,
  "state_root" VARCHAR,
  "height" STRING,
  "ts_key" VARCHAR,
  "parent_ts_key" VARCHAR,
  "addr" VARCHAR,
  "state" VARCHAR,
  "deleted" BOOLEAN
) WITH (kafka_topic='actor_states_stream', value_format='JSON');

CREATE STREAM ACTOR_STATES_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM ACTOR_STATES_STREAM EMIT CHANGES;

CREATE STREAM MINER_INFOS_STREAM (
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
) WITH (kafka_topic='miner_infos_stream', value_format='JSON');

CREATE STREAM MINER_INFOS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM MINER_INFOS_STREAM EMIT CHANGES;

CREATE STREAM MINER_SECTORS_STREAM (
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
) WITH (kafka_topic='miner_sectors_stream', value_format='JSON');

CREATE STREAM MINER_SECTORS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM MINER_SECTORS_STREAM EMIT CHANGES;

CREATE STREAM REWARD_ACTOR_STATES_STREAM (
  "epoch" STRING,
  "actor_code" VARCHAR,
  "actor_head" VARCHAR,
  "nonce" STRING,
  "balance" STRING,
  "state_root" VARCHAR,
  "ts_key" VARCHAR,
  "parent_ts_key" VARCHAR,
  "addr" VARCHAR,
  "cumsum_baseline" STRING,
  "cumsum_realized" STRING,
  "effective_baseline_power" STRING,
  "effective_network_time" INT,
  "this_epoch_baseline_power" STRING,
  "this_epoch_reward" STRING,
  "total_mined" STRING,
  "simple_total" STRING,
  "baseline_total" STRING,
  "total_storage_power_reward" STRING,
  "this_epoch_reward_smoothed_position_estimate" STRING,
  "this_epoch_reward_smoothed_velocity_estimate" STRING
) WITH (kafka_topic='reward_actor_states_stream', value_format='JSON');

CREATE STREAM REWARD_ACTOR_STATES_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM REWARD_ACTOR_STATES_STREAM EMIT CHANGES;