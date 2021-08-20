package services

import (
	"errors"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"sync"
)

const (
	actorNameMiner = "storageminer"

	methodSend        = "Send"
	methodConstructor = "Constructor"
)

var (
	addressMap         sync.Map
	idAddressMap       sync.Map
	addressIdToTypeMap sync.Map
    builtinActors      map[cid.Cid]string
)

var addrTypeToHumanMap = map[string]string{
	"system":           "system",
	"init":             "init",
	"cron":             "cron",
	"storagepower":     "storage power",
	"storageminer":     "miner",
	"storagemarket":    "storage market",
	"paymentchannel":   "payment channel",
	"reward":           "reward",
	"verifiedregistry": "verified registry",
	"account":          "account",
	"multisig":         "multisig",
}

var humanToAddrTypeMap = map[string]string{
	"system":           "system",
	"init":             "init",
	"cron":             "cron",
	"storagepower":     "storage power",
	"storageminer":     "miner",
	"storagemarket":    "storage market",
	"paymentchannel":   "payment channel",
	"reward":           "reward",
	"verifiedregistry": "verified registry",
	"account":          "account",
	"multisig":         "multisig",
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
	"init":             initMethods,
	"cron":             cronMethods,
	"storagepower":     powerMethods,
	"storageminer":     minerMethods,
	"storagemarket":    marketMethods,
	"paymentchannel":   paychMethods,
	"reward":           rewardMethods,
	"verifiedregistry": verifiedRegistryMethods,
	"account":          accountMethods,
	"multisig":         multisigMethods,
}

func init() {
	builder := cid.V1Builder{Codec: cid.Raw, MhType: mh.IDENTITY}
	builtinActors = make(map[cid.Cid]string)

	for i := 1; i < 10; i++ {
		for _, name := range []string{"system", "init", "cron", "storagepower", "storageminer", "storagemarket",
			"paymentchannel", "reward", "verifiedregistry", "account", "multisig",
		} {
			c, err := builder.Sum([]byte(fmt.Sprintf("fil/%d/%s", i, name)))
			if err != nil {
				panic(err)
			}
			builtinActors[c] = name
		}
	}
}

// IsBuiltinActor returns true if the code belongs to an actor defined in this repo.
func IsBuiltinActor(code cid.Cid) bool {
	_, isBuiltin := builtinActors[code]
	return isBuiltin
}

// ActorNameByCode returns the (string) name of the actor given a cid code.
func ActorNameByCode(code cid.Cid) string {
	if !code.Defined() {
		return "<undefined>"
	}

	name, ok := builtinActors[code]
	if !ok {
		return "<unknown>"
	}
	return name
}

func getMethodName(addrType string, methodNum abi.MethodNum) string {
	if methodNum == 0 {
		return methodConstructor
	} else if methodNum == 1 {
		return methodSend
	}
	return addrTypeToMethods[addrType][methodNum]
}

func addrTypeToHuman(tp string) string {
	return addrTypeToHumanMap[tp]
}

func humanToAddrType(tp string) string {
	return humanToAddrTypeMap[tp]
}

func getAddressType(addr address.Address, tsk *types.TipSetKey) (string, error) {
	if addr.Protocol() != address.ID {
		return "", errors.New("address should be of id protocol")
	}

	raw, ok := addressIdToTypeMap.Load(addr.String())
	if ok {
		switch raw.(type) {
		case string:
			return raw.(string), nil
		default:
			return "", errors.New("failed to cast address type to string")
		}
	}

	act := App().StateService().GetActor(addr, tsk)
	if act == nil {
		return "", errors.New("Couldn't get actor " + addr.String())
	}
	actorName := ActorNameByCode(act.Code)
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

	id := App().StateService().LookupID(addr, tsk)
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

	addr := App().StateService().AccountKey(id, tsk)
	if addr != nil {
		idAddressMap.Store(id.String(), addr)
		addressMap.Store(addr.String(), &id)
	}

	return addr
}

func MethodsList() []string {
	methods := []string{methodConstructor, methodSend}
	for _, v := range accountMethods {
		methods = append(methods, v)
	}
	for _, v := range initMethods {
		methods = append(methods, v)
	}
	for _, v := range cronMethods {
		methods = append(methods, v)
	}
	for _, v := range rewardMethods {
		methods = append(methods, v)
	}
	for _, v := range multisigMethods {
		methods = append(methods, v)
	}
	for _, v := range marketMethods {
		methods = append(methods, v)
	}
	for _, v := range powerMethods {
		methods = append(methods, v)
	}
	for _, v := range minerMethods {
		methods = append(methods, v)
	}
	for _, v := range verifiedRegistryMethods {
		methods = append(methods, v)
	}
	return methods
}
