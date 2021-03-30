package services

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v3/actors/builtin"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
	"log"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	defaultHeight           = 5000
	batchCapacity uint32    = 12

	// current event is current head. We receive it once right after subscription on head updates
	HeadEventCurrent = "current"
	HeadEventApply   = "apply"
	HeadEventRevert  = "revert"
)

type SyncService struct {
	config  *config.Config
	ds      *datastore.KafkaDatastore
}

func Init(config *config.Config, ds *datastore.KafkaDatastore) (*SyncService, error) {
	return &SyncService{
		config:  config,
		ds:      ds,
	}, nil
}

func (s *SyncService) UpdateHeads(ctx context.Context) {
	log.Println("[Sync][Debug][UpdateHeads]", "subscribing on head updates...")

	headUpdatesCtx, cancelHeadUpdates := context.WithCancel(ctx)
	// Buffer is that big for channel to be able to store some head updates while we are syncing till "current" block
	// TODO: This approach is not solid. Think how we can do it better.
	headUpdates := make(chan []*api.HeadChange, 5000)
	App().BlocksService().GetHeadUpdates(headUpdatesCtx, &headUpdates)

	for {
		select {
		case <-ctx.Done():
			cancelHeadUpdates()
			log.Println("[Sync][Debug][UpdateHeads]", "unsubscribed from head updates")
			return
		case update := <-headUpdates:
			for _, hu := range update {
				// TODO: I don't know for sure (but I'm pretty confident) that we will not get null tipset via head updates subscription.
				// Shall we handle it here or maybe leave it to the watchdog (if there will be some)? In theory there should not be null tip sets in high epochs of mainnet.
				tipSet := &tipsets.TipSetWithState{
					TipSet: hu.Val,
					State:  tipsets.StateNormal,
				}

				switch hu.Type {

				case HeadEventCurrent:
					currentHeight := int(hu.Val.Height())
					maxHeightInDb, err := App().BlocksService().GetMaxHeightFromDB()
					if err != nil {
						log.Println("[Sync][Error][UpdateHeads]", "couldn't get max height from DB. Error:", err)
						cancelHeadUpdates()
						return
					}
					if currentHeight > maxHeightInDb {
						s.SyncTo(maxHeightInDb, currentHeight, ctx)
					}
					// Pushing block and its messages to kafka just in case. Duplicates are handled by db.
					s.CollectAndPushOtherEntitiesByTipSet(tipSet, ctx)

				case HeadEventRevert:
					App().TipSetsService().PushTipSetsToRevert(tipSet, ctx)

				case HeadEventApply:
					s.CollectAndPushOtherEntitiesByTipSet(tipSet, ctx)

				default:
					log.Println("[Sync][Debug][UpdateHeads]", "yet unknown event encountered:", hu.Type)
					if hu.Val != nil {
						// Pushing just in case
						s.CollectAndPushOtherEntitiesByTipSet(tipSet, ctx)
					}
				}
			}
		}
	}
}

func (s *SyncService) SyncToHead(from int, ctx context.Context) {
	head := App().TipSetsService().GetHead()
	if head != nil {
		s.SyncTo(from, int(head.Height()), ctx)
	} else {
		log.Println("[Sync][Debug][SyncToHead]", "Head is nil!")
		s.SyncTo(from, 0, ctx)
	}
}

func (s *SyncService) SyncTo(from int, to int, ctx context.Context) {
	syncHeight := abi.ChainEpoch(to)
	if to <= from {
		log.Println("[Sync][Debug][Sync]", "Specified sync height is too small, syncing to default height:", defaultHeight)
		syncHeight = defaultHeight
	}

	defer log.Println("[Sync][Debug][Sync]", "finished sync")

	startHeight := abi.ChainEpoch(from)
	if startHeight <= 1 {
		log.Println("getting genesis")
		genesis := App().TipSetsService().GetGenesis()
		App().TipSetsService().PushNormalState(genesis, ctx)
		App().BlocksService().Push(genesis.Blocks(), ctx)
	}

	wg := sync.WaitGroup{}
	var runningWorkers uint32

	for height := startHeight; height < syncHeight; {
		select {
		case <-ctx.Done():
			wg.Wait()
			return

		default:

			if atomic.LoadUint32(&runningWorkers) < batchCapacity {

				atomic.AddUint32(&runningWorkers, 1)
				wg.Add(1)

				go func(height abi.ChainEpoch) {
					defer func() {
						atomic.AddUint32(&runningWorkers, ^uint32(0))
						wg.Done()
					}()
					tipSet, nullRound := s.SyncTipSetForHeight(height)

					if tipSet == nil {
						log.Println("[Sync][Error][Sync]", "Tipset is nil! Height:", height)
						return
					}

					if nullRound {
						App().TipSetsService().Push(tipSet, ctx)
						return
					}

					s.CollectAndPushOtherEntitiesByTipSet(tipSet, ctx)
				}(height)

				height++
			}
		}
	}
}

