package worker

import (
	"errors"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"sync"
)

const (
	actorNameMiner = "fil/1/storageminer"

	methodSend        = "Send"
	methodConstructor = "Constructor"
)

var (
	addressMap         sync.Map
	idAddressMap       sync.Map
	addressIdToTypeMap sync.Map
)

var addrTypeToHumanMap = map[string]string{
	"fil/1/system":           "system",
	"fil/1/init":             "init",
	"fil/1/cron":             "cron",
	"fil/1/storagepower":     "storage power",
	"fil/1/storageminer":     "miner",
	"fil/1/storagemarket":    "storage market",
	"fil/1/paymentchannel":   "payment channel",
	"fil/1/reward":           "reward",
	"fil/1/verifiedregistry": "verified registry",
	"fil/1/account":          "account",
	"fil/1/multisig":         "multisig",
}

// https://github.com/filecoin-project/specs-actors/blob/v3.0.0/actors/builtin/methods.go
var accountMethods = map[abi.MethodNum]string{2: "PubkeyAddress"}
var initMethods    = map[abi.MethodNum]string{2: "Exec"}
var cronMethods    = map[abi.MethodNum]string{2: "EpochTick"}

var rewardMethods = map[abi.MethodNum]string{
	2: "AwardBlockReward",
	3: "ThisEpochReward",
	4: "UpdateNetworkKPI",
}

var multisigMethods = map[abi.MethodNum]string{
	2: "Propose",
	3: "Approve",
	4: "Cancel",
	5: "AddSigner",
	6: "RemoveSigner",
	7: "SwapSigner",
	8: "ChangeNumApprovalsThreshold",
	9: "LockBalance",
}

var paychMethods = map[abi.MethodNum]string{
	2: "UpdateChannelState",
	3: "Settle",
	4: "Collect",
}

var marketMethods = map[abi.MethodNum]string{
	2: "AddBalance",
	3: "WithdrawBalance",
	4: "PublishStorageDeals",
	5: "VerifyDealsForActivation",
	6: "ActivateDeals",
	7: "OnMinerSectorsTerminate",
	8: "ComputeDataCommitment",
	9: "CronTick",
}

var powerMethods = map[abi.MethodNum]string{
	2: "CreateMiner",
	3: "UpdateClaimedPower",
	4: "EnrollCronEvent",
	5: "OnEpochTickEnd",
	6: "UpdatePledgeTotal",
	7: "Deprecated1",
	8: "SubmitPoRepForBulkVerify",
	9: "CurrentTotalPower",
}

var minerMethods = map[abi.MethodNum]string{
	2:  "ControlAddresses",
	3:  "ChangeWorkerAddress",
	4:  "ChangePeerID",
	5:  "SubmitWindowedPoSt",
	6:  "PreCommitSector",
	7:  "ProveCommitSector",
	8:  "ExtendSectorExpiration",
	9:  "ExtendSectorExpiration",
	10: "DeclareFaults",
	11: "DeclareFaultsRecovered",
	12: "OnDeferredCronEvent",
	13: "CheckSectorProven",
	14: "ApplyRewards",
	15: "ReportConsensusFault",
	16: "WithdrawBalance",
	17: "ConfirmSectorProofsValid",
	18: "ChangeMultiaddrs",
	19: "CompactPartitions",
	20: "CompactSectorNumbers",
	21: "ConfirmUpdateWorkerKey",
	22: "RepayDebt",
	23: "ChangeOwnerAddress",
	24: "DisputeWindowedPoSt",
}

var verifiedRegistryMethods = map[abi.MethodNum]string{
	2: "AddVerifier",
	3: "RemoveVerifier",
	4: "AddVerifiedClient",
	5: "UseBytes",
	6: "RestoreBytes",
}

var addrTypeToMethods = map[string]map[abi.MethodNum]string{
	"fil/1/init":             initMethods,
	"fil/1/cron":             cronMethods,
	"fil/1/storagepower":     powerMethods,
	"fil/1/storageminer":     minerMethods,
	"fil/1/storagemarket":    marketMethods,
	"fil/1/paymentchannel":   paychMethods,
	"fil/1/reward":           rewardMethods,
	"fil/1/verifiedregistry": verifiedRegistryMethods,
	"fil/1/account":          accountMethods,
	"fil/1/multisig":         multisigMethods,
}

func getMethodName(addrType string, methodNum abi.MethodNum) string {
	return addrTypeToMethods[addrType][methodNum]
}

func addrTypeToHuman(tp string) string {
	return addrTypeToHumanMap[tp]
}

func getAddressType(addr address.Address, tsk *types.TipSetKey) (string, error) {
	if addr.Protocol() != address.ID {
		return "", errors.New("address should be of id protocol")
	}

	raw, ok := addressMap.Load(addr.String())
	if ok {
		switch raw.(type) {
		case string:
			return raw.(string), nil
		default:
			return "", errors.New("failed to cast address type to string")
		}
	}

	act := services.App().StateService().GetActor(addr, tsk)
	if act == nil {
		return "", errors.New("Couldn't get actor " + addr.String())
	}
	actorName := builtin.ActorNameByCode(act.Code)
	addressIdToTypeMap.Store(addr.String(), actorName)

	return actorName, nil
}

func addAddressType(addr address.Address, tp string) {
	if addr.Protocol() != address.ID {
		return
	}
	addressIdToTypeMap.Store(addr.String(), tp)
}

func lookupIdAddress(addr address.Address, tsk *types.TipSetKey) *address.Address {
	if addr.Protocol() == address.ID {
		return &addr
	}

	raw, ok := addressMap.Load(addr.String())
	if ok {
		switch raw.(type) {
		case address.Address:
			return raw.(*address.Address)
		}
	}

	id := services.App().StateService().LookupID(addr, tsk)
	if id != nil {
		addressMap.Store(addr.String(), id)
		idAddressMap.Store(id.String(), &addr)
	}

	return id
}

func lookupAccountKeyByAddress(id address.Address, tsk *types.TipSetKey) *address.Address {
	if id.Protocol() == address.BLS {
		return &id
	}
	raw, ok := idAddressMap.Load(id.String())
	if ok {
		switch raw.(type) {
		case address.Address:
			return raw.(*address.Address)
		}
	}

	addr := services.App().StateService().AccountKey(id, tsk)
	if addr != nil {
		idAddressMap.Store(id.String(), addr)
		addressMap.Store(addr.String(), &id)
	}

	return addr
}
