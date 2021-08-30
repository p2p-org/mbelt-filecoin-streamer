package services

import (
	"github.com/afiskon/promtail-client/promtail"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/blocks"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/processor"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
	"log"
	"time"
)

const (
	blocksServiceLokiJob   = "blocks_service"
	messagesServiceLokiJob = "messages_service"
	tipsetsServiceLokiJob  = "tipsets_service"
	stateServiceLokiJob    = "state_service"
	syncServiceLokiJob     = "sync_service"
)

var (
	provider ServiceProvider
)

type ServiceProvider struct {
	syncService      *SyncService
	blocksService    *blocks.BlocksService
	messagesService  *messages.MessagesService
	tipsetsService   *tipsets.TipSetsService
	processorService *processor.ProcessorService
	stateService     *state.StateService
	pgDatastore      *pg.PgDatastore
}

func (p *ServiceProvider) Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, pgDs *pg.PgDatastore, apiClient *client.APIClient) error {
	var err error

	blocksLogger, err := InitLogger(config.LokiUrl, config.LokiSourceName, blocksServiceLokiJob)
	if err != nil {
		return err
	}
	p.blocksService, err = blocks.Init(config, kafkaDs, apiClient, blocksLogger)
	if err != nil {
		return err
	}

	messagesLogger, err := InitLogger(config.LokiUrl, config.LokiSourceName, messagesServiceLokiJob)
	if err != nil {
		return err
	}
	p.messagesService, err = messages.Init(config, kafkaDs, apiClient, messagesLogger)
	if err != nil {
		return err
	}

	tipsetsLogger, err := InitLogger(config.LokiUrl, config.LokiSourceName, tipsetsServiceLokiJob)
	if err != nil {
		return err
	}
	p.tipsetsService, err = tipsets.Init(config, kafkaDs, apiClient, tipsetsLogger)
	if err != nil {
		return err
	}

	p.processorService, err = processor.Init(config, kafkaDs, apiClient)
	if err != nil {
		return err
	}

	stateLogger, err := InitLogger(config.LokiUrl, config.LokiSourceName, stateServiceLokiJob)
	if err != nil {
		return err
	}
	p.stateService, err = state.Init(config, kafkaDs, apiClient, stateLogger)
	if err != nil {
		return err
	}

	syncLogger, err := InitLogger(config.LokiUrl, config.LokiSourceName, syncServiceLokiJob)
	if err != nil {
		return err
	}
	p.syncService, err = Init(config, kafkaDs, syncLogger)
	if err != nil {
		return err
	}

	p.pgDatastore = pgDs

	return nil
}

func (p *ServiceProvider) SyncService() *SyncService {
	return p.syncService
}

func (p *ServiceProvider) BlocksService() *blocks.BlocksService {
	return p.blocksService
}

func (p *ServiceProvider) MessagesService() *messages.MessagesService {
	return p.messagesService
}

func (p *ServiceProvider) TipSetsService() *tipsets.TipSetsService {
	return p.tipsetsService
}

func (p *ServiceProvider) ProcessorService() *processor.ProcessorService {
	return p.processorService
}

func (p *ServiceProvider) StateService() *state.StateService {
	return p.stateService
}

func (p *ServiceProvider) PgDatastore() *pg.PgDatastore {
	return p.pgDatastore
}

func App() *ServiceProvider {
	return &provider
}

func InitLogger(url, source, job string) (logger promtail.Client, err error) {
	labels := "{source=\""+source+"\",job=\""+job+"\"}"
	promtailConfig := promtail.ClientConfig{
		PushURL:            url,
		Labels:             labels,
		BatchWait:          5 * time.Second,
		BatchEntriesNumber: 10000,
		SendLevel: 			promtail.INFO,
		PrintLevel: 		promtail.ERROR,
	}

	if logger, err = promtail.NewClientProto(promtailConfig); err != nil {
		log.Println("Cannot init api client")
	}

	return logger, err
}

