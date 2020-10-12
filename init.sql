CREATE SCHEMA IF NOT EXISTS filecoin;

CREATE TABLE IF NOT EXISTS filecoin.blocks
(
    "cid"           VARCHAR(256) NOT NULL,
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
    "cid"        VARCHAR(256) NOT NULL,
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

-- Temp tbls

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
            to_jsonb(NEW."parents"),
            NEW."win_count",
            NEW."miner",
            NEW."messages_cid",
            NEW."validated",
            to_jsonb(NEW."blocksig"),
            to_jsonb(NEW."bls_aggregate"),
            to_jsonb(NEW."block"),
            to_timestamp(NEW."block_time"));

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
            to_jsonb(NEW."gas"),
            NEW."params",
            to_jsonb(NEW."data"),
            to_timestamp(NEW."block_time"));

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


-- Create indexes

CREATE INDEX filecoin.block_height_idx ON filecoin.blocks ("height");

-- tmp
CREATE INDEX filecoin.block_height_idx ON filecoin.blocks ("cid");
CREATE INDEX filecoin.block_height_idx ON filecoin.messages ("cid");

