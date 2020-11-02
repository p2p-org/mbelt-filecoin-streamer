CREATE TYPE actor_states_return AS
(
    "actor_state_key"  VARCHAR(256),
    "actor_code"       VARCHAR(256),
    "actor_head"       VARCHAR(256),
    "nonce"            BIGINT,
    "balance"          BIGINT,
    "is_account_actor" BOOLEAN,
    "state_root"       VARCHAR(256),
    "height"           BIGINT,
    "ts_key"           VARCHAR(256),
    "parent_ts_key"    VARCHAR(256),
    "addr"             VARCHAR(256),
    "state"            JSONB
);

CREATE OR REPLACE FUNCTION graphql.all_actor_states()
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
ORDER BY a.height DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_account_actor_states()
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
WHERE a.is_account_actor = TRUE
ORDER BY a.height DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_actor_states_by_addr(addr_query VARCHAR(256))
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
WHERE a.addr = addr_query
ORDER BY a.height DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_actor_states_by_height(height_query BIGINT)
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
WHERE a.height = height_query
ORDER BY a.height DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.get_actor_state_by_addr_and_height(addr_query VARCHAR(256), height_query BIGINT)
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
WHERE a.addr = addr_query AND a.height = height_query
ORDER BY a.height DESC
LIMIT 1
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_actor_state_by_ts_key(ts_key_query VARCHAR(256))
    RETURNS SETOF actor_states_return
AS
$$
SELECT a.actor_state_key,
       a.actor_code,
       a.actor_head,
       a.nonce,
       a.balance,
       a.is_account_actor,
       a.state_root,
       a.height,
       a.ts_key,
       a.parent_ts_key,
       a.addr,
       a.state
FROM filecoin.actor_states AS a
WHERE a.ts_key = ts_key_query
ORDER BY a.height DESC
$$
    LANGUAGE sql STABLE;