package state

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/utils"
	"log"
	"math/big"
)

type StateService struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

type ActorInfo struct {
	Act types.Actor

	StateRoot cid.Cid
	Height    abi.ChainEpoch // so that we can walk the actor changes in chronological order.

	TsKey       types.TipSetKey
	ParentTsKey types.TipSetKey

	Addr  address.Address
	State string
}

type MinerInfo struct {
	*miner.MinerInfo
	*api.MinerPower
	Miner  address.Address
	Height abi.ChainEpoch
}

type MinerSector struct {
	*miner.SectorOnChainInfo
	Miner  address.Address
	Height abi.ChainEpoch
}

type RewardActor struct {
	Act         types.Actor
	StateRoot   cid.Cid
	TsKey       types.TipSetKey
	ParentTsKey types.TipSetKey
	Addr        address.Address
	State       *RewardActorState
}

type RewardActorState struct {
	CumsumBaseline          big.Int
	CumsumRealized          big.Int
	EffectiveBaselinePower  big.Int
	EffectiveNetworkTime    int
	Epoch                   abi.ChainEpoch
	ThisEpochBaselinePower  big.Int
	ThisEpochReward         big.Int
	TotalMined              big.Int
	SimpleTotal             big.Int
	BaselineTotal           big.Int
	TotalStoragePowerReward big.Int

	ThisEpochRewardSmoothedPositionEstimate big.Int
	ThisEpochRewardSmoothedVelocityEstimate big.Int
}

func Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, apiClient *client.APIClient) (*StateService, error) {
	return &StateService{
		config: config,
		ds:     kafkaDs,
		api:    apiClient,
	}, nil
}

func (s *StateService) GetChangedActors(start, end cid.Cid) map[string]types.Actor {
	return s.api.GetChangedActors(start, end)
}

func (s *StateService) ChainHasObj(cid cid.Cid) (bool, error) {
	return s.api.ChainHasObj(cid)
}

func (s *StateService) ReadState(actor address.Address, tsk types.TipSetKey) *client.ActorState {
	return s.api.ReadState(actor, tsk)
}

func (s *StateService) ListMiners(tsk types.TipSetKey) []address.Address {
	return s.api.ListMiners(tsk)
}

func (s *StateService) GetMinerInfo(actor address.Address, tsk types.TipSetKey) *miner.MinerInfo {
	return s.api.GetMinerInfo(actor, tsk)
}

func (s *StateService) GetMinerPower(actor address.Address, tsk types.TipSetKey) *api.MinerPower {
	return s.api.GetMinerPower(actor, tsk)
}

func (s *StateService) GetMinerSectors(actor address.Address, tsk types.TipSetKey) []*miner.SectorOnChainInfo {
	return s.api.GetMinerSectors(actor, tsk)
}

func (s *StateService) PushActors(actors []*ActorInfo) {
	// Empty actor produces panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[StateService][Recover]", "Throw panic", r)
		}
	}()

	if actors == nil {
		return
	}

	m := map[string]interface{}{}

	for _, actor := range actors {
		m[actor.Addr.String()+"_"+actor.TsKey.String()] = serializeActor(actor)
	}

	s.ds.Push(datastore.TopicActorStates, m)
}

func (s *StateService) PushMinersInfo(minersInfo []*MinerInfo) {
	// Empty miner info produces panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[StateService][Recover]", "Throw panic", r)
		}
	}()

	if minersInfo == nil {
		return
	}

	m := map[string]interface{}{}

	for _, info := range minersInfo {
		m[fmt.Sprintf("%s_%d", info.Miner.String(), info.Height)] = serializeMinerInfo(info)
	}

	s.ds.Push(datastore.TopicMinerInfos, m)
}

func (s *StateService) PushMinersSectors(minersSectors []*MinerSector) {
	// Empty miner sector produces panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[StateService][Recover]", "Throw panic", r)
		}
	}()

	if minersSectors == nil {
		return
	}

	m := map[string]interface{}{}

	for _, sector := range minersSectors {
		m[fmt.Sprintf("%s_%d", sector.Miner.String(), sector.Height)] = serializeMinerSector(sector)
	}

	s.ds.Push(datastore.TopicMinerSectors, m)
}

func (s *StateService) PushRewardActorStates(rewardStates []*RewardActor) {
	// Empty reward actor produces panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[StateService][Recover]", "Throw panic", r)
		}
	}()

	if rewardStates == nil {
		return
	}

	m := map[string]interface{}{}

	for _, actor := range rewardStates {
		m[actor.State.Epoch.String()] = serializeRewardActor(actor)
	}

	s.ds.Push(datastore.TopicRewardActorStates, m)
}

