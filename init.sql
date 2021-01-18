CREATE SCHEMA IF NOT EXISTS filecoin;


CREATE TABLE  IF NOT EXISTS filecoin._config (
    "key" VARCHAR (100) PRIMARY KEY,
    "value" TEXT
);


CREATE TABLE IF NOT EXISTS filecoin.blocks
(
    "cid"           VARCHAR(256) NOT NULL PRIMARY KEY,
    "height"        BIGINT,
    "parents"       JSONB,
    "win_count"     INT,
    "miner"         VARCHAR(128),
    "messages_cid"  VARCHAR(256),
    "validated"     BOOLEAN,
    "blocksig"      JSONB,
    "bls_aggregate" JSONB,
    "block"         JSONB,
    "block_time"    TIMESTAMP
);


CREATE TABLE IF NOT EXISTS filecoin.messages
(
    "cid"        VARCHAR(256) NOT NULL PRIMARY KEY,
    "block_cid"  VARCHAR(256),
    "method"     INT,
    "from"       VARCHAR(256),
    "to"         VARCHAR(256),
    "value"      DECIMAL(100, 0),
    "gas"        JSONB,
    "gas_used"   BIGINT,
    "exit_code"  INT,
    "return"     TEXT,
    "params"     TEXT,
    "data"       JSONB,
    "block_time" TIMESTAMP
);

CREATE TABLE IF NOT EXISTS filecoin.tipsets
(
    "height"        BIGINT NOT NULL PRIMARY KEY,
    "parents"       VARCHAR(256)[],
    "parent_weight" BIGINT,
    "parent_state"  VARCHAR,
    "blocks"        VARCHAR(256)[],
    "min_timestamp" TIMESTAMP,
    "state"         SMALLINT
);

CREATE TABLE IF NOT EXISTS filecoin.actor_states
(
    "actor_state_key"  VARCHAR(512) NOT NULL PRIMARY KEY,
    "actor_code"       VARCHAR(512),
    "actor_head"       VARCHAR(256),
    "nonce"            DECIMAL(100, 0),
    "balance"          DECIMAL(100, 0),
    "state_root"       VARCHAR(256),
    "height"           BIGINT,
    "ts_key"           TEXT,
    "parent_ts_key"    TEXT,
    "addr"             VARCHAR(128),
    "state"            JSONB,
    "deleted"          BOOLEAN
);

CREATE TABLE IF NOT EXISTS filecoin.miner_infos (
  "miner_info_key"                VARCHAR(512) NOT NULL PRIMARY KEY,
  "miner"                         VARCHAR(128),
  "owner"                         VARCHAR(128),
  "worker"                        VARCHAR(128),
  "control_addresses"             VARCHAR(128)[],
  "new_worker_address"            VARCHAR(128),
  "new_worker_effective_at"       BIGINT,
  "peer_id"                       VARCHAR(256),
  "multiaddrs"                    VARCHAR(256)[],
  "seal_proof_type"               INT,
  "sector_size"                   BIGINT,
  "window_post_partition_sectors" BIGINT,
  "miner_raw_byte_power"          BIGINT,
  "miner_quality_adj_power"       BIGINT,
  "total_raw_byte_power"          BIGINT,
  "total_quality_adj_power"       BIGINT,
  "height"                        BIGINT
);

CREATE TABLE IF NOT EXISTS filecoin.miner_sectors (
  "miner_sector_key"        VARCHAR(512) NOT NULL PRIMARY KEY,
  "sector_number"           BIGINT,
  "seal_proof"              INT,
  "sealed_cid"              VARCHAR(256),
  "deal_ids"                INT[],
  "activation"              DECIMAL(100, 0),
  "expiration"              DECIMAL(100, 0),
  "deal_weight"             DECIMAL(100, 0),
  "verified_deal_weight"    DECIMAL(100, 0),
  "initial_pledge"          DECIMAL(100, 0),
  "expected_day_reward"     DECIMAL(100, 0),
  "expected_storage_pledge" DECIMAL(100, 0),
  "miner"                   VARCHAR(128),
  "height"                  BIGINT
);

