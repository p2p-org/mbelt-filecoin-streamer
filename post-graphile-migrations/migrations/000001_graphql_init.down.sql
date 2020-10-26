BEGIN;

DROP SCHEMA graphql;

DROP TYPE block_return;

DROP TYPE tipset_return;

DROP TYPE messages_return;

DROP FUNCTION graphql.all_blocks;

DROP FUNCTION graphql.block_by_cid;

DROP FUNCTION graphql.blocks_by_height;

DROP FUNCTION graphql.blocks_by_miner;

DROP FUNCTION graphql.validated_blocks;

DROP FUNCTION graphql.unvalidated_blocks;

DROP FUNCTION graphql.all_tipsets;

DROP FUNCTION graphql.tipset_by_height;

DROP FUNCTION graphql.tipsets_by_block_cid;

DROP FUNCTION graphql.all_messages;

DROP FUNCTION graphql.message_by_cid;

DROP FUNCTION graphql.messages_by_block_cid;

DROP FUNCTION graphql.message_by_from;

DROP FUNCTION graphql.message_by_to;

DROP FUNCTION graphql.message_by_from_and_to;

DROP FUNCTION graphql.message_by_to_and_method;

DROP FUNCTION graphql.message_by_from_and_method;

DROP FUNCTION graphql.message_by_from_and_to_and_method;

COMMIT;