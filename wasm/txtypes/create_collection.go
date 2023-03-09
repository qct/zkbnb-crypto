/*
 * Copyright © 2022 ZkBNB Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package txtypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/bnb-chain/zkbnb-crypto/wasm/signature"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"hash"
	"log"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

type CreateCollectionSegmentFormat struct {
	AccountIndex      int64  `json:"account_index"`
	Name              string `json:"name"`
	Introduction      string `json:"introduction"`
	GasAccountIndex   int64  `json:"gas_account_index"`
	GasFeeAssetId     int64  `json:"gas_fee_asset_id"`
	GasFeeAssetAmount string `json:"gas_fee_asset_amount"`
	ExpiredAt         int64  `json:"expired_at"`
	Nonce             int64  `json:"nonce"`
}

/*
ConstructCreateCollectionTxInfo: construct mint nft tx, sign txInfo
*/
func ConstructCreateCollectionTxInfo(sk *PrivateKey, segmentStr string) (txInfo *CreateCollectionTxInfo, err error) {
	var segmentFormat *CreateCollectionSegmentFormat
	err = json.Unmarshal([]byte(segmentStr), &segmentFormat)
	if err != nil {
		log.Println("[ConstructCreateCollectionTxInfo] err info:", err)
		return nil, err
	}
	gasFeeAmount, err := StringToBigInt(segmentFormat.GasFeeAssetAmount)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	gasFeeAmount, _ = CleanPackedFee(gasFeeAmount)
	txInfo = &CreateCollectionTxInfo{
		AccountIndex:      segmentFormat.AccountIndex,
		Name:              segmentFormat.Name,
		Introduction:      segmentFormat.Introduction,
		GasAccountIndex:   segmentFormat.GasAccountIndex,
		GasFeeAssetId:     segmentFormat.GasFeeAssetId,
		GasFeeAssetAmount: gasFeeAmount,
		ExpiredAt:         segmentFormat.ExpiredAt,
		Nonce:             segmentFormat.Nonce,
		Sig:               nil,
	}
	// compute call data hash
	hFunc := mimc.NewMiMC()
	// compute msg hash
	msgHash, err := txInfo.Hash(hFunc)
	if err != nil {
		log.Println("[ConstructCreateCollectionTxInfo] unable to compute hash:", err)
		return nil, err
	}
	// compute signature
	hFunc.Reset()
	sigBytes, err := sk.Sign(msgHash, hFunc)
	if err != nil {
		log.Println("[ConstructCreateCollectionTxInfo] unable to sign:", err)
		return nil, err
	}
	txInfo.Sig = sigBytes
	return txInfo, nil
}

type CreateCollectionTxInfo struct {
	AccountIndex      int64
	CollectionId      int64
	Name              string
	Introduction      string
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
	L1Sig             string
}

func (txInfo *CreateCollectionTxInfo) Validate() error {
	// AccountIndex
	if txInfo.AccountIndex < minAccountIndex {
		return ErrAccountIndexTooLow
	}
	if txInfo.AccountIndex > maxAccountIndex {
		return ErrAccountIndexTooHigh
	}

	// Name
	if len(txInfo.Name) < minCollectionNameLength {
		return ErrCollectionNameTooShort
	}
	if len(txInfo.Name) > maxCollectionNameLength {
		return ErrCollectionNameTooLong
	}

	// Introduction
	if len(txInfo.Introduction) > maxCollectionIntroductionLength {
		return ErrIntroductionTooLong
	}

	// GasAccountIndex
	if txInfo.GasAccountIndex < minAccountIndex {
		return ErrGasAccountIndexTooLow
	}
	if txInfo.GasAccountIndex > maxAccountIndex {
		return ErrGasAccountIndexTooHigh
	}

	// GasFeeAssetId
	if txInfo.GasFeeAssetId < minAssetId {
		return ErrGasFeeAssetIdTooLow
	}
	if txInfo.GasFeeAssetId > maxAssetId {
		return ErrGasFeeAssetIdTooHigh
	}

	// GasFeeAssetAmount
	if txInfo.GasFeeAssetAmount == nil {
		return fmt.Errorf("GasFeeAssetAmount should not be nil")
	}
	if txInfo.GasFeeAssetAmount.Cmp(minPackedFeeAmount) < 0 {
		return ErrGasFeeAssetAmountTooLow
	}
	if txInfo.GasFeeAssetAmount.Cmp(maxPackedFeeAmount) > 0 {
		return ErrGasFeeAssetAmountTooHigh
	}

	// Nonce
	if txInfo.Nonce < minNonce {
		return ErrNonceTooLow
	}
	if len(txInfo.L1Sig) == 0 {
		return ErrL1SigInvalid
	}
	return nil
}

func (txInfo *CreateCollectionTxInfo) VerifySignature(pubKey string) error {
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := txInfo.Hash(hFunc)
	if err != nil {
		return err
	}
	// verify signature
	hFunc.Reset()
	pk, err := ParsePublicKey(pubKey)
	if err != nil {
		return err
	}
	isValid, err := pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		return err
	}

	if !isValid {
		return errors.New("invalid signature")
	}
	return nil
}

func (txInfo *CreateCollectionTxInfo) GetTxType() int {
	return TxTypeCreateCollection
}

func (txInfo *CreateCollectionTxInfo) GetAccountIndex() int64 {
	return txInfo.AccountIndex
}

func (txInfo *CreateCollectionTxInfo) GetFromAccountIndex() int64 {
	return txInfo.AccountIndex
}

func (txInfo *CreateCollectionTxInfo) GetToAccountIndex() int64 {
	return txInfo.AccountIndex
}

func (txInfo *CreateCollectionTxInfo) GetL1Signature() string {
	signatureBody := fmt.Sprintf(signature.SignatureTemplateCreateCollection, txInfo.AccountIndex,
		txInfo.Name, util.FormatWeiToEtherStr(txInfo.GasFeeAssetAmount), txInfo.GasAccountIndex, txInfo.Nonce)
	return signatureBody
}

func (txInfo *CreateCollectionTxInfo) GetL1AddressBySignatureInfo() (common.Address, common.Address) {
	message := accounts.TextHash([]byte(txInfo.L1Sig))
	//Decode from signature string to get the signature byte array
	signatureContent, err := hexutil.Decode(txInfo.GetL1Signature())
	if err != nil {
		return [20]byte{}, [20]byte{}
	}
	signatureContent[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	//Calculate the public key from the signature and source string
	signaturePublicKey, err := crypto.SigToPub(message, signatureContent)
	if err != nil {
		return [20]byte{}, [20]byte{}
	}

	//Calculate the address from the public key
	publicAddress := crypto.PubkeyToAddress(*signaturePublicKey)
	return publicAddress, [20]byte{}
}

func (txInfo *CreateCollectionTxInfo) GetNonce() int64 {
	return txInfo.Nonce
}

func (txInfo *CreateCollectionTxInfo) GetExpiredAt() int64 {
	return txInfo.ExpiredAt
}

func (txInfo *CreateCollectionTxInfo) Hash(hFunc hash.Hash) (msgHash []byte, err error) {
	packedFee, err := ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		log.Println("[ComputeTransferMsgHash] unable to packed amount", err.Error())
		return nil, err
	}
	msgHash = Poseidon(ChainId, TxTypeCreateCollection, txInfo.AccountIndex, txInfo.Nonce, txInfo.ExpiredAt,
		txInfo.GasFeeAssetId, packedFee)
	return msgHash, nil
}

func (txInfo *CreateCollectionTxInfo) GetGas() (int64, int64, *big.Int) {
	return txInfo.GasAccountIndex, txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount
}