func (s *SyncService) SyncTipSetForHeight(height abi.ChainEpoch) (*tipsets.TipSetWithState, bool) {
	log.Println("[Sync][Debug]", "Load height:", height)

	tipSet, isHeightNotReached := App().TipSetsService().GetByHeight(height)

	tipSetWithState := &tipsets.TipSetWithState{
		TipSet: tipSet,
		State:  tipsets.StateNormal,
	}

	if tipSet != nil && tipSet.Height() < height {
		log.Println("[Sync][Debug]", "Got null tipset on height:", height)

		// sorry for pointer arithmetics magic but I need to change received tipsets height (which is unexported)
		//to requested height without a lot of useless code only to solve this
		p := unsafe.Pointer(tipSetWithState.TipSet)
		*(*abi.ChainEpoch)(unsafe.Pointer(uintptr(p) + unsafe.Sizeof(tipSet.Cids()) + unsafe.Sizeof(tipSet.Blocks()))) = height

		tipSetWithState.State = tipsets.StateNull

		return tipSetWithState, true
	}

	if !isHeightNotReached {
		log.Println("[Sync][Debug]", "Height reached")
		return nil, false
	}

	return tipSetWithState, false
}

// Getting this tipsets messages and previous tipsets receipts
func (s *SyncService) GetMessagesAndReceipts(tipSet *types.TipSet) (msgs []*messages.MessageExtended, rcpts []*messages.MessageReceiptWithCid) {
	if len(tipSet.Blocks()) > 0 {
		firstBlockCid := tipSet.Blocks()[0].Cid()
		tipSetMsgs := App().MessagesService().GetParentMessages(firstBlockCid)
		receipts := App().MessagesService().GetParentReceipts(firstBlockCid)
		for i, msg := range tipSetMsgs {
			rcpts = append(rcpts, &messages.MessageReceiptWithCid{
				Cid:       msg.Cid,
				Receipt:   receipts[i],
			})
		}
	}

	//tsk := types.NewTipSetKey(tipSet.Cids()...)

	for _, block := range tipSet.Blocks() {
		blockCid := block.Cid()
		blockMessages := App().MessagesService().GetBlockMessages(blockCid)

		if blockMessages == nil || len(blockMessages.BlsMessages) + len(blockMessages.SecpkMessages) == 0 {
			continue
		}

		for _, msg := range blockMessages.BlsMessages {
			fromId := lookupIdAddress(msg.From, nil)
			toId   := lookupIdAddress(msg.To, nil)

			var fromType, toType string
			if fromId != nil {
				fromType, _ = getAddressType(*fromId, nil)
			}
			if toId != nil {
				toType, _   = getAddressType(*toId, nil)
			}

			methodName := getMethodName(toType, msg.Method)
			msgs = append(msgs, &messages.MessageExtended{
				Cid:           msg.Cid(),
				Height:        block.Height,
				BlockCid:      blockCid,
				Message:       msg,
				FromId:        fromId,
				ToId:          toId,
				FromType:      addrTypeToHuman(fromType),
				ToType:        addrTypeToHuman(toType),
				MethodName:    methodName,
				ParentBaseFee: block.ParentBaseFee,
				Timestamp:     block.Timestamp,
			})
		}

		for _, msg := range blockMessages.SecpkMessages {
			fromId := lookupIdAddress(msg.Message.From, nil)
			toId   := lookupIdAddress(msg.Message.To, nil)

			var fromType, toType string
			if fromId != nil {
				fromType, _ = getAddressType(*fromId, nil)
			}
			if toId != nil {
				toType, _   = getAddressType(*toId, nil)
			}

			methodName := getMethodName(toType, msg.Message.Method)
			msgs = append(msgs, &messages.MessageExtended{
				Cid:           msg.Cid(),
				Height:        block.Height,
				BlockCid:      blockCid,
				Message:       &msg.Message,
				FromId:        fromId,
				ToId:          toId,
				FromType:      addrTypeToHuman(fromType),
				ToType:        addrTypeToHuman(toType),
				MethodName:    methodName,
				ParentBaseFee: block.ParentBaseFee,
				Timestamp:     block.Timestamp,
			})
		}

	}

	return msgs, rcpts
}

