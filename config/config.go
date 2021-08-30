package config

type Config struct {
	APIUrl          string
	APIWsUrl        string
	APIToken        string
	KafkaPrefix     string
	KafkaHosts      string
	KafkaAsyncWrite bool
	PgUrl           string
	LokiUrl         string
	LokiSourceName  string
}
