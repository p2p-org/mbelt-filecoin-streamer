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
* **FIlecoinBlocksEnrichment** - golang service that listens for fresh blocks in Kafka topic and for each one queries and puts corresponding messages and deals in corresponding Kafka topics

## What is happening inside ksqlDB?
Data flow:
![flow](images/flow.png "Data flow")

1. Data produced by FilecoinBlockStreamer gets into Kafka Topic **filecoin_blocks** and further into ksqlDB stream **FILECOIN_BLOCKS**;
    1. A stream called **FILECOIN_BLOCKS_AVRO** collects raw data from **FILECOIN_BLOCKS** with AVRO format and sends it to topic **FILECOIN_BLOCKS_AVRO** to sink it to a Postgres table **FILECOIN_BLOCK_RAW**;
2. Data produced by FIlecoinBlocksEnrichment gets into Kafka topic **FILECOIN_BLOCK_MESSAGES** and further into KSQLDB stream **FILECOIN_BLOCK_MESSAGES**:
    1. **FILECOIN_BLOCK_MESSAGES** formats the data to AVRO and sends it to a topic **FILECOIN_BLOCK_MESSAGES_AVRO** to sink it to a Postgres table **FILECOIN_BLOCK_MESSAGE**;

## How does data sink to a DB?

KSQLDB has a special JDBCSinkConnector.

It collects data from some Kafka topic and inserts the data to a DB table.

## Message format examples

* **FILECOIN_BLOCKS** and **FILECOIN_BLOCKS_AVRO** topics:
 ```json

```


* **FILECOIN_BLOCK_MESSAGES** AND **FILECOIN_BLOCK_MESSAGES_AVRO** topics:

```json

```

Input and enrichment data are intentionally non-optimal to test and show how different edge cases are handled

## End-to-end latency of streams(topics)
* **FILECOIN_BLOCKS** (intermediate stream) - XXXms
* **FILECOIN_BLOCKS_AVRO** (raw block's data sinking to Postgres) - XXXms
* **FILECOIN_BLOCK_MESSAGES** (intermediate stream to process data) - XXXms
* **FILECOIN_BLOCK_MESSAGES_AVRO** (parsed extrinsic's data sinking to Postgres) - XXXXms

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
