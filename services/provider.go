package services

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/blocks"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/processor"
)

var (
	provider ServiceProvider
)

type ServiceProvider struct {
	blocksService    *blocks.BlocksService
	messagesService  *messages.MessagesService
	processorService *processor.ProcessorService
}

func (p *ServiceProvider) Init(config *config.Config, ds *datastore.Datastore, apiClient *client.APIClient) error {
	var err error

	p.blocksService, err = blocks.Init(config, ds, apiClient)

	if err != nil {
		return err
	}

	p.messagesService, err = messages.Init(config, ds, apiClient)

	if err != nil {
		return err
	}

	p.processorService, err = processor.Init(config, ds, apiClient)

	if err != nil {
		return err
	}

	return nil
}

func (p *ServiceProvider) BlocksService() *blocks.BlocksService {
	return p.blocksService
}

func (p *ServiceProvider) MessagesService() *messages.MessagesService {
	return p.messagesService
}

func (p *ServiceProvider) ProcessorService() *processor.ProcessorService {
	return p.processorService
}

func App() *ServiceProvider {
	return &provider
}