CREATE TABLE IF NOT EXISTS filecoin.reward_actor_states (
  "epoch"                                        BIGINT NOT NULL PRIMARY KEY,
  "actor_code"                                   VARCHAR(512),
  "actor_head"                                   VARCHAR(256),
  "nonce"                                        DECIMAL(100, 0),
  "balance"                                      DECIMAL(100, 0),
  "state_root"                                   VARCHAR(256),
  "ts_key"                                       TEXT,
  "parent_ts_key"                                TEXT,
  "addr"                                         VARCHAR(128),
  "cumsum_baseline"                              DECIMAL(100, 0),
  "cumsum_realized"                              DECIMAL(100, 0),
  "effective_baseline_power"                     DECIMAL(100, 0),
  "effective_network_time"                       INT,
  "this_epoch_baseline_power"                    DECIMAL(100, 0),
  "this_epoch_reward"                            DECIMAL(100, 0),
  "total_mined"                                  DECIMAL(100, 0),
  "simple_total"                                 DECIMAL(100, 0),
  "baseline_total"                               DECIMAL(100, 0),
  "total_storage_power_reward"                   DECIMAL(100, 0),
  "this_epoch_reward_smoothed_position_estimate" DECIMAL(100, 0),
  "this_epoch_reward_smoothed_velocity_estimate" DECIMAL(100, 0)
);

-- Fix for unquoting varchar json
CREATE OR REPLACE FUNCTION varchar_to_jsonb(varchar) RETURNS jsonb AS
$$
SELECT to_jsonb($1)
$$ LANGUAGE SQL;

CREATE CAST (varchar as jsonb) WITH FUNCTION varchar_to_jsonb(varchar) AS IMPLICIT;

-- Internal tables

CREATE TABLE IF NOT EXISTS filecoin._blocks
(
    "cid"           VARCHAR(256) NOT NULL PRIMARY KEY,
    "height"        TEXT,
    "parents"       TEXT,
    "win_count"     INT,
    "miner"         VARCHAR(128),
    "messages_cid"  VARCHAR(256),
    "validated"     BOOLEAN,
    "blocksig"      TEXT,
    "bls_aggregate" TEXT,
    "block"         TEXT,
    "block_time"    BIGINT
);


CREATE TABLE IF NOT EXISTS filecoin._messages
(
    "cid"        VARCHAR(256) NOT NULL PRIMARY KEY,
    "block_cid"  VARCHAR(256),
    "method"     INT,
    "from"       VARCHAR(256),
    "to"         VARCHAR(256),
    "value"      TEXT,
    "gas"        TEXT,
    "params"     TEXT,
    "data"       TEXT,
    "block_time" BIGINT
);

CREATE TABLE IF NOT EXISTS filecoin._message_receipts
(
    "cid"        VARCHAR(256) NOT NULL PRIMARY KEY,
    "gas_used"   BIGINT,
    "exit_code"  INT,
    "return"     TEXT
);

CREATE TABLE IF NOT EXISTS filecoin._tipsets
(
    "height"        TEXT NOT NULL PRIMARY KEY,
    "parents"       TEXT,
    "parent_weight" TEXT,
    "parent_state"  VARCHAR,
    "blocks"        TEXT,
    "min_timestamp" BIGINT,
    "state"         SMALLINT
);

CREATE TABLE IF NOT EXISTS filecoin._tipsets_to_revert
(
    "height"        TEXT NOT NULL PRIMARY KEY,
    "parents"       VARCHAR(256)[],
    "parent_weight" TEXT,
    "parent_state"  VARCHAR,
    "blocks"        VARCHAR(256)[],
    "min_timestamp" TIMESTAMP,
    "state"         SMALLINT
);

CREATE TABLE IF NOT EXISTS filecoin._actor_states
(
    "actor_state_key"  VARCHAR(512) NOT NULL PRIMARY KEY,
    "actor_code"       VARCHAR(256),
    "actor_head"       VARCHAR(256),
    "nonce"            TEXT,
    "balance"          TEXT,
    "state_root"       VARCHAR(256),
    "height"           TEXT,
    "ts_key"           TEXT,
    "parent_ts_key"    TEXT,
    "addr"             VARCHAR(128),
    "state"            TEXT,
    "deleted"          BOOLEAN
);

CREATE TABLE IF NOT EXISTS filecoin._miner_infos (
    "miner_info_key"                VARCHAR(512) NOT NULL PRIMARY KEY,
    "miner"                         VARCHAR(128),
    "owner"                         VARCHAR(128),
    "worker"                        VARCHAR(128),
    "control_addresses"             TEXT,
    "new_worker_address"            VARCHAR(128),
    "new_worker_effective_at"       TEXT,
    "peer_id"                       VARCHAR(256),
    "multiaddrs"                    TEXT,
    "seal_proof_type"               INT,
    "sector_size"                   TEXT,
    "window_post_partition_sectors" TEXT,
    "miner_raw_byte_power"          TEXT,
    "miner_quality_adj_power"       TEXT,
    "total_raw_byte_power"          TEXT,
    "total_quality_adj_power"       TEXT,
    "height"                        TEXT
);

