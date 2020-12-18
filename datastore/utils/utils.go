package utils

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/mr-tron/base58"
	"strconv"
	"strings"
)

// Temporary function for Postgres inserts formatting
func CidsToVarcharArray(elems []cid.Cid) string {
	var result strings.Builder

	result.WriteString(`{`)

	last := len(elems) - 1
	for i := range elems {
		result.WriteString(`"`)
		result.WriteString(elems[i].String())
		result.WriteString(`"`)

		if i != last {
			result.WriteString(`, `)
		}
	}

	result.WriteString(`}`)

	return result.String()
}

func AddressesToVarcharArray(elems []address.Address) string {
	var result strings.Builder

	result.WriteString(`{`)

	last := len(elems) - 1
	for i := range elems {
		result.WriteString(`"`)
		result.WriteString(elems[i].String())
		result.WriteString(`"`)

		if i != last {
			result.WriteString(`, `)
		}
	}

	result.WriteString(`}`)

	return result.String()
}

func MultiaddrsToVarcharArray(elems []abi.Multiaddrs) string {
	var result strings.Builder

	result.WriteString(`{`)

	last := len(elems) - 1
	for i := range elems {
		result.WriteString(`"`)
		result.WriteString(base58.Encode(elems[i]))
		result.WriteString(`"`)

		if i != last {
			result.WriteString(`, `)
		}
	}

	result.WriteString(`}`)

	return result.String()
}

func DealIdsToIntArray(elems []abi.DealID) string {
	var result strings.Builder

	result.WriteString(`{`)

	last := len(elems) - 1
	for i := range elems {
		result.WriteString(strconv.FormatUint(uint64(elems[i]), 10))

		if i != last {
			result.WriteString(`, `)
		}
	}

	result.WriteString(`}`)

	return result.String()
}