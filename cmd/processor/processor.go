package main

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
)

var conf *config.Config

func init() {
	/*conf = &config.Config{
		APIUrl: "ws://116.203.240.62:1234/rpc/v0",
		APIToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.3b1-x0KwOB1-60NzHNGsMdMyzr6b0kokzh-Z4bc400Y",
		KafkaHosts: "localhost:9092",
	}*/

	conf = &config.Config{
		APIUrl:     "ws://127.0.0.1:1234/rpc/v0",
		KafkaHosts: "localhost:9092",
	}
}

func main() {
	err := services.InitServices(conf)

	if err != nil {
		log.Fatal(err)
	}

	services.App().ProcessorService().ProcessStreams()

}