func (s *SyncService) CollectActorChanges(tipset *types.TipSet) (out []*state.ActorInfo, nullRounds []types.TipSetKey,
	minerInfo []*state.MinerInfo, minerSectors []*state.MinerSector, reward *state.RewardActor) {

	start := time.Now()
	defer func() {
		log.Println("Collected Actor Changes", "duration:", time.Since(start).String(), "actors count:", len(out),
			"miner info count:", len(minerInfo), "miner sectors count:", len(minerSectors), "reward actor:", reward != nil)
	}()

	parentTipSet := App().TipSetsService().GetByKey(tipset.Parents())
	if parentTipSet == nil {
		log.Println("[Sync][Debug][CollectActorChanges] parent is nil. height: ", tipset.Height())
		return nil, nil, nil, nil, nil
	}
	if parentTipSet.ParentState().Equals(tipset.ParentState()) {
		nullRounds = append(nullRounds, parentTipSet.Key())
		// TODO: probably need to return here
	}

	// collect all actors that had state changes between the tipset's parent-state and its grandparent-state.
	// TODO: changes will contain deleted actors, this causes needless processing further down the pipeline, consider
	// a separate strategy for deleted actors
	// (these comments as well as basic logic were copied from lotus/cmd/lotus-chainwatch/processor/processor.go)
	changes := App().StateService().GetChangedActors(parentTipSet.ParentState(), tipset.ParentState())

	out = make([]*state.ActorInfo, 0, len(changes))
	actorsSeen := map[cid.Cid]struct{}{}

	// record the state of all actors that have changed
	for a, act := range changes {
		var deleted bool
		has, err := App().StateService().ChainHasObj(act.Head)
		if err != nil {
			log.Println("[Sync][Error][CollectActorChanges]", err)
		}
		if !has {
			deleted = true
		}

		addr, err := address.NewFromString(a)
		if err != nil {
			log.Println("[Sync][Error][CollectActorChanges]", err)
			continue
		}

		actorName, err := getAddressType(addr, nil)
		if err != nil {
			log.Println("[Sync][Error][CollectActorChanges]", err)
		}

		// miner info collection
		if actorName == actorNameMiner {
			info := App().StateService().GetMinerInfo(addr, tipset.Key())
			power := App().StateService().GetMinerPower(addr, tipset.Key())
			// removed miners sectors collection because it takes too much resources to collect it like it's implemented here
			//sectors := services.App().StateService().GetMinerSectors(addr, tipset.Key())
			minerInfo = append(minerInfo, &state.MinerInfo{
				MinerInfo:  info,
				MinerPower: power,
				Miner:      addr,
				Height:     tipset.Height(),
			})
			//for _, sector := range sectors {
			//	minerSectors = append(minerSectors, &state.MinerSector{
			//		SectorOnChainInfo: sector,
			//		Miner:             addr,
			//		Height:            tipset.Height(),
			//	})
			//}
			// We can skip the rest of loop if we don't want miner's account states to be collected.
			//continue
		}

		ast := App().StateService().ReadState(addr, tipset.Key())

		if ast == nil {
			log.Println("[Sync][Error][CollectActorChanges]", "empty state!", "address:", addr.String())
			continue
		}

		actorState, err := json.Marshal(ast.State)
		if err != nil {
			log.Println("[Sync][Error][CollectActorChanges]", err)
			continue
		}

		// parse reward
		if addr == builtin.RewardActorAddr {
			rewardState := parseRewardActorState(ast.State.(map[string]interface{}))
			reward = &state.RewardActor{
				Act:         act,
				StateRoot:   tipset.ParentState(),
				TsKey:       tipset.Key(),
				ParentTsKey: tipset.Parents(),
				Addr:        addr,
				State:       rewardState,
			}
		}

		if _, ok := actorsSeen[act.Head]; !ok {
			out = append(out, &state.ActorInfo{
				Act:         act,
				StateRoot:   tipset.ParentState(),
				Height:      tipset.Height(),
				TsKey:       tipset.Key(),
				ParentTsKey: tipset.Parents(),
				Addr:        addr,
				State:       string(actorState),
				Deleted:     deleted,
			})
		}
		actorsSeen[act.Head] = struct{}{}
	}

	return
}

