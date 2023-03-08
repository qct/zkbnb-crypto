package txtypes

import (
	"hash"
	"math/big"
)

type TxInfo interface {
	GetTxType() int

	Validate() error

	VerifySignature(pubKey string) error

	GetAccountIndex() int64

	GetFromAccountIndex() int64

	GetToAccountIndex() int64

	GetNonce() int64

	GetExpiredAt() int64

	Hash(hFunc hash.Hash) (msgHash []byte, err error)

	GetGas() (int64, int64, *big.Int)
}
