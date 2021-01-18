#!/bin/bash
docker-compose up -d --force-recreate --renew-anon-volumes zookeeper broker

docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic blocks_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic messages_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic message_receipts_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic tipsets_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic tipsets_to_revert_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic actor_states_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic miner_infos_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic miner_sectors_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic reward_actor_states_stream

docker-compose up -d --force-recreate schema-registry connect control-center ksqldb-server ksqldb-cli ksql-datagen rest-proxy db

echo "Waiting for ksql containers..."
/bin/sleep 210 # we should wait a little bit. Don't know why, but sleep 3m 30 doesn't work on macOS but sleep 210 works just right and does the same thing

curl -X "POST" "http://localhost:8088/ksql" \
     -H "Content-Type: application/vnd.ksql.v1+json; charset=utf-8" \
     --data @transformer_queries.json


curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/blocks_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/messages_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/message_receipts_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/tipsets_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/tipsets_to_revert_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/actor_states_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/miner_infos_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/miner_sectors_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/reward_actor_states_stream_avro_sink.json http://localhost:8083/connectors