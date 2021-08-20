package common

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
)

var (
	OperationStatusSuccess = &types.OperationStatus{
		Status: "Success",
		Successful: true,
	}

	OperationStatusFailure = &types.OperationStatus{
		Status: "Failure",
		Successful: false,
	}

	OperationStatusOutOfGas = &types.OperationStatus{
		Status: "Out Of Gas",
		Successful: false,
	}

	FILCurrency = &types.Currency{
		Symbol:   "FIL",
		Decimals: 18,
	}
)

func OperationsStatuses() []*types.OperationStatus {
	return []*types.OperationStatus{OperationStatusSuccess, OperationStatusFailure, OperationStatusOutOfGas}
}

func RosettaTransactionsFromMessagesFromDb(msgs []*messages.MessageFromDb) []*types.Transaction {
	trxs := make([]*types.Transaction, 0, 256)
	for _, msg := range msgs {
		trxs = append(trxs, RosettaTransactionFromMessageFromDb(msg))
	}

	return trxs
}

func RosettaTransactionFromMessageFromDb(msg *messages.MessageFromDb) *types.Transaction {
	// TODO: Add supposrt for other exit codes
	status := OperationStatusSuccess.Status
	if msg.ExitCode == 7 {
		status = OperationStatusOutOfGas.Status
	} else if msg.ExitCode != 0 {
		status = OperationStatusFailure.Status
	}

	acc := &types.AccountIdentifier{
		Address:  msg.From,
		Metadata: map[string]interface{}{"id": msg.FromId, "type": msg.FromType},
	}

	operation := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{Index: 0},
		RelatedOperations:   nil,
		Type:                msg.MethodName,
		Status:              &status,
		Account:             acc,
		Amount:              &types.Amount{
			Value:    msg.Value.String(),
			Currency: FILCurrency,
		},
	}

	trxMeta := map[string]interface{}{"to": msg.To, "to_id": msg.ToId, "to_type": msg.ToType}

	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: msg.Cid},
		Operations:            []*types.Operation{operation},
		Metadata:              trxMeta,
	}
}
