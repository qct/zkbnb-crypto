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

type FullExitNftTx struct {
	AccountIndex        int64
	L1Address           string
	CreatorAccountIndex int64
	CreatorL1Address    string
	CreatorTreasuryRate int64
	NftIndex            int64
	CollectionId        int64
	NftContentHash      []byte
}

type FullExitNftTxConstraints struct {
	AccountIndex        Variable
	L1Address           Variable
	CreatorAccountIndex Variable
	CreatorL1Address    Variable
	CreatorTreasuryRate Variable
	NftIndex            Variable
	CollectionId        Variable
	NftContentHash      [2]Variable
}

func EmptyFullExitNftTxWitness() (witness FullExitNftTxConstraints) {
	return FullExitNftTxConstraints{
		AccountIndex:        ZeroInt,
		L1Address:           ZeroInt,
		CreatorAccountIndex: ZeroInt,
		CreatorL1Address:    ZeroInt,
		CreatorTreasuryRate: ZeroInt,
		NftIndex:            ZeroInt,
		CollectionId:        ZeroInt,
		NftContentHash:      [2]Variable{ZeroInt, ZeroInt},
	}
}

func SetFullExitNftTxWitness(tx *FullExitNftTx) (witness FullExitNftTxConstraints) {
	witness = FullExitNftTxConstraints{
		AccountIndex:        tx.AccountIndex,
		L1Address:           tx.L1Address,
		CreatorAccountIndex: tx.CreatorAccountIndex,
		CreatorL1Address:    tx.CreatorL1Address,
		CreatorTreasuryRate: tx.CreatorTreasuryRate,
		NftIndex:            tx.NftIndex,
		CollectionId:        tx.CollectionId,
		NftContentHash:      GetNftContentHashFromBytes(tx.NftContentHash),
	}
	return witness
}

func VerifyFullExitNftTx(
	api API, flag Variable,
	tx FullExitNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
	fromAccount := 0
	creatorAccount := 1

	pubData = CollectPubDataFromFullExitNft(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.L1Address, accountsBefore[fromAccount].L1Address)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[creatorAccount].AccountIndex)
	IsVariableEqual(api, flag, tx.CreatorL1Address, accountsBefore[creatorAccount].L1Address)
	isOwner := api.And(api.IsZero(api.Sub(tx.AccountIndex, nftBefore.OwnerAccountIndex)), flag)
	IsVariableEqual(api, isOwner, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
	IsVariableEqual(api, isOwner, tx.CreatorTreasuryRate, nftBefore.CreatorTreasuryRate)
	IsVariableEqual(api, isOwner, tx.NftContentHash[0], nftBefore.NftContentHash[0])
	IsVariableEqual(api, isOwner, tx.NftContentHash[1], nftBefore.NftContentHash[1])
	tx.NftContentHash[0] = api.Select(isOwner, tx.NftContentHash[0], 0)
	tx.NftContentHash[1] = api.Select(isOwner, tx.NftContentHash[1], 0)
	return pubData
}
