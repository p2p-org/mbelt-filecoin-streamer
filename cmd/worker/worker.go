package worker

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
	"log"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	defaultHeight = 5000
	batchCapacity = 5

	// current event is current head. We receive it once right after subscription on head updates
	HeadEventCurrent = "current"
	HeadEventApply   = "apply"
	HeadEventRevert  = "revert"
)

func Start(conf *config.Config, sync bool, syncForce bool, updHead bool, syncFrom int, syncFromDbOffset int) {
	exitCode := 0
	defer os.Exit(exitCode)

	err := services.InitServices(conf)
	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		exitCode = 1
		return
	}

	syncCtx, syncCancel := context.WithCancel(context.Background())
	updCtx, updCancel := context.WithCancel(context.Background())

	go gracefulStop(syncCancel, updCancel)

	if syncForce {
		syncToHead(0, syncCtx)
	}

	if sync {
		if syncFrom >= 0 {
			syncToHead(syncFrom, syncCtx)
		} else {
			heightFromDb, err := services.App().BlocksService().GetMaxHeightFromDB()
			if err != nil {
				log.Println("Can't get max height from postgres DB, stopping...")
				log.Println(err)
				exitCode = 1
				return
			}

			if heightFromDb < syncFromDbOffset {
				syncToHead(0, syncCtx)
			} else {
				syncToHead(heightFromDb-syncFromDbOffset, syncCtx)
			}
		}
	}

	if updHead {
		updateHeads(updCtx)
	}

	log.Println("mbelt-filecoin-streamer gracefully stopped")
}

func updateHeads(ctx context.Context) {
	headUpdatesCtx, cancelHeadUpdates := context.WithCancel(ctx)
	// Buffer is that big for channel to be able to store some head updates while we are syncing till "current" block
	// TODO: This approach is not solid. Think how we can do it better.
	headUpdates := make(chan []*api.HeadChange, 5000)
	services.App().BlocksService().GetHeadUpdates(headUpdatesCtx, &headUpdates)

	for {
		select {
		case <-ctx.Done():
			cancelHeadUpdates()
			log.Println("[App][Debug][updateHeads]", "unsubscribed from head updates")
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
					maxHeightInDb, err := services.App().BlocksService().GetMaxHeightFromDB()
					if err != nil {
						log.Println("[App][Error][updateHeads]", "couldn't get max height from DB. Error:", err)
						cancelHeadUpdates()
						return
					}
					if currentHeight > maxHeightInDb {
						syncTo(maxHeightInDb, currentHeight, ctx)
					}
					// Pushing block and its messages to kafka just in case. Duplicates are handled by db.
					collectAndPushOtherEntitiesByTipSet(tipSet)

				case HeadEventRevert:
					services.App().TipSetsService().PushTipSetsToRevert(tipSet)

				case HeadEventApply:
					collectAndPushOtherEntitiesByTipSet(tipSet)

				default:
					log.Println("[App][Debug][updateHeads]", "yet unknown event encountered:", hu.Type)
					if hu.Val != nil {
						// Pushing just in case
						collectAndPushOtherEntitiesByTipSet(tipSet)
					}
				}
			}
		}
	}
}

func syncToHead(from int, ctx context.Context) {
	head := services.App().TipSetsService().GetHead()
	if head != nil {
		syncTo(from, int(head.Height()), ctx)
	} else {
		log.Println("[App][Debug][syncToHead]", "Head is nil!")
		syncTo(from, 0, ctx)
	}
}

