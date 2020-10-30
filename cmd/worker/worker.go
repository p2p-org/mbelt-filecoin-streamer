package worker

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	defaultHeight = 5000
	batchCapacity = 20

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

	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGHUP)

		sig := <-gracefulStop
		log.Printf("Caught sig: %+v", sig)
		log.Println("Wait for graceful shutdown to finish.")
		syncCancel()
		updCancel()
	}()

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
	headUpdates := make(chan []*api.HeadChange, 1000)
	services.App().BlocksService().GetHeadUpdates(headUpdatesCtx, &headUpdates)

	for {
		select {
		case update := <-headUpdates:
			for _, hu := range update {
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
					// Pushing block and its messages to kafka just in case.
					// TODO: Duplicates should be handled on db's side
					pushTipsetWithBlocksAndMessagesAndActorChanges(hu.Val)

				case HeadEventRevert:
					services.App().TipSetsService().PushTipSetsToRevert(hu.Val)

				case HeadEventApply:
					pushTipsetWithBlocksAndMessagesAndActorChanges(hu.Val)

				default:
					log.Println("[App][Debug][updateHeads]", "yet unknown event encountered:", hu.Type)
					// Pushing just in case
					pushTipsetWithBlocksAndMessagesAndActorChanges(hu.Val)
				}
			}
		case <-ctx.Done():
			cancelHeadUpdates()
			log.Println("[App][Debug][updateHeads]", "unsubscribed from head updates")
			return
		}
	}
}

func syncToHead(from int, ctx context.Context) {
	head := services.App().TipSetsService().GetHead()
	if head != nil {
		syncTo(from, int(head.Height()), ctx)
	} else {
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
		log.Println(genesis.MarshalJSON())
		services.App().TipSetsService().Push(genesis)
		services.App().BlocksService().Push(genesis.Blocks())
		services.App().MessagesService().Push(getBlockMessages(genesis.Blocks()))
	}

	for height := startHeight; height < syncHeight; {
		select {
		default:
			wg := sync.WaitGroup{}
			wg.Add(batchCapacity)

			for workers := 0; workers < batchCapacity; workers++ {

				go func(height abi.ChainEpoch) {
					defer wg.Done()
					_, tipSet, blocks, msgs, changes := syncForHeight(height)
					services.App().TipSetsService().Push(tipSet)
					services.App().BlocksService().Push(blocks)
					services.App().MessagesService().Push(msgs)
					services.App().StateService().PushActors(changes)
				}(height)

				height++
			}

			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func syncForHeight(height abi.ChainEpoch) (isHeightNotReached bool, tipSet *types.TipSet, blocks []*types.BlockHeader, extMessages []*messages.MessageExtended, changes []*state.ActorInfo) {
	log.Println("[Datastore][Debug]", "Load height:", height)

	tipSet, isHeightNotReached = services.App().TipSetsService().GetByHeight(height)

	if !isHeightNotReached {
		log.Println("[App][Debug]", "Height reached")
		return
	}

	// Empty TipSet, skipping
	if tipSet == nil {
		return
	}

	blocks = tipSet.Blocks()
	extMessages = getBlockMessages(blocks)

	// ignoring null rounds
	changes, _ = collectActorChangesForBlocks(blocks)

	return
}

func getBlockMessages(blocks []*types.BlockHeader) (msgs []*messages.MessageExtended) {
	for _, block := range blocks {
		blockMessages := services.App().MessagesService().GetBlockMessages(block.Cid())

		if blockMessages == nil || len(blockMessages.BlsMessages) == 0 {
			continue
		}

		for _, blsMessage := range blockMessages.BlsMessages {
			msgs = append(msgs, &messages.MessageExtended{
				BlockCid:  block.Cid(),
				Message:   blsMessage,
				Timestamp: block.Timestamp,
			})
		}

	}

	return msgs
}

func collectActorChanges(block *types.BlockHeader) (out []*state.ActorInfo, nullRounds []types.TipSetKey) {
	//start := time.Now()
	//defer func() {
	//	log.Println("Collected Actor Changes", "duration", time.Since(start).String(), "len", len(out))
	//}()

	parentTipSet := services.App().TipSetsService().GetByKey(types.NewTipSetKey(block.Parents...))
	if parentTipSet == nil {
		log.Println("[App][Debug][collectActorChanges] parent is nil. height: ", block.Height)
		return nil, nil
	}
	if parentTipSet.ParentState().Equals(block.ParentStateRoot) {
		nullRounds = append(nullRounds, parentTipSet.Key())
	}

	// collect all actors that had state changes between the block's parent-state and its grandparent-state.
	// TODO: changes will contain deleted actors, this causes needless processing further down the pipeline, consider
	// a separate strategy for deleted actors
	// (these comments as well as algorithm were copied from lotus/cmd/lotus-chainwatch/processor/processor.go)
	changes := services.App().StateService().GetChangedActors(parentTipSet.ParentState(), block.ParentStateRoot)

	out = make([]*state.ActorInfo, 0, len(changes))
	actorsSeen := map[cid.Cid]struct{}{}

	// record the state of all actors that have changed
	for a, act := range changes {
		// ignore actors that were deleted. (Do we actually need to ignore them?)
		has, err := services.App().StateService().ChainHasObj(act.Head)
		if err != nil {
			log.Println("[App][Error][collectActorChanges]", err)
		}
		if !has {
			continue
		}

		addr, err := address.NewFromString(a)
		if err != nil {
			log.Println("[App][Error][collectActorChanges]", err)
			continue
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

		if _, ok := actorsSeen[act.Head]; !ok {
			out = append(out, &state.ActorInfo{
				Act:         act,
				StateRoot:   block.ParentStateRoot,
				Height:      block.Height,
				TsKey:       parentTipSet.Key(),
				ParentTsKey: parentTipSet.Parents(),
				Addr:        addr,
				State:       string(actorState),
			})
		}
		actorsSeen[act.Head] = struct{}{}
	}

	return out, nullRounds
}

func collectActorChangesForBlocks(blocks []*types.BlockHeader) (changes []*state.ActorInfo, nullRrounds []types.TipSetKey) {
	wg := sync.WaitGroup{}
	for _, block := range blocks {
		// collecting changes async five blocks at a time
		for i := 0; i < 5; i++ {
			go func() {
				wg.Add(1)
				blockChanges, blockNullRounds := collectActorChanges(block)
				changes = append(changes, blockChanges...)
				nullRrounds = append(nullRrounds, blockNullRounds...)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	return
}

func pushTipsetWithBlocksAndMessagesAndActorChanges(tipset *types.TipSet) {
	blocks := tipset.Blocks()
	services.App().TipSetsService().Push(tipset)
	services.App().BlocksService().Push(blocks)
	services.App().MessagesService().Push(getBlockMessages(blocks))
	changes, _ := collectActorChangesForBlocks(blocks)
	services.App().StateService().PushActors(changes)
}