CREATE TABLE IF NOT EXISTS filecoin._miner_sectors (
    "miner_sector_key"        VARCHAR(512) NOT NULL PRIMARY KEY,
    "sector_number"           TEXT,
    "seal_proof"              INT,
    "sealed_cid"              VARCHAR(256),
    "deal_ids"                TEXT,
    "activation"              TEXT,
    "expiration"              TEXT,
    "deal_weight"             TEXT,
    "verified_deal_weight"    TEXT,
    "initial_pledge"          TEXT,
    "expected_day_reward"     TEXT,
    "expected_storage_pledge" TEXT,
    "miner"                   VARCHAR(128),
    "height"                  TEXT
);

CREATE TABLE IF NOT EXISTS filecoin._reward_actor_states (
    "epoch"                                        TEXT NOT NULL PRIMARY KEY,
    "actor_code"                                   VARCHAR(512),
    "actor_head"                                   VARCHAR(256),
    "nonce"                                        TEXT,
    "balance"                                      TEXT,
    "state_root"                                   VARCHAR(256),
    "ts_key"                                       TEXT,
    "parent_ts_key"                                TEXT,
    "addr"                                         VARCHAR(128),
    "cumsum_baseline"                              TEXT,
    "cumsum_realized"                              TEXT,
    "effective_baseline_power"                     TEXT,
    "effective_network_time"                       INT,
    "this_epoch_baseline_power"                    TEXT,
    "this_epoch_reward"                            TEXT,
    "total_mined"                                  TEXT,
    "simple_total"                                 TEXT,
    "baseline_total"                               TEXT,
    "total_storage_power_reward"                   TEXT,
    "this_epoch_reward_smoothed_position_estimate" TEXT,
    "this_epoch_reward_smoothed_velocity_estimate" TEXT
);

-- Blocks

CREATE OR REPLACE FUNCTION filecoin.sink_blocks_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.blocks("cid",
                                "height",
                                "parents",
                                "win_count",
                                "miner",
                                "messages_cid",
                                "validated",
                                "blocksig",
                                "bls_aggregate",
                                "block",
                                "block_time")
    VALUES (NEW."cid",
            NEW."height"::BIGINT,
            NEW."parents"::jsonb,
            NEW."win_count",
            NEW."miner",
            NEW."messages_cid",
            NEW."validated",
            NEW."blocksig"::jsonb,
            NEW."bls_aggregate"::jsonb,
            NEW."block"::jsonb,
            to_timestamp(NEW."block_time"))
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_sink_upsert
    BEFORE INSERT
    ON filecoin._blocks
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_blocks_insert();


-- Blocks

CREATE OR REPLACE FUNCTION filecoin.sink_trim_blocks_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._blocks WHERE "cid" = NEW."cid";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._blocks
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_blocks_after_insert();

-- Tipsets to revert

CREATE OR REPLACE FUNCTION filecoin.sink_revert_tipsets()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin.tipsets WHERE tipsets."height" = NEW."height";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_tipsets_sink_revert
    BEFORE INSERT
    ON filecoin._tipsets_to_revert
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_revert_tipsets();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_tipsets_to_revert_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._tipsets_to_revert WHERE "height" = NEW."height";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_tipsets_to_revert_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._tipsets_to_revert
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_tipsets_to_revert_after_insert();

-- Messages

CREATE OR REPLACE FUNCTION filecoin.sink_messages_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.messages("cid",
                                  "block_cid",
                                  "method",
                                  "from",
                                  "to",
                                  "value",
                                  "gas",
                                  "params",
                                  "data",
                                  "block_time")
    VALUES (NEW."cid",
            NEW."block_cid",
            NEW."method",
            NEW."from",
            NEW."to",
            NEW."value"::DECIMAL(100, 0),
            NEW."gas"::jsonb,
            NEW."params",
            NEW."data"::jsonb,
            to_timestamp(NEW."block_time"))
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_messages_sink_upsert
    BEFORE INSERT
    ON filecoin._messages
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_messages_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_messages_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._messages WHERE "cid" = NEW."cid";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_messages_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._messages
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_messages_after_insert();

-- Message receipts

CREATE OR REPLACE FUNCTION filecoin.sink_message_receipts_insert()
    RETURNS trigger AS
