CREATE STREAM {APP_PREFIX}_ACTOR_STATES_STREAM (
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
) WITH (kafka_topic='{APP_PREFIX}_ACTOR_STATES_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_ACTOR_STATES_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM {APP_PREFIX}_ACTOR_STATES_STREAM EMIT CHANGES;