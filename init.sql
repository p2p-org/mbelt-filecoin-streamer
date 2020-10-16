CREATE SCHEMA IF NOT EXISTS filecoin;

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
    "value"      BIGINT,
    "gas"        JSONB,
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
    "min_timestamp" TIMESTAMP
);

-- Internal tables
CREATE TABLE IF NOT EXISTS filecoin.blocks_to_revert
(
    "cid"           VARCHAR(256) NOT NULL PRIMARY KEY,
    "height"        BIGINT,
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

-- Temp tbls


-- Fix for unquoting varchar json
CREATE OR REPLACE FUNCTION varchar_to_jsonb(varchar) RETURNS jsonb AS
$$
SELECT to_jsonb($1)
$$ LANGUAGE SQL;

CREATE CAST (varchar as jsonb) WITH FUNCTION varchar_to_jsonb(varchar) AS IMPLICIT;

CREATE TABLE IF NOT EXISTS filecoin._blocks
(
    "cid"           VARCHAR(256) NOT NULL PRIMARY KEY,
    "height"        BIGINT,
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
    "value"      BIGINT,
    "gas"        TEXT,
    "params"     TEXT,
    "data"       TEXT,
    "block_time" BIGINT
);

CREATE TABLE IF NOT EXISTS filecoin._tipsets
(
    "height"        BIGINT NOT NULL PRIMARY KEY,
    "parents"       TEXT,
    "parent_weight" BIGINT,
    "parent_state"  VARCHAR,
    "blocks"        TEXT,
    "min_timestamp" BIGINT
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
            NEW."height",
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

-- Blocks to revert

CREATE OR REPLACE FUNCTION filecoin.sink_revert_blocks()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin.blocks WHERE "cid" = NEW."cid";
    DELETE FROM filecoin.messages WHERE "block_cid" = NEW."cid";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_sink_revert
    BEFORE INSERT
    ON filecoin.blocks_to_revert
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_revert_blocks();

CREATE OR REPLACE FUNCTION filecoin.sink_trim_blocks_to_revert_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM filecoin.blocks_to_revert WHERE "cid" = NEW."cid";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_to_revert_sink_trim_after_upsert
    AFTER INSERT
    ON filecoin.blocks_to_revert
    FOR EACH ROW
EXECUTE PROCEDURE filecoin.sink_trim_blocks_to_revert_after_insert();

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
            NEW."value",
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
                                 "min_timestamp")
    VALUES (NEW."height",
            NEW."parents"::varchar(256)[],
            NEW."parent_weight",
            NEW."parent_state",
            NEW."blocks"::varchar(256)[],
            to_timestamp(NEW."min_timestamp"))
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

-- Create indexes

CREATE INDEX filecoin_block_height_idx ON filecoin.blocks ("height");

-- tmp
CREATE INDEX filecoin_block_cid_idx ON filecoin.blocks ("cid");
CREATE INDEX filecoin_messages_cid_idx ON filecoin.messages ("cid");