func parseRewardActorState(stateMap map[string]interface{}) *state.RewardActorState {
	cumsumBaseline, cumsumRealized, effectiveBaselinePower, thisEpochBaselinePower, thisEpochReward, totalMined,
	simpleTotal, baselineTotal, totalStoragePowerReward, positionEstimate, velocityEstimate := new(big.Int),
		new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int),
		new(big.Int), new(big.Int)

	var effectiveNetworkTime int = 0
	var epoch abi.ChainEpoch = 0

	if v, ok := stateMap["EffectiveNetworkTime"]; ok {
		switch v.(type) {
		case float64:
			effectiveNetworkTime = int(v.(float64))
		case int, int64, int32:
			effectiveNetworkTime = v.(int)
		}

	}
	if v, ok := stateMap["Epoch"]; ok {
		switch v.(type) {
		case float64:
			epoch = abi.ChainEpoch(v.(float64))
		case int, int64, int32, abi.ChainEpoch:
			epoch = v.(abi.ChainEpoch)
		}

	}

	if v, ok := stateMap["CumsumBaseline"]; ok {
		cumsumBaseline, _ = cumsumBaseline.SetString(v.(string), 10)
	}
	if v, ok := stateMap["CumsumRealized"]; ok {
		cumsumRealized, _ = cumsumRealized.SetString(v.(string), 10)
	}
	if v, ok := stateMap["EffectiveBaselinePower"]; ok {
		effectiveBaselinePower, _ = effectiveBaselinePower.SetString(v.(string), 10)
	}
	if v, ok := stateMap["ThisEpochBaselinePower"]; ok {
		thisEpochBaselinePower, _ = thisEpochBaselinePower.SetString(v.(string), 10)
	}
	if v, ok := stateMap["ThisEpochReward"]; ok {
		thisEpochReward, _ = thisEpochReward.SetString(v.(string), 10)
	}
	if v, ok := stateMap["SimpleTotal"]; ok {
		simpleTotal, _ = simpleTotal.SetString(v.(string), 10)
	}
	if v, ok := stateMap["BaselineTotal"]; ok {
		baselineTotal, _ = baselineTotal.SetString(v.(string), 10)
	}
	if v, ok := stateMap["TotalStoragePowerReward"]; ok {
		totalStoragePowerReward, _ = totalStoragePowerReward.SetString(v.(string), 10)
	}

	if m, ok := stateMap["ThisEpochRewardSmoothed"]; ok {
		thisEpochRewardSmoothed := m.(map[string]interface{})
		if v, ok := thisEpochRewardSmoothed["PositionEstimate"]; ok {
			positionEstimate, _ = positionEstimate.SetString(v.(string), 10)
		}
		if v, ok := thisEpochRewardSmoothed["VelocityEstimate"]; ok {
			velocityEstimate, _ = velocityEstimate.SetString(v.(string), 10)
		}
	}

	if _, ok := stateMap["TotalMined"]; ok {
		totalMined, _ = totalMined.SetString(stateMap["TotalMined"].(string), 10)
	}

	return &state.RewardActorState{
		CumsumBaseline:                          *cumsumBaseline,
		CumsumRealized:                          *cumsumRealized,
		EffectiveBaselinePower:                  *effectiveBaselinePower,
		EffectiveNetworkTime:                    effectiveNetworkTime,
		Epoch:                                   epoch,
		ThisEpochBaselinePower:                  *thisEpochBaselinePower,
		ThisEpochReward:                         *thisEpochReward,
		TotalMined:                              *totalMined,
		SimpleTotal:                             *simpleTotal,
		BaselineTotal:                           *baselineTotal,
		TotalStoragePowerReward:                 *totalStoragePowerReward,
		ThisEpochRewardSmoothedPositionEstimate: *positionEstimate,
		ThisEpochRewardSmoothedVelocityEstimate: *velocityEstimate,
	}
}

func (s *SyncService) CollectAndPushOtherEntitiesByTipSet(tipset *tipsets.TipSetWithState, ctx context.Context) {
	App().TipSetsService().Push(tipset, ctx)
	App().BlocksService().Push(tipset.Blocks(), ctx)

	// ignoring null rounds and sectors
	changes, _, minersInfo, _, rewardStates := s.CollectActorChanges(tipset.TipSet)
	App().StateService().PushActors(changes, ctx)
	App().StateService().PushMinersInfo(minersInfo, ctx)
	//services.App().StateService().PushMinersSectors(minersSectors)
	App().StateService().PushRewardActorStates(rewardStates, ctx)

	msgs, receipts := s.GetMessagesAndReceipts(tipset.TipSet)
	App().MessagesService().Push(msgs, ctx)
	App().MessagesService().PushReceipts(receipts, ctx)
}