$$
BEGIN
    UPDATE
        filecoin.messages
    SET "gas_used" = NEW."gas_used",
        "exit_code" = NEW."exit_code",
        "return" = NEW."return"
    WHERE "cid" = NEW."cid";

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_message_receipts_sink_upsert
    BEFORE INSERT
    ON filecoin._message_receipts
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_message_receipts_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_message_receipts_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._message_receipts WHERE "cid" = NEW."cid";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_message_receipts_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._message_receipts
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_message_receipts_after_insert();

-- TipSets

CREATE OR REPLACE FUNCTION filecoin.sink_tipsets_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.tipsets("height",
                                 "parents",
                                 "parent_weight",
                                 "parent_state",
                                 "blocks",
                                 "min_timestamp",
                                 "state")
    VALUES (NEW."height"::BIGINT,
            NEW."parents"::VARCHAR(256)[],
            NEW."parent_weight"::BIGINT,
            NEW."parent_state",
            NEW."blocks"::VARCHAR(256)[],
            to_timestamp(NEW."min_timestamp"),
            NEW."state")
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_tipsets_sink_upsert
    BEFORE INSERT
    ON filecoin._tipsets
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_tipsets_insert();



CREATE OR REPLACE FUNCTION filecoin.sink_trim_tipsets_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._tipsets WHERE "height" = NEW."height";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_tipsets_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._tipsets
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_tipsets_after_insert();

-- Actors

CREATE OR REPLACE FUNCTION filecoin.sink_actor_states_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.actor_states("actor_state_key",
                                "actor_code",
                                "actor_head",
                                "nonce",
                                "balance",
                                "state_root",
                                "height",
                                "ts_key",
                                "parent_ts_key",
                                "addr",
                                "state",
                                "deleted")
    VALUES (NEW."actor_state_key",
            NEW."actor_code",
            NEW."actor_head",
            NEW."nonce"::DECIMAL(100, 0),
            NEW."balance"::DECIMAL(100, 0),
            NEW."state_root",
            NEW."height"::BIGINT,
            NEW."ts_key",
            NEW."parent_ts_key",
            NEW."addr",
            NEW."state"::jsonb,
            NEW."deleted")
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_actor_states_sink_upsert
    BEFORE INSERT
    ON filecoin._actor_states
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_actor_states_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_actor_states_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._actor_states WHERE "actor_state_key" = NEW."actor_state_key";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_actor_states_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._actor_states
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_actor_states_after_insert();

-- Miners info

CREATE OR REPLACE FUNCTION filecoin.sink_miner_infos_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.miner_infos(miner_info_key,
                                     miner,
                                     owner,
                                     worker,
                                     control_addresses,
                                     new_worker_address,
                                     new_worker_effective_at,
                                     peer_id,
                                     multiaddrs,
                                     seal_proof_type,
                                     sector_size,
                                     window_post_partition_sectors,
                                     miner_raw_byte_power,
                                     miner_quality_adj_power,
                                     total_raw_byte_power,
                                     total_quality_adj_power,
                                     height)
    VALUES (NEW."miner_info_key",
            NEW."miner",
            NEW."owner",
            NEW."worker",
            NEW."control_addresses"::VARCHAR(128)[],
            NEW."new_worker_address",
            NEW."new_worker_effective_at"::BIGINT,
            NEW."peer_id",
            NEW."multiaddrs"::VARCHAR(256)[],
            NEW."seal_proof_type",
            NEW."sector_size"::DECIMAL(100, 0),
            NEW."window_post_partition_sectors"::DECIMAL(100, 0),
            NEW."miner_raw_byte_power"::DECIMAL(100, 0),
            NEW."miner_quality_adj_power"::DECIMAL(100, 0),
            NEW."total_raw_byte_power"::DECIMAL(100, 0),
            NEW."total_quality_adj_power"::DECIMAL(100, 0),
            NEW."height"::BIGINT)
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_miner_infos_sink_upsert
    BEFORE INSERT
    ON filecoin._miner_infos
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_miner_infos_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_miner_infos_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._miner_infos WHERE "miner_info_key" = NEW."miner_info_key";
    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_miner_infos_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._miner_infos
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_miner_infos_after_insert();

-- Miners sectors

