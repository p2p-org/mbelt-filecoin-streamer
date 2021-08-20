package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"strconv"
	"strings"
	"time"
)

const (
	RosettaVersion = "1.4.10"
	MbeltFilecoinRosettaApiVersion = "0.0.1"

	versionUpdateInterval = 5 * time.Minute
	statusUpdateInterval = 5 * time.Minute
)

var CurrentNetworkId *types.NetworkIdentifier

type NetworkAPI struct {
	networkName       string
	networkVersion    int
	versionLastUpdate int64

	statusResponse   *types.NetworkStatusResponse
	statusLastUpdate int64
}

func NewNetworkAPI() server.NetworkAPIServicer {
	api := &NetworkAPI{networkName: services.App().StateService().NetworkName()}
	api.updateNetworkVersion()
	return api
}

func (s *NetworkAPI) NetworkList(
	ctx context.Context, request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	CurrentNetworkId = &types.NetworkIdentifier{Blockchain: "Filecoin", Network: s.networkName}
	return &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{CurrentNetworkId}}, nil
}

func (s *NetworkAPI) NetworkStatus(
	ctx context.Context, request *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {
	if time.Now().Sub(time.Unix(s.statusLastUpdate, 0)) >= statusUpdateInterval {
		return s.statusResponse, nil
	}

	currTs, err :=  services.App().PgDatastore().GetMaxHeightTipSet()
	if err != nil {
		return nil, common.NewErrorWithMessage(common.TipSetNotFoundError,
			fmt.Sprintf("unable to get currrent tipset: %v", err.Error()))
	}

	oldHeight, oldBlocks, err := services.App().PgDatastore().GetMinHeightTipSetBlocks()
	if err != nil {
		return nil, common.NewErrorWithMessage(common.TipSetNotFoundError,
			fmt.Sprintf("unable to get oldest tipset: %v", err.Error()))
	}
	oldBlkHash := fmt.Sprintf("[%s]", strings.Join(oldBlocks, ","))
	oldBlkId := &types.BlockIdentifier{
		Index: oldHeight,
		Hash:  oldBlkHash,
	}

	var genesisID *types.BlockIdentifier
	if oldHeight == 0 {
		genesisID = oldBlkId
	}

	nodePeers := services.App().StateService().NetPeers()
	chainHead := services.App().TipSetsService().GetHead()

	peers := make([]*types.Peer, 0, len(nodePeers))
	for _, peer := range nodePeers {
		peers = append(peers, &types.Peer{
			PeerID: peer.ID.String(),
			Metadata: map[string]interface{}{
				"topics": peer.Addrs,
			},
		})
	}

	targetHeight := int64(chainHead.Height())
	syncStatus := &types.SyncStatus{
		CurrentIndex: &currTs.Height,
		TargetIndex:  &targetHeight,
	}

	syncStage := "blocks sync"
	synced := false
	if targetHeight - currTs.Height < 3 {
		syncStage = "chain synced and following head updates"
		synced = true
	}

	syncStatus.Stage = &syncStage
	syncStatus.Synced = &synced

	statusResponse := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: currTs.Height,
			Hash:  fmt.Sprintf("[%s]", strings.Join(currTs.Blocks, ",")),
		},
		CurrentBlockTimestamp:  currTs.MinTs,
		GenesisBlockIdentifier: genesisID,
		OldestBlockIdentifier:  oldBlkId,
		SyncStatus:             syncStatus,
		Peers:                  peers,
	}

	s.statusResponse = statusResponse
	s.statusLastUpdate = time.Now().Unix()

	return statusResponse, nil
}

// TODO rework cache (not thread-safe now at least) and add BalanceExemptions
func (s *NetworkAPI) NetworkOptions(
	ctx context.Context, request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {
	s.updateNetworkVersion()
	v := MbeltFilecoinRosettaApiVersion
	version := &types.Version{
		RosettaVersion:    RosettaVersion,
		NodeVersion:       strconv.Itoa(s.networkVersion),
		MiddlewareVersion: &v,
	}

	operationStatuses := common.OperationsStatuses()

	operationTypes := services.MethodsList()

	errors := common.AllErrors()

	minTimestamp, err := services.App().PgDatastore().GetMinBlockTimestamp()
	if err != nil {
		return nil, common.NewErrorWithMessage(common.TipSetNotFoundError,
			fmt.Sprintf("unable to get min timestamp from tipsets: %v", err.Error()))
	}

	allow := &types.Allow{
		OperationStatuses: operationStatuses,
		OperationTypes: operationTypes,
		Errors: errors,
		HistoricalBalanceLookup: true,
		TimestampStartIndex: &minTimestamp,
		//TODO //BalanceExemptions:
		MempoolCoins: false,
	}

	return &types.NetworkOptionsResponse{Version: version, Allow: allow}, nil
}

func (s *NetworkAPI) updateNetworkVersion() {
	if s.versionLastUpdate == 0 || time.Now().Sub(time.Unix(s.versionLastUpdate, 0)) >= versionUpdateInterval {
		s.networkVersion = services.App().StateService().NetworkVersion(nil)
		s.versionLastUpdate = time.Now().Unix()
	}

	return
}

func assertValidNetworkIdentifier(netID *types.NetworkIdentifier) *types.Error {
	if netID == nil || types.Hash(CurrentNetworkId) != types.Hash(netID) {
		return &common.InvalidNetworkError
	}
	return nil

}