CREATE STREAM {APP_PREFIX}_REWARD_ACTOR_STATES_STREAM (
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
) WITH (kafka_topic='{APP_PREFIX}_REWARD_ACTOR_STATES_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_REWARD_ACTOR_STATES_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM {APP_PREFIX}_REWARD_ACTOR_STATES_STREAM EMIT CHANGES;