func serializeActor(actor *ActorInfo) map[string]interface{} {

	result := map[string]interface{}{
		"actor_state_key":  actor.Addr.String() + "_" + actor.TsKey.String(),
		"actor_code":       actor.Act.Code.String(),
		"actor_head":       actor.Act.Head.String(),
		"nonce":            actor.Act.Nonce,
		"balance":          actor.Act.Balance,
		"state_root":       actor.StateRoot.String(),
		"height":           actor.Height,
		"ts_key":           actor.TsKey.String(),
		"parent_ts_key":    actor.ParentTsKey.String(),
		"addr":             actor.Addr.String(),
		"state":            actor.State,
	}

	return result
}

func serializeMinerInfo(info *MinerInfo) map[string]interface{} {

	result := map[string]interface{}{
		"miner_info_key":                fmt.Sprintf("%s_%d", info.Miner.String(), info.Height),
		"miner":                         info.Miner.String(),
		"owner":                         info.Owner.String(),
		"worker":                        info.Worker.String(),
		"control_addresses":             utils.AddressesToVarcharArray(info.ControlAddresses),
		"new_worker_address":            info.PendingWorkerKey.NewWorker.String(),
		"new_worker_effective_at":       info.PendingWorkerKey.EffectiveAt,
		"peer_id":                       string(info.PeerId),
		"multiaddrs":                    utils.MultiaddrsToVarcharArray(info.Multiaddrs),
		"seal_proof_type":               info.SealProofType,
		"sector_size":                   info.SectorSize,
		"window_post_partition_sectors": info.WindowPoStPartitionSectors,
		"miner_raw_byte_power":          info.MinerPower.MinerPower.RawBytePower,
		"miner_quality_adj_power":       info.MinerPower.MinerPower.QualityAdjPower,
		"total_raw_byte_power":          info.TotalPower.RawBytePower,
		"total_quality_adj_power":       info.TotalPower.QualityAdjPower,
		"height":                        info.Height,
	}

	return result
}

func serializeMinerSector(sector *MinerSector) map[string]interface{} {

	result := map[string]interface{}{
		"miner_sector_key":        fmt.Sprintf("%s_%d", sector.Miner.String(), sector.Height),
		"sector_number":           sector.SectorNumber,
		"seal_proof":              sector.SealProof,
		"sealed_cid":              sector.SealedCID,
		"deal_ids":                utils.DealIdsToIntArray(sector.DealIDs),
		"activation":              sector.Activation,
		"expiration":              sector.Expiration,
		"deal_weight":             sector.DealWeight,
		"verified_deal_weight":    sector.VerifiedDealWeight,
		"initial_pledge":          sector.InitialPledge,
		"expected_day_reward":     sector.ExpectedDayReward,
		"expected_storage_pledge": sector.ExpectedStoragePledge,
		"miner":                   sector.Miner.String(),
		"height":                  sector.Height,
	}

	return result
}

func serializeRewardActor(actor *RewardActor) map[string]interface{} {

	result := map[string]interface{}{
		"epoch":                      actor.State.Epoch,
		"actor_code":                 actor.Act.Code.String(),
		"actor_head":                 actor.Act.Head.String(),
		"nonce":                      actor.Act.Nonce,
		"balance":                    actor.Act.Balance,
		"state_root":                 actor.StateRoot.String(),
		"ts_key":                     actor.TsKey.String(),
		"parent_ts_key":              actor.ParentTsKey.String(),
		"addr":                       actor.Addr.String(),
		"cumsum_baseline":            actor.State.CumsumBaseline,
		"cumsum_realized":            actor.State.CumsumRealized,
		"effective_baseline_power":   actor.State.EffectiveBaselinePower,
		"effective_network_time":     actor.State.EffectiveNetworkTime,
		"this_epoch_baseline_power":  actor.State.ThisEpochBaselinePower,
		"this_epoch_reward":          actor.State.ThisEpochReward,
		"total_mined":                actor.State.TotalMined,
		"simple_total":               actor.State.SimpleTotal,
		"baseline_total":             actor.State.BaselineTotal,
		"total_storage_power_reward": actor.State.TotalStoragePowerReward,

		"this_epoch_reward_smoothed_position_estimate": actor.State.ThisEpochRewardSmoothedPositionEstimate,
		"this_epoch_reward_smoothed_velocity_estimate": actor.State.ThisEpochRewardSmoothedVelocityEstimate,
	}

	return result
}
