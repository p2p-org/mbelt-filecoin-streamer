#!/bin/bash
docker-compose up -d zookeeper broker

docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic blocks_stream
docker-compose exec broker kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic messages_stream

docker-compose up -d --build schema-registry connect control-center ksqldb-server ksqldb-cli ksql-datagen rest-proxy db

echo "Starting ksql containers..."
sleep 3m # we should wait a little bit

# create streams
curl -X "POST" "http://localhost:8088/ksql" \
     -H "Content-Type: application/vnd.ksql.v1+json; charset=utf-8" \
     -d $'{
  "ksql": "CREATE STREAM BLOCKS_STREAM (Miner VARCHAR, Ticket MAP<VARCHAR, VARCHAR>, ElectionProof MAP<VARCHAR, VARCHAR>, BeaconEntries VARCHAR, WinPoStProof VARCHAR, Parents VARCHAR, ParentWeight VARCHAR, Height BIGINT, ParentStateRoot MAP<VARCHAR, VARCHAR>, ParentMessageReceipts MAP<VARCHAR, VARCHAR>, Messages MAP<VARCHAR, VARCHAR>, BLSAggregate MAP<VARCHAR, VARCHAR>, Timestamp BIGINT, BlockSig MAP<VARCHAR, VARCHAR>, ForkSignaling BIGINT) WITH (kafka_topic=\'blocks_stream\', value_format=\'JSON\', TIMESTAMP=\'Timestamp\'); CREATE STREAM BLOCKS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT=\'AVRO\', REPLICAS=1) AS SELECT MINER, TICKET[\'VRFProof\'] as ticket_vrfproof, ELECTIONPROOF[\'WinCount\'] as electionproof_wincount, ELECTIONPROOF[\'VRFProof\'] as electionproof_vrfproof, BEACONENTRIES, WINPOSTPROOF, PARENTS, PARENTWEIGHT, HEIGHT, PARENTSTATEROOT[\'/\'] as parentstateroot, PARENTMESSAGERECEIPTS[\'/\'] as parentmessagereceipts, MESSAGES[\'/\'] as messages, BLSAGGREGATE[\'Type\'] as blsaggreagate_type, BLSAGGREGATE[\'Data\'] as blsaggregate_data, TIMESTAMP, BLOCKSIG[\'Type\'] as blocksig_type, BLOCKSIG[\'Data\'] as blocksig_data, FORKSIGNALING from BLOCKS_STREAM EMIT CHANGES; CREATE STREAM MESSAGES_STREAM (Version BIGINT, \\"TO\\" VARCHAR, \\"FROM\\" VARCHAR, Nonce BIGINT, Value VARCHAR, GasPrice VARCHAR, GasLimit BIGINT, Method BIGINT, Params VARCHAR) WITH (kafka_topic=\'messages_stream\', value_format=\'JSON\'); CREATE STREAM MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT=\'AVRO\') AS select * from MESSAGES_STREAM EMIT CHANGES;",
  "streamsProperties": {}
}'

curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/blocks_stream_avro_sink.json http://localhost:8083/connectors
curl -X "POST" -H "Accept:application/json" -H "Content-Type: application/json" --data @connectors/messages_stream_avro_sink.json http://localhost:8083/connectors