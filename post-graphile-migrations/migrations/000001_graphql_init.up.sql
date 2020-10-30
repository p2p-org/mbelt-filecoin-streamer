BEGIN;

CREATE SCHEMA IF NOT EXISTS graphql;

CREATE TYPE block_return AS
(
    "cid"           VARCHAR(256),
    "height"        BIGINT,
    "parents"       JSONB,
    "win_count"     INT,
    "msg_count"     BIGINT,
    "miner"         VARCHAR(128),
    "validated"     BOOLEAN,
    "blocksig"      JSONB,
    "bls_aggregate" JSONB,
    "block"         JSONB,
    "block_time"    TIMESTAMP
);

CREATE TYPE tipset_return AS
(
    "height"        BIGINT,
    "parents"       VARCHAR(256)[],
    "parent_weight" BIGINT,
    "parent_state"  VARCHAR,
    "blocks_count"  BIGINT,
    "min_timestamp" TIMESTAMP
);

CREATE TYPE message_return AS
(
    "cid"        VARCHAR(256),
    "block_cid"  VARCHAR(256),
    "method"     INT,
    "from"       VARCHAR(256),
    "to"         VARCHAR(256),
    "value"      BIGINT,
    "gas"        JSONB,
    "params"     TEXT,
    "data"       JSONB,
    "block_time" TIMESTAMP
);

CREATE OR REPLACE FUNCTION graphql.all_blocks()
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
ORDER BY b.block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.block_by_cid(cid_query VARCHAR(256))
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
WHERE b.cid = cid_query
ORDER BY b.block_time DESC
LIMIT 1
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.blocks_by_height(height_query BIGINT)
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
WHERE b.height = height_query
ORDER BY b.block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.blocks_by_miner(miner_query VARCHAR(128))
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
WHERE b.miner = miner_query
ORDER BY b.block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.validated_blocks()
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
WHERE b.validated = TRUE
ORDER BY b.block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.unvalidated_blocks()
    RETURNS SETOF block_return
AS
$$
SELECT b.cid,
       b.height,
       b.parents,
       b.win_count,
       (SELECT count(*) FROM filecoin.messages WHERE block_cid = b.cid) as msg_count,
       b.miner,
       b.validated,
       b.blocksig,
       b.bls_aggregate,
       b.block,
       b.block_time
FROM filecoin.blocks AS b
WHERE b.validated = FALSE
ORDER BY b.block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_tipsets()
    RETURNS SETOF tipset_return
AS
$$
SELECT t.height,
       t.parents,
       t.parent_weight,
       t.parent_state,
       (SELECT count(*) FROM filecoin.blocks WHERE height = t.height) as blocks_count,
       t.min_timestamp
FROM filecoin.tipsets AS t
ORDER BY t.min_timestamp DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.tipset_by_height(height_query BIGINT)
    RETURNS SETOF tipset_return
AS
$$
SELECT t.height,
       t.parents,
       t.parent_weight,
       t.parent_state,
       (SELECT count(*) FROM filecoin.blocks WHERE height = t.height) as blocks_count,
       t.min_timestamp
FROM filecoin.tipsets AS t
WHERE t.height = height_query
ORDER BY t.min_timestamp DESC
LIMIT 1
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.tipsets_by_block_cid(cid_query VARCHAR(256))
    RETURNS SETOF tipset_return
AS
$$
SELECT t.height,
       t.parents,
       t.parent_weight,
       t.parent_state,
       (SELECT count(*) FROM filecoin.blocks WHERE height = t.height) as blocks_count,
       t.min_timestamp
FROM filecoin.tipsets AS t
WHERE t.height IN (SELECT height FROM filecoin.blocks WHERE cid = cid_query)
ORDER BY t.min_timestamp DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.all_messages()
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.message_by_cid(cid_query VARCHAR(256))
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE cid = cid_query
ORDER BY block_time DESC
LIMIT 1
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_block_cid(cid_query VARCHAR(256))
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE block_cid = cid_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_from(from_query VARCHAR(256))
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "from" = from_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_to(to_query VARCHAR(256))
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "to" = to_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_from_and_to(from_query VARCHAR(256), to_query VARCHAR(256))
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "from" = from_query AND "to" = to_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_to_and_method(to_query VARCHAR(256), method_query INT)
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "to" = to_query AND "method" = method_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_from_and_method(from_query VARCHAR(256), method_query INT)
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "from" = from_query AND "method" = method_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION graphql.messages_by_from_and_to_and_method(from_query VARCHAR(256), to_query VARCHAR(256), method_query INT)
    RETURNS SETOF message_return
AS
$$
SELECT *
FROM filecoin.messages
WHERE "from" = from_query AND "to" = to_query AND "method" = method_query
ORDER BY block_time DESC
$$
    LANGUAGE sql STABLE;

COMMIT;