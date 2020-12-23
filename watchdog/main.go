package watchdog

import (
	"context"
	"time"

	"github.com/p2p-org/mbelt-filecoin-streamer/services"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/p2p-org/mbelt-filecoin-streamer/client"

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
	maxBlockHeight  int
	maxTipsetHeight int
}

//
//func main() {
//	childFormatter := logrus.TextFormatter{}
//	runtimeFormatter := &runtime.Formatter{ChildFormatter: &childFormatter}
//	runtimeFormatter.Line = true
//	runtimeFormatter.File = true
//	logrus.SetFormatter(runtimeFormatter)
//	cfg, err := config.NewConfig(".env")
//	if err != nil {
//		logrus.Error(err)
//		return
//	}
//
//	InitWatcher(cfg)
//}

func InitWatcher(cfg *config.Config) {
	watcher, err := NewWatcher(cfg)
	if err != nil {
		logrus.Error(err)
		return
	}
	watcher.Start()
}

func (w *Watcher) Start() {
	w.startTime = time.Now()
	currentMaxHeight, err := w.db.GetMaxHeight()
	if err != nil {
		logrus.Error(err)
	}
	//pp.Println("currentMaxHeight from DB ", currentMaxHeight)

	w.cs.maxBlockHeight = currentMaxHeight
	currentMaxHeightTipsets, err := w.db.GetMaxHeightOfTipsets()
	if err != nil {
		logrus.Error(err)
	}
	w.cs.maxTipsetHeight = currentMaxHeightTipsets

	w.checkTipsetsConsistency(currentMaxHeight)
}

func (w *Watcher) checkTipsetsConsistency(currentHeigth int) {
	ctx := context.Background()
	tipsetGenesis, err := w.db.GetGenesisTipset(ctx)
	if err != nil {
		logrus.Error(err)
		return
	}
	//updatedTipset, ok := w.api.GetByHeight(abi.ChainEpoch(9))
	//pp.Println(ok)
	//pp.Println(updatedTipset.String())
	//pp.Println(updatedTipset.ParentState())
	//pp.Println(updatedTipset.Blocks())
	//pp.Println(updatedTipset.Cids())
	//pp.Println(updatedTipset.Height().String())
	//return
	pp.Println(tipsetGenesis)

	for i := 0; i < w.cs.maxTipsetHeight; i++ {
		currentTipset, err := w.db.GetTipsetByHeight(ctx, int64(i))
		if err != nil {
			logrus.Error(err)
			return
		}
		if currentTipset != nil {
			switch {
			case len(currentTipset.Blocks) == 0:
				logrus.Warn("empty tipset with height: ", i)
				updatedTipset, ok := w.api.GetByHeight(abi.ChainEpoch(i))
				pp.Println(ok)
				pp.Println(updatedTipset)
				w.ss.TipSetsService().PushNormalState(updatedTipset)
			}
		}
		pp.Println(currentTipset)

		//block, err := w.db.GetBlockByHeight(ctx, int64(i))
		//if err != nil {
		//	if strings.Contains(err.Error(), "no rows in result set") {
		//		castedCID, err := cid.Cast([]byte(block.Cid))
		//		if err != nil {
		//			logrus.Error(err, ", cid.Cast : ", block.Cid)
		//			return
		//		}
		//		missingBlock := w.api.GetBlock(castedCID)
		//		pp.Println(missingBlock)
		//	}
		//	logrus.Error(err, ", bad block height: ", i)
		//	return
		//}
		//if i == 0 {
		//	continue
		//}
		//parentBlock, err := w.db.GetParentBlockByCID(ctx, block.Cid)
		//if err != nil {
		//	if strings.Contains(err.Error(), "no rows in result set") {
		//		castedCID, err := cid.Cast([]byte(block.Cid))
		//		if err != nil {
		//			logrus.Error(err, ", cid.Cast : ", castedCID)
		//			return
		//		}
		//		missingBlock := w.api.GetBlock(castedCID)
		//		pp.Println(missingBlock)
		//	}
		//	logrus.Error(err, ", bad block height, cid: ", i, block.Cid)
		//	return
		//}

		//pp.Println("block cid: ", block.Cid)
		//pp.Println("block parents: ", block.Parents)
		//pp.Println("parent block cid: ", parentBlock.Cid)
		//pp.Println("parent block parents: ", parentBlock.Parents)
		//
		if i > 5 {
			break
		}

	}
}

func (w *Watcher) checkBlockConsistency() {

}
