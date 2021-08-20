package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
	"strings"
)

type EventsAPI struct {
	ds *pg.PgDatastore
}

func NewEventsAPI(ds *pg.PgDatastore) server.EventsAPIServicer {
	return &EventsAPI{ds: ds}
}

func (s *EventsAPI) EventsBlocks(
	ctx context.Context, request *types.EventsBlocksRequest,
) (*types.EventsBlocksResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier); err != nil {
		return nil, err
	}
	var limit, offset int64 = 10, 0
	if request.Limit != nil {
		limit = *request.Limit
	}

	if request.Offset != nil {
		offset = *request.Offset
	}

	heights, keys, err := s.ds.GetTipSetsWithLimitAndOffset(limit, offset)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.TipSetNotFoundError, err)
	}

	events := make([]*types.BlockEvent, 0, len(heights))
	for i, h := range heights {
		events = append(events, &types.BlockEvent{
			Sequence:        h,
			BlockIdentifier: &types.BlockIdentifier{
				Index: h,
				Hash:  fmt.Sprintf("{%s}", strings.Join(keys[i], ",")),
			},
			Type:            types.ADDED,
		})
	}

	return &types.EventsBlocksResponse{
		MaxSequence: heights[len(heights)-1],
		Events:      events,
	}, nil
}

