package core

import (
	"reflect"

	"github.com/izqui/helpers"
)

var (
	TRANSACTION_POW = helpers.ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	BLCOK_POW       = helpers.ArrayOfBytes(BLCOK_POW_COMPLEXITY, POW_PREFIX)
)

func CheckProofOfWork(prefix []byte, hash []byte) bool {
	if len(prefix) > 0 {
		return reflect.DeepEqual(prefix, hash[:len(prefix)])
	}
	return true
}
