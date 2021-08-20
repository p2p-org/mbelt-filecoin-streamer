package common

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
)

var (
	CatchAllError = types.Error{
		Code:      0,
		Message:   "catch all error",
		Retriable: false,
	}

	MarshallingError = types.Error{
		Code:        1,
		Message:     "error during json marshalling/unmarshalling",
		Retriable:   false,
	}

	SanityCheckError = types.Error{
		Code:      2,
		Message:   "sanity check error",
		Retriable: false,
	}

	InvalidNetworkError = types.Error{
		Code:      3,
		Message:   "invalid network error",
		Retriable: false,
	}

	TipSetNotFoundError = types.Error{
		Code:      4,
		Message:   "tipset not found error",
		Retriable: false,
	}

	BlockNotFoundError = types.Error{
		Code:      5,
		Message:   "block not found error",
		Retriable: false,
	}

	TransactionNotFoundError = types.Error{
		Code:      6,
		Message:   "transaction (aka message in filecoin) not found",
		Retriable: false,
	}

	AccountNotFoundError = types.Error{
		Code:      7,
		Message:   "account not found error",
		Retriable: false,
	}

	NotImplementedError = types.Error{
		Code:      8,
		Message:   "method nod implemented error",
		Retriable: false,
	}
)

func NewError(rosettaError types.Error, detailStructure interface{}) *types.Error {
	newError := rosettaError
	details, err := types.MarshalMap(detailStructure)
	if err != nil {
		newError.Details = map[string]interface{}{
			"message": fmt.Sprintf("unable to get error details: %v", err.Error()),
		}
	} else {
		newError.Details = details
	}
	return &newError
}

func NewErrorWithMessage(rosettaError types.Error, message string) *types.Error {
	newError := rosettaError
	newError.Details = map[string]interface{}{
		"message": message,
	}
	return &newError
}

func NewRosettaErrorFromError(rosettaError types.Error, err error) *types.Error {
	newError := rosettaError
	newError.Details = map[string]interface{}{
		"message": err.Error(),
	}
	return &newError
}

func AllErrors() []*types.Error {
	return []*types.Error{&CatchAllError, &SanityCheckError, &InvalidNetworkError, &TipSetNotFoundError,
		&BlockNotFoundError, &TransactionNotFoundError}
}