CREATE OR REPLACE FUNCTION filecoin.sink_miner_sectors_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.miner_sectors(miner_sector_key,
                                       sector_number,
                                       seal_proof,
                                       sealed_cid,
                                       deal_ids,
                                       activation,
                                       expiration,
                                       deal_weight,
                                       verified_deal_weight,
                                       initial_pledge,
                                       expected_day_reward,
                                       expected_storage_pledge,
                                       miner,
                                       height)
    VALUES (NEW."miner_sector_key",
            NEW."sector_number"::BIGINT,
            NEW."seal_proof",
            NEW."sealed_cid",
            NEW."deal_ids"::INT[],
            NEW."activation"::DECIMAL(100, 0),
            NEW."expiration"::DECIMAL(100, 0),
            NEW."deal_weight"::DECIMAL(100, 0),
            NEW."verified_deal_weight"::DECIMAL(100, 0),
            NEW."initial_pledge"::DECIMAL(100, 0),
            NEW."expected_day_reward"::DECIMAL(100, 0),
            NEW."expected_storage_pledge"::DECIMAL(100, 0),
            NEW."miner",
            NEW."height"::BIGINT)
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_miner_sectors_sink_upsert
    BEFORE INSERT
    ON filecoin._miner_sectors
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_miner_sectors_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_miner_sectors_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._miner_sectors WHERE "miner_sector_key" = NEW."miner_sector_key";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_miner_sectors_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._miner_sectors
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_miner_sectors_after_insert();

-- Reward actor

CREATE OR REPLACE FUNCTION filecoin.sink_reward_actor_states_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO filecoin.reward_actor_states(epoch,
                                             actor_code,
                                             actor_head,
                                             nonce,
                                             balance,
                                             state_root,
                                             ts_key,
                                             parent_ts_key,
                                             addr,
                                             cumsum_baseline,
                                             cumsum_realized,
                                             effective_baseline_power,
                                             effective_network_time,
                                             this_epoch_baseline_power,
                                             this_epoch_reward,
                                             total_mined,
                                             simple_total,
                                             baseline_total,
                                             total_storage_power_reward,
                                             this_epoch_reward_smoothed_position_estimate,
                                             this_epoch_reward_smoothed_velocity_estimate)
    VALUES (NEW."epoch"::BIGINT,
            NEW."actor_code",
            NEW."actor_head",
            NEW."nonce"::DECIMAL(100, 0),
            NEW."balance"::DECIMAL(100, 0),
            NEW."state_root",
            NEW."ts_key",
            NEW."parent_ts_key",
            NEW."addr",
            NEW."cumsum_baseline"::DECIMAL(100, 0),
            NEW."cumsum_realized"::DECIMAL(100, 0),
            NEW."effective_baseline_power"::DECIMAL(100, 0),
            NEW."effective_network_time",
            NEW."this_epoch_baseline_power"::DECIMAL(100, 0),
            NEW."this_epoch_reward"::DECIMAL(100, 0),
            NEW."total_mined"::DECIMAL(100, 0),
            NEW."simple_total"::DECIMAL(100, 0),
            NEW."baseline_total"::DECIMAL(100, 0),
            NEW."total_storage_power_reward"::DECIMAL(100, 0),
            NEW."this_epoch_reward_smoothed_position_estimate"::DECIMAL(100, 0),
            NEW."this_epoch_reward_smoothed_velocity_estimate"::DECIMAL(100, 0))
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_reward_actor_states_sink_upsert
    BEFORE INSERT
    ON filecoin._reward_actor_states
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_reward_actor_states_insert();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_reward_actor_states_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin._reward_actor_states WHERE "epoch" = NEW."epoch";
    RETURN NEW;
END;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_reward_actor_states_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin._reward_actor_states
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_reward_actor_states_after_insert();

-- Create indexes

CREATE INDEX filecoin_block_height_idx ON filecoin.blocks ("height");
CREATE INDEX filecoin_actor_states_height_idx ON filecoin.actor_states ("height");
CREATE INDEX filecoin_actor_states_addr_idx ON filecoin.actor_states ("addr");
CREATE INDEX filecoin_miner_infos_height_idx ON filecoin.miner_infos ("height");
CREATE INDEX filecoin_miner_infos_miner_idx ON filecoin.miner_infos ("miner");
CREATE INDEX filecoin_miner_sectors_height_idx ON filecoin.miner_sectors ("height");
CREATE INDEX filecoin_miner_sectors_miner_idx ON filecoin.miner_sectors ("miner");

-- tmp
CREATE INDEX filecoin_block_cid_idx ON filecoin.blocks ("cid");
CREATE INDEX filecoin_messages_cid_idx ON filecoin.messages ("cid");

