## Requirements
* Docker v1.11+
* Docker Compose
* Docker memory is allocated minimally at 8 GB

## How to start
```shell script
./start-docker.sh
```

## What does the script do?
The script will start all necessary KSQLDB, Kafka, ZooKeeper, PostgreSQL containers, create all demo streams and connectors for DB (It'll take about 5 minutes).

Also, it'll start:
* **FilecoinBlockStreamer** - golang service that streams Filecoin blocks the node and sends it to a Kafka topic;
* **FilecoinBlocksEnrichment** - golang service that listens for fresh blocks in Kafka topic and for each one queries and puts corresponding messages and deals in corresponding Kafka topics

## What is happening inside ksqlDB?
Data flow:
![flow](images/flow.png "Data flow")

1. Data produced by FilecoinBlockStreamer gets into Kafka Topic **blocks_stream** and further into ksqlDB stream **BLOCKS_STREAM**;
    1. A stream called **BLOCKS_STREAM_AVRO** collects raw data from **BLOCKS_STREAM** with AVRO format and sinks it to a Postgres table **blocks**;
2. Data produced by FIlecoinBlocksEnrichment gets into Kafka topic **messages_stream** and further into KSQLDB stream **MESSAGES_STREAM**:
    1. **MESSAGES_STREAM** formats the data to AVRO and sends it to a topic **MESSAGES_STREAM_AVRO** to sink it to a Postgres table **messages**;

## How does data sink to a DB?

KSQLDB has a special JDBCSinkConnector.

It collects data from some Kafka topic and inserts the data to a DB table.

## Message format examples

* **BLOCKS_STREAM** and **BLOCKS_STREAM_AVRO** topics:
 ```json
 [
    {
        "ROWTIME": 1596341400,
        "ROWKEY": "bafy2bzaceasu3xmu4mjhfan55f7cwki2yb5bd5e3j5rdqrkjvqv6v55cqxt5i",
        "MINER": "t01000",
        "TICKET": {
            "VRFProof": "iImSFyDqKRYo4JvE8/1y0oJypeuMxjBDXGnYBHz7bCzpf0tMzC4Nc3LhLumB2hjOCYyJNl9cLbWr+svhf7wpaZt3cW6twFSXxtKJtBUQ5icMfQYik+4s+mxrEsO1GdoJ"
        },
        "ELECTIONPROOF": {
            "WinCount": "1",
            "VRFProof": "t9pVFrq3zhA/+imtiMVw3hAd5u3U7gVxOOeT1iuSfgPHtwIRZFhCRmzWMdNxvHs8BayByHiPYgvH0i+V0KBtWbYb9NwbAMHQtfr/pS6/PpCArrwG36mccdxrQJfE5RNC"
        },
        "BEACONENTRIES": "[{\"Round\":30344,\"Data\":\"jKQkPpIsnrM+vZrOHtr5dOrA/1l2L3tmbmiabJNV5BB1bwRE9kp9Kga2aSCxRYmLAlvfl8krry1esr4S0GGjRSCOEO/fgqmB87BZZEz+44KGj2WSKM3yDmYkHcRKCBeh\"}]",
        "WINPOSTPROOF": "[{\"PoStProof\":3,\"ProofBytes\":\"kNMMcn7zcUXJTw1cIzOFZAAo1lwKU58+f8pYlPYOb5P5cO1k/GIQzHW7RJzBuyAxiBkmQmpC9E3Tj6LNrs8jYHPjiBO7VLeJ8/rL6+5T14bY8/saFcvZQ2ZKdEtAj5yIBlWpKo4zt8VLCAlscYmE/VZvQmHqDVJhMOG/y/ZH0hDuEabFBraXrtd5jotjKdqzhVTa8gZUKcMW0Q6RKnmuGN7ETp+1m/PT1ow7pM+YvwyqADRsJpWVeUd5JwWs4wBz\"}]",
        "PARENTS": "[{\"/\":\"bafy2bzacea27ccbenqe33rvwyhquzjfecvsay4zr3totzryohi6lzwkhnm5vm\"}, {\"/\":\"bafy2bzacebltxaropeu5naxq4vs2hndku4ywjnepgoznod4ieij6szjymkk6u\"}]",
        "PARENTWEIGHT": "48995659",
        "HEIGHT": 3036,
        "PARENTSTATEROOT": {
            "/": "bafy2bzaceb4kpu7sz24nvtgpdm2rzf2jenlztw6x2yfn5xfzlsdydkais6pde"
        },
        "PARENTMESSAGERECEIPTS": {
            "/": "bafy2bzacea7ucaqw6kpwao2r2xpgc5uqtjlpcmjyrcort3nqerxkkofa4spp6"
        },
        "MESSAGES": {
            "/": "bafy2bzacebq27ky2shnnjjyg2bifkfsyfugossb2s2whywrr26q7iwoyybt66"
        },
        "BLSAGGREGATE": {
            "Type": "2",
            "Data": "jVAtOf+qrdE3OiRuKT7uS6xjsIFnchKulXKUFmbaU0C6HL96hBVk5qycqP7IlhejETi89eKTos780afZmxNSKkc8n8qM1sVzE+FE/OtOLINg/Bv/pJ806i5VjYmuEkTa"
        },
        "TIMESTAMP": 1596341400,
        "BLOCKSIG": {
            "Type": "2",
            "Data": "uacLD4ItD8/s/WqlNKbVynZAIlOWUXJs2nm7jkBaBJ5Ek79S9sQ0wx0DHBKpWwQEFC/G1praIYH1HoiMmjzJfP73kaLlHnNgXBtl23fwQPRBq4mO6SYlrcv+dDDy4RVH"
        },
        "FORKSIGNALING": 0
    }
]
```


* **MESSAGES_STREAM** AND **MESSAGES_STREAM_AVRO** topics:

```json
[
    {
        "ROWTIME": 1596473104242,
        "ROWKEY": null,
        "VERSION": 1413335359942951700,
        "TO": "t2jlfk7cgbrbm2mwowtqjl4yq2ujvltguuhe53cfy",
        "FROM": "t22io6tplblvuharkwr6ypse2mdm4iqyingstahvq",
        "NONCE": -1581211366622027800,
        "VALUE": "4978641665194418674",
        "GASPRICE": "5468568668644106462",
        "GASLIMIT": 3600359985866212000,
        "METHOD": 366435164561523700,
        "PARAMS": "a0p3UXNyem9WVw=="
    }
]
```

Input and enrichment data are intentionally non-optimal to test and show how different edge cases are handled

## End-to-end latency of streams(topics)
* **BLOCKS_STREAM** (intermediate stream) - 3ms
* **BLOCKS_STREAM_AVRO** (raw block's data sinking to Postgres) - 63ms
* **MESSAGES_STREAM** (intermediate stream to process data) - 3ms
* **MESSAGES_STREAM_AVRO** (parsed extrinsic's data sinking to Postgres) - 63ms

## How to see PoC?

Navigate to the Control Center web interface at http://localhost:9021/ and select your cluster.
![cluster](images/cluster.png "Cluster")

You'll see an overview of the cluster and some tabs:
![overview](images/overview.png "Overview")

* **Topics** tab contains all topics created by the script and streams. Click by a topic and you see a full statistic about it.
![topic](images/topic.png "Topic")

* **Connect** tab contains an info about all connectors in the system.
![connect](images/connect.png "Connect")

* **ksqlDB** tab provides an interface for working with queries, streams and tables
![ksqldb](images/ksqlDB.png "ksqlDB")
    * Sub-tab **Running queries** provides information about running queries in the system.
    ![queries](images/queries.png "Queries")
