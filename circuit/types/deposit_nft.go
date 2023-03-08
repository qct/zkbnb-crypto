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

package types

type DepositNftTx struct {
	AccountIndex        int64
	NftIndex            int64
	L1Address           string
	NftContentHash      []byte
	CreatorAccountIndex int64
	CreatorTreasuryRate int64
	CollectionId        int64
	NftContentType      int8
}

type DepositNftTxConstraints struct {
	AccountIndex        Variable
	L1Address           Variable
	NftIndex            Variable
	NftContentHash      [2]Variable
	CreatorAccountIndex Variable
	CreatorTreasuryRate Variable
	CollectionId        Variable
	NftContentType      Variable
}

func EmptyDepositNftTxWitness() (witness DepositNftTxConstraints) {
	return DepositNftTxConstraints{
		AccountIndex:        ZeroInt,
		L1Address:           ZeroInt,
		NftIndex:            ZeroInt,
		NftContentHash:      [2]Variable{ZeroInt, ZeroInt},
		CreatorAccountIndex: ZeroInt,
		CreatorTreasuryRate: ZeroInt,
		CollectionId:        ZeroInt,
		NftContentType:      ZeroInt,
	}
}

func SetDepositNftTxWitness(tx *DepositNftTx) (witness DepositNftTxConstraints) {
	witness = DepositNftTxConstraints{
		AccountIndex:        tx.AccountIndex,
		L1Address:           tx.L1Address,
		NftIndex:            tx.NftIndex,
		NftContentHash:      GetNftContentHashFromBytes(tx.NftContentHash),
		CreatorAccountIndex: tx.CreatorAccountIndex,
		CreatorTreasuryRate: tx.CreatorTreasuryRate,
		CollectionId:        tx.CollectionId,
		NftContentType:      tx.NftContentType,
	}
	return witness
}

func VerifyDepositNftTx(
	api API,
	flag Variable,
	tx DepositNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
	pubData = CollectPubDataFromDepositNft(api, tx)

	// verify params
	// check empty nft
	CheckEmptyNftNode(api, flag, nftBefore)
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.L1Address, accountsBefore[0].L1Address)
	//NftContentType
	IsVariableEqual(api, flag, tx.NftContentType, nftBefore.NftContentType)
	return pubData
}
