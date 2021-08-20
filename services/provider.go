package services

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/blocks"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/processor"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
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

	p.blocksService, err = blocks.Init(config, kafkaDs, apiClient)

	if err != nil {
		return err
	}

	p.messagesService, err = messages.Init(config, kafkaDs, apiClient)

	if err != nil {
		return err
	}

	p.tipsetsService, err = tipsets.Init(config, kafkaDs, apiClient)

	if err != nil {
		return err
	}

	p.processorService, err = processor.Init(config, kafkaDs, apiClient)

	if err != nil {
		return err
	}

	p.stateService, err = state.Init(config, kafkaDs, apiClient)

	if err != nil {
		return err
	}

	p.syncService, err = Init(config, kafkaDs)

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
