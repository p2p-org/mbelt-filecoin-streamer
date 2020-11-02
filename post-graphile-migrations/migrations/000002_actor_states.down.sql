BEGIN;

DROP TYPE actor_states_return;

DROP FUNCTION graphql.all_actor_states;

DROP FUNCTION graphql.all_account_actor_states;

DROP FUNCTION graphql.all_actor_states_by_addr;

DROP FUNCTION graphql.all_actor_states_by_height;

DROP FUNCTION graphql.get_actor_state_by_addr_and_height;

DROP FUNCTION graphql.all_actor_state_by_ts_key;

COMMIT;