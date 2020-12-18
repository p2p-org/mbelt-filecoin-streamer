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

docker-compose up -d --force-recreate schema-registry connect control-center ksqldb-server ksqldb-cli ksql-datagen rest-proxy db redis email redash-server redash-scheduler redash-worker

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

docker-compose run --rm redash-server create_db
curl 'http://localhost:5000/setup' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9' \
  --cookie-jar cookies.txt \
  --data-raw 'name=admin&email=admin%40p2p.org&password=supersecret123&security_notifications=y&org_name=p2p' \
  --compressed

#curl 'http://localhost:5000/api/data_sources' \
#  -H 'Connection: keep-alive' \
#  -H 'Pragma: no-cache' \
#  -H 'Cache-Control: no-cache' \
#  -H 'Accept: application/json, text/plain, */*' \
#  -H 'DNT: 1' \
#  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36' \
#  -H 'Content-Type: application/json;charset=UTF-8' \
#  -H 'Origin: http://localhost:5000' \
#  -H 'Sec-Fetch-Site: same-origin' \
#  -H 'Sec-Fetch-Mode: cors' \
#  -H 'Sec-Fetch-Dest: empty' \
#  -H 'Referer: http://localhost:5000/data_sources' \
#  -H 'Accept-Language: en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7' \
#  -H 'Cookie: JSESSIONID=node0p2s19x7kw2v8if1lnudgom091.node0; OAuth_Token_Request_State=fc1e3179-461d-4723-aeb4-60a97b1fbf45; Goland-d8adb5d2=877b5e79-3973-47e2-87f5-f6988e0b9788; ajs_anonymous_id=%22d2f78e54-c242-4db1-b288-6d14d57edd55%22; ajs_user_id=%22d2f78e54-c242-4db1-b288-6d14d57edd55%22; session=.eJwdjs2KwjAUhV9luGsHmia2WJiFQ7R04N5iqZZkI5bGtrHZ1BE14rtPmNWB8_NxXnA8z-Y6QPY738wCjmMH2Qs-WsigzH-clger6h3Tji7YbK2ue0bNdkS_v6sGn1oOtswPAzma0PYRur1Xvn_oGuPQ5ejXd5LfA_ki0m4nynrNURaMfPACHyVdlNtw5WmkPHDz4kEeY3JKBBW6qSbylUPbBVUxykukLD5JVhOGTcgH3Wy-4L2A29XM__-BfSatOJm4Y8v4nIqVSHmybNkqZYYZnpxSDu8_ehdQUQ.EpYJNw.OZmHXdXOqWxV1BczVHJ92G3ou1I' \
#  --data-binary '{"options":{"host":"db","port":5432,"user":"sink","password":"Ekj31R2_03S2IwLoPsWVa28_sMx_xoS","sslmode":"disable","dbname":"raw"},"type":"pg","name":"mmm"}' \
#  --compressed
curl 'http://localhost:5000/api/data_sources' \
  -b cookies.txt \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Content-Type: application/json;charset=UTF-8' \
  --data-binary '{"options":{"host":"db","port":5432,"user":"sink","password":"Ekj31R2_03S2IwLoPsWVa28_sMx_xoS","sslmode":"disable","dbname":"raw"},"type":"pg","name":"mbelt"}' \
  --compressed