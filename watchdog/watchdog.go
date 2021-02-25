package watchdog

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/p2p-org/mbelt-filecoin-streamer/client"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/k0kubun/pp"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/sirupsen/logrus"
)

const (
	ChainType        = "filecoin"
	defaultMinHeight = 0
)

type Watcher struct {
	cfg       *config.Config
	db        *pg.PgDatastore
	api       *client.APIClient
	ss        *services.ServiceProvider
	cs        *CurrentStatus
	startTime time.Time
}

func NewWatcher(cfg *config.Config) (*Watcher, error) {
	w := &Watcher{}
	pgDs, err := pg.Init(cfg)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	apiClient, err := client.Init(cfg.APIUrl, cfg.APIWsUrl, cfg.APIToken)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	childFormatter := logrus.TextFormatter{}
	runtimeFormatter := &runtime.Formatter{ChildFormatter: &childFormatter}
	runtimeFormatter.Line = true
	runtimeFormatter.File = true
	logrus.SetFormatter(runtimeFormatter)

	err = services.InitServices(cfg)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	w.ss = services.App()
	w.db = pgDs
	w.cfg = cfg
	w.api = apiClient
	w.cs = new(CurrentStatus)
	return w, nil

}

type CurrentStatus struct {
	sync.RWMutex
	maxBlockHeight    int
	startBlockHeight  int
	maxTipsetHeight   int
	lastCheckedHeight int
	skipped           []skippedEntity
}
type skippedEntity struct {
	entity     interface{}
	skipReason string
}

type checkedHeights struct {
	lastTipsetHeight  int
	lastBlockHeight   int
	lastMessageHeight int
}

func (w *Watcher) SetLastCheckedHeight(height int) {
	w.cs.Lock()
	w.cs.lastCheckedHeight = height
	w.cs.Unlock()
}

func (w *Watcher) GetLastCheckedHeight() int {
	return w.cs.lastCheckedHeight
}

func (w *Watcher) AddSkippedEntity(entity interface{}, reason string) {
	w.cs.Lock()
	w.cs.skipped = append(w.cs.skipped, skippedEntity{
		entity:     entity,
		skipReason: reason,
	})
	w.cs.Unlock()
}

func InitWatcher(cfg *config.Config, startHeight int) {
	logrus.Info("InitWatcher startHeight: ", startHeight)
	watcher, err := NewWatcher(cfg)
	if err != nil {
		logrus.Error(err)
		return
	}
	watcher.Start(startHeight)
}

//Параметры watchdog
//При старте инициализировать watermark высоты проверенных блоков, например
//дефолтные значения `watchdog_verify_height=0`, и сравнивать с max(height)
//(последним значением типсета в бд), а при достижении определенной высоты,
//инкрементировать.  В качестве опции, в случае, если была перезапись этих
//блоков (DELETE, UPDATE), сбрасывать это значение через триггер.
//
//2. Доработать watchdog, с параметрами запуска, к примеру
//`--verify  --start 888`, где start  - это высота с которой
//начнётся верификация. Если start > watchdog_verify_height,
//вернуть ошибку, чтобы не было пропусков. Если меньше,
//watchdog_verify_height будет сброшен до start и начнёт
//инкрементироваться в соответсвии с прогрессом верификатора.
//При верификации необходимо загружать данные с ноды, сверять
//их с загруженными в базу
//( количество сущностей, их hash (cid),  parents).
//В случае ошибок подгрузить необходимые данные.
func (w *Watcher) Start(startHeight int) {

	w.cs.startBlockHeight = startHeight

	w.startTime = time.Now()

	currentMaxHeight, err := w.db.GetMaxHeight()
	if err != nil {
		logrus.Error(err)
		return
	}
	defer func() {
		w.db.SaveLastCheckedHeight(w.GetLastCheckedHeight())
	}()

	w.cs.maxBlockHeight = currentMaxHeight
	currentMaxHeightTipsets, err := w.db.GetMaxHeightOfTipsets()
	if err != nil {
		logrus.Error(err)
		return
	}
	w.cs.maxTipsetHeight = currentMaxHeightTipsets

	if startHeight > currentMaxHeightTipsets {
		logrus.Error("start height more than current stored tipset height")
		return
	}
	ctx := context.Background()
	var wg sync.WaitGroup
	//wg.Add(1)
	//go func() {
	//	w.checkTipsetsConsistency(ctx, w.cs.startBlockHeight, w.cs.maxTipsetHeight)
	//	wg.Done()
	//}()
	wg.Add(1)
	go func() {
		//w.startBlockChecker(ctx, w.cs.startBlockHeight, w.cs.maxTipsetHeight)
		w.startBlockChecker(ctx, w.cs.startBlockHeight, 6)
		wg.Done()
	}()
	wg.Wait()

	pp.Println(w.cs)
}

func (w *Watcher) checkTipsetsConsistency(ctx context.Context, fromHeight, toHeight int) {

	for i := fromHeight; i < w.cs.maxTipsetHeight; i++ {
		defer func() {
			w.cs.lastCheckedHeight = i
		}()
		currentTipset, err := w.db.GetTipsetByHeight(ctx, int64(i))
		if err != nil {
			logrus.Error(err)
			continue
		}
		if currentTipset.State == 1 {
			continue
		}

		//pp.Println(currentTipset)
		//
		//if i > 5 {
		//	break
		//}

	}
}

func (w *Watcher) startBlockChecker(ctx context.Context, fromHeight, toHeight int) {

	for i := fromHeight; i < toHeight; i++ {
		defer func() {
			w.SetLastCheckedHeight(i)
		}()
		block, err := w.db.GetBlockByHeight(ctx, int64(i))
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				logrus.Warn("block not found, with height: ", i, "; trying to retrieve")
				// as we didn't found block in db with height, we must find it,s
				// ancestor
				if block.Cid == "" || block == nil {
					pp.Println("empty cid of block with height: ", i)
					pp.Println("trying to retrieve from blockchain node")
					missingTipset, ok := w.api.GetByHeight(abi.ChainEpoch(i))
					if !ok {
						pp.Println("w.api.GetByHeight is not OK, height: ", i)
					}
					w.ss.BlocksService().Push(missingTipset.Blocks())
					nextParentHeight := i + 1
					parent, err := w.db.GetParentBlockByHeight(ctx, int64(nextParentHeight))
					if err != nil {
						logrus.Error("can't find parent by height: ", nextParentHeight, err)
						continue
					}
					nextBlock, err := w.db.GetBlockByHeight(ctx, int64(nextParentHeight))
					if err != nil {
						logrus.Error("can't find parent by height: ", i, err)
						continue
					}
					logrus.Warn(fmt.Sprintf("parent cid: %s \n parent height: %d \n missing block height: %d \n,next existing block height: %d \n next existing block cid: %s \n",
						parent.Cid,
						parent.Height,
						i,
						nextParentHeight,
						nextBlock.Cid))
					w.AddSkippedEntity(pg.Block{Height: int64(i)}, "block missing in DB and API")
					continue
				}
				castedCID, err := cid.Cast([]byte(block.Cid))
				if err != nil {
					logrus.Error(err, ", cid.Cast : ", block.Cid)
					continue
				}
				missingBlock := w.api.GetBlock(castedCID)
				w.ss.BlocksService().Push([]*types.BlockHeader{missingBlock})
				continue
			}
			//logrus.Error(err, ", bad block height: ", i)
			continue
		}

		parentBlock, err := w.db.GetParentBlockByCID(ctx, block.Cid)
		if err != nil {
			logrus.Error(err, ", bad parentBlock height, cid: ", i, block.Cid)
			if strings.Contains(err.Error(), "no rows in result set") {
				castedCID, err := cid.Cast([]byte(block.Cid))
				if err != nil {
					logrus.Error(err, ", cid.Cast : ", castedCID)
					continue
				}
				missingBlock := w.api.GetBlock(castedCID)
				w.ss.BlocksService().Push([]*types.BlockHeader{missingBlock})
			}
			//logrus.Error(err, ", bad block height, cid: ", i, block.Cid)
			//return
		}
		_ = parentBlock
		//pp.Println("block cid: ", block.Cid)
		//pp.Println("block parents: ", block.Parents)
		//pp.Println("parent block cid: ", parentBlock.Cid)
		//pp.Println("parent block parents: ", parentBlock.Parents)
	}

	//
}

func (w *Watcher) checkBlockConsistency(ctx context.Context, blockHeight int) {
	block, err := w.db.GetBlockByHeight(ctx, int64(blockHeight))
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			logrus.Warn("block not found, with height: ", blockHeight, "; trying to retrieve")
			// as we didn't found block in db with height, we must find it,s
			// ancestor
			if block.Cid == "" || block == nil {
				pp.Println("empty cid of block with height: ", blockHeight)
				pp.Println("trying to retrieve from blockchain node")
				missingTipset, ok := w.api.GetByHeight(abi.ChainEpoch(blockHeight))
				if !ok {
					pp.Println("w.api.GetByHeight is not OK, height: ", blockHeight)
				}
				w.ss.BlocksService().Push(missingTipset.Blocks())
				nextParentHeight := blockHeight + 1
				parent, err := w.db.GetParentBlockByHeight(ctx, int64(nextParentHeight))
				if err != nil {
					logrus.Error("can't find parent by height: ", nextParentHeight, err)
					return
				}
				nextBlock, err := w.db.GetBlockByHeight(ctx, int64(nextParentHeight))
				if err != nil {
					logrus.Error("can't find parent by height: ", blockHeight, err)
					return
				}
				logrus.Warn(fmt.Sprintf("parent cid: %s \n parent height: %d \n missing block height: %d \n,next existing block height: %d \n next existing block cid: %s \n",
					parent.Cid,
					parent.Height,
					blockHeight,
					nextParentHeight,
					nextBlock.Cid))
				w.AddSkippedEntity(pg.Block{Height: int64(blockHeight)}, "block missing in DB and API")
				return
			}
			castedCID, err := cid.Cast([]byte(block.Cid))
			if err != nil {
				logrus.Error(err, ", cid.Cast : ", block.Cid)
				return
			}
			missingBlock := w.api.GetBlock(castedCID)
			w.ss.BlocksService().Push([]*types.BlockHeader{missingBlock})
			return
		}
		//logrus.Error(err, ", bad block height: ", blockHeight)
		return
	}
}

func (w *Watcher) checkMessageConsistency(ctx context.Context, fromHeight, toHeight int) {

	for i := fromHeight; i < w.cs.maxTipsetHeight; i++ {
		defer func() {
			w.SetLastCheckedHeight(i)
		}()

		block, err := w.db.GetBlockByHeight(ctx, int64(i))
		if err != nil {
			time.Sleep(1 * time.Second)
			block, err = w.db.GetBlockByHeight(ctx, int64(i))
			if err != nil {
				logrus.Error("while checking messages, can't get block from DB by height: ", i, "; error: ", err)
				continue
			}
			continue
		}
		_, err = w.db.GetMessageByBlockCid(ctx, block.Cid)
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				castedCID, err := cid.Cast([]byte(block.Cid))
				if err != nil {
					logrus.Error(err, ", cid.Cast : ", castedCID)
					continue
				}
				bMessages := w.api.GetBlockMessages(castedCID)
				w.ss.MessagesService().Push(pg.ParseMessageExtended(bMessages.BlsMessages, block))
				continue
			}

		}

	}
}