func syncTo(from int, to int, ctx context.Context) {
	syncHeight := abi.ChainEpoch(to)
	if to <= from {
		log.Println("[App][Debug][sync]", "Specified sync height is too small, syncing to default height:", defaultHeight)
		syncHeight = defaultHeight
	}

	defer log.Println("[App][Debug][sync]", "finished sync")

	startHeight := abi.ChainEpoch(from)
	if startHeight <= 1 {
		log.Println("getting genesis")
		genesis := services.App().TipSetsService().GetGenesis()
		services.App().TipSetsService().PushNormalState(genesis)
		services.App().BlocksService().Push(genesis.Blocks())
	}

	for height := startHeight; height < syncHeight; {
		select {
		default:
			wg := sync.WaitGroup{}
			wg.Add(batchCapacity)

			for workers := 0; workers < batchCapacity; workers++ {

				go func(height abi.ChainEpoch) {
					defer wg.Done()
					tipSet, nullRound := syncTipSetForHeight(height)

					if tipSet == nil {
						log.Println("[App][Error][sync]", "Tipset is nil! Height:", height)
						return
					}

					if nullRound {
						tipSet.State = tipsets.StateNull
						services.App().TipSetsService().Push(tipSet)
						return
					}

					collectAndPushOtherEntitiesByTipSet(tipSet)
				}(height)

				height++
			}

			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func syncTipSetForHeight(height abi.ChainEpoch) (*tipsets.TipSetWithState, bool) {
	log.Println("[App][Debug]", "Load height:", height)

	tipSet, isHeightNotReached := services.App().TipSetsService().GetByHeight(height)

	tipSetWithState := &tipsets.TipSetWithState{
		TipSet: tipSet,
		State:  tipsets.StateNormal,
	}

	if tipSet != nil && tipSet.Height() < height {
		log.Println("[App][Debug]", "Got null tipset on height:", height)

		// sorry for pointer arithmetics magic but I need to change received tipsets height (which is unexported)
		//to requested height without a lot of useless code only to solve this
		p := unsafe.Pointer(tipSetWithState.TipSet)
		*(*abi.ChainEpoch)(unsafe.Pointer(uintptr(p) + unsafe.Sizeof(tipSet.Cids()) + unsafe.Sizeof(tipSet.Blocks()))) = height

		tipSetWithState.State = tipsets.StateNull

		return tipSetWithState, true
	}

	if !isHeightNotReached {
		log.Println("[App][Debug]", "Height reached")
		return nil, false
	}

	return tipSetWithState, false
}

// Getting this tipsets messages and previous tipsets receipts
func getMessagesAndReceipts(tipSet *types.TipSet) (msgs []*messages.MessageExtended, rcpts []*messages.MessageReceiptWithCid) {
	if len(tipSet.Blocks()) > 0 {
		firstBlockCid := tipSet.Blocks()[0].Cid()
		tipSetMsgs := services.App().MessagesService().GetParentMessages(firstBlockCid)
		receipts := services.App().MessagesService().GetParentReceipts(firstBlockCid)
		for i, msg := range tipSetMsgs {
			rcpts = append(rcpts, &messages.MessageReceiptWithCid{
				Cid:       msg.Cid,
				Receipt:   receipts[i],
			})
		}
	}

	for _, block := range tipSet.Blocks() {
		blockCid := block.Cid()
		blockMessages := services.App().MessagesService().GetBlockMessages(blockCid)

		if blockMessages == nil || len(blockMessages.BlsMessages) == 0 {
			continue
		}

		for _, msg := range blockMessages.BlsMessages {
			msgs = append(msgs, &messages.MessageExtended{
				Cid:       msg.Cid(),
				BlockCid:  blockCid,
				Message:   msg,
				Timestamp: block.Timestamp,
			})
		}

		for _, msg := range blockMessages.SecpkMessages {
			msgs = append(msgs, &messages.MessageExtended{
				Cid:       msg.Cid(),
				BlockCid:  blockCid,
				Message:   &msg.Message,
				Timestamp: block.Timestamp,
			})
		}

	}

	return msgs, rcpts
}

func collectActorChanges(tipset *types.TipSet) (out []*state.ActorInfo, nullRounds []types.TipSetKey,
	minerInfo []*state.MinerInfo, minerSectors []*state.MinerSector, reward *state.RewardActor) {

	start := time.Now()
	defer func() {
		log.Println("Collected Actor Changes", "duration:", time.Since(start).String(), "actors count:", len(out),
			"miner info count:", len(minerInfo), "miner sectors count:", len(minerSectors), "reward actor:", reward != nil)
	}()

	miners := services.App().StateService().ListMiners(tipset.Key())
	minersMap := make(map[address.Address]struct{}, len(miners))
	for _, miner := range miners {
		minersMap[miner] = struct{}{}
	}

	parentTipSet := services.App().TipSetsService().GetByKey(tipset.Parents())
	if parentTipSet == nil {
		log.Println("[App][Debug][collectActorChanges] parent is nil. height: ", tipset.Height())
		return nil, nil, nil, nil, nil
	}
	if parentTipSet.ParentState().Equals(tipset.ParentState()) {
		nullRounds = append(nullRounds, parentTipSet.Key())
	}

	// collect all actors that had state changes between the tipset's parent-state and its grandparent-state.
	// TODO: changes will contain deleted actors, this causes needless processing further down the pipeline, consider
	// a separate strategy for deleted actors
	// (these comments as well as basic logic were copied from lotus/cmd/lotus-chainwatch/processor/processor.go)
	changes := services.App().StateService().GetChangedActors(parentTipSet.ParentState(), tipset.ParentState())

	out = make([]*state.ActorInfo, 0, len(changes))
	actorsSeen := map[cid.Cid]struct{}{}

	// record the state of all actors that have changed
	for a, act := range changes {
		var deleted bool
		has, err := services.App().StateService().ChainHasObj(act.Head)
		if err != nil {
			log.Println("[App][Error][collectActorChanges]", err)
		}
		if !has {
			deleted = true
		}

		addr, err := address.NewFromString(a)
		if err != nil {
			log.Println("[App][Error][collectActorChanges]", err)
			continue
		}

		// miner info collection
		if _, ok := minersMap[addr]; ok {
			info := services.App().StateService().GetMinerInfo(addr, tipset.Key())
			power := services.App().StateService().GetMinerPower(addr, tipset.Key())
			sectors := services.App().StateService().GetMinerSectors(addr, tipset.Key())
			minerInfo = append(minerInfo, &state.MinerInfo{
				MinerInfo:  info,
				MinerPower: power,
				Miner:      addr,
				Height:     tipset.Height(),
			})
			for _, sector := range sectors {
				minerSectors = append(minerSectors, &state.MinerSector{
					SectorOnChainInfo: sector,
					Miner:             addr,
					Height:            tipset.Height(),
				})
			}
			// We can skip the rest of loop if we don't want miner's account states to be collected.
			// continue
		}

		ast := services.App().StateService().ReadState(addr, parentTipSet.Key())

		if ast == nil {
			log.Println("[App][Error][collectActorChanges]", "empty state!")
			continue
		}

		actorState, err := json.Marshal(ast.State)
		if err != nil {
			log.Println("[App][Error][collectActorChanges]", err)
			continue
		}

		// parse reward
		if addr == builtin.RewardActorAddr {
			rewardState := parseRewardActorState(ast.State.(map[string]interface{}))
			reward = &state.RewardActor{
				Act:         act,
				StateRoot:   tipset.ParentState(),
				TsKey:       parentTipSet.Key(),
				ParentTsKey: parentTipSet.Parents(),
				Addr:        addr,
				State:       rewardState,
			}
		}

		if _, ok := actorsSeen[act.Head]; !ok {
			out = append(out, &state.ActorInfo{
				Act:         act,
				StateRoot:   tipset.ParentState(),
				Height:      tipset.Height(),
				TsKey:       parentTipSet.Key(),
				ParentTsKey: parentTipSet.Parents(),
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

func collectAndPushOtherEntitiesByTipSet(tipset *tipsets.TipSetWithState) {
	services.App().TipSetsService().Push(tipset)
	services.App().BlocksService().Push(tipset.Blocks())

	msgs, receipts := getMessagesAndReceipts(tipset.TipSet)
	services.App().MessagesService().Push(msgs)
	services.App().MessagesService().PushReceipts(receipts)

	// ignoring null rounds
	changes, _, minersInfo, minersSectors, rewardStates := collectActorChanges(tipset.TipSet)
	services.App().StateService().PushActors(changes)
	services.App().StateService().PushMinersInfo(minersInfo)
	services.App().StateService().PushMinersSectors(minersSectors)
	services.App().StateService().PushRewardActorStates(rewardStates)
}

func gracefulStop(syncCancel, updCancel context.CancelFunc) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGHUP)

	sig := <-gracefulStop
	log.Printf("Caught sig: %+v", sig)
	log.Println("Wait for graceful shutdown to finish.")
	syncCancel()
	updCancel()
}
