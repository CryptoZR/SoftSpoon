// Copyright 2024 The multi-geth Authors
// This file is part of the multi-geth library.
//
// The multi-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The multi-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the multi-geth library. If not, see <http://www.gnu.org/licenses/>.

package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

// PreDAOForkChainConfig is the chain config for the "Soft Spoon" art-project network:
// Ethereum's Soft Spoon happens at block 1428757 (the block right after the last pre-theDAO-contract
// block, 1428756). History 0..1428756 validates with original Frontier->Homestead rules;
// from 1428757 the chain stays PoW forever with EIP155 replay protection (chainID 7198)
// and no difficulty bomb. The one-time difficulty reset at 1428757 lives in
// consensus/ethash (gated by chainID 7198) since it is not expressible via config.
var PreDAOForkChainConfig = &coregeth.CoreGethChainConfig{
	NetworkID:                 7198,
	ChainID:                   big.NewInt(7198),
	Ethash:                    new(ctypes.EthashConfig),
	SupportedProtocolVersions: vars.DefaultProtocolVersions,

	// Pre-Soft-Spoon history rules — must match Ethereum mainnet history.
	EIP2FBlock: big.NewInt(1_150_000), // Homestead difficulty adjustment
	EIP7FBlock: big.NewInt(1_150_000), // Homestead DELEGATECALL

	// DAOForkBlock left nil: never execute the DAO fork.

	// Our Soft Spoon rules, effective from the first self-mined block.
	EIP155Block:   big.NewInt(1_428_757), // replay protection with chainID 7198
	DisposalBlock: big.NewInt(1_428_757), // ECIP1041: remove difficulty bomb after the Soft Spoon

	// TrustedCheckpoint pins the canonical Soft Spoon chain for snap-syncing full nodes.
	// Section 43 covers blocks [1409024, 1441791]; SectionHead is the hash of block 1441791.
	// CHTRoot/BloomRoot are unused by the full-node sync challenge (light/les removed) and left zero.
	TrustedCheckpoint: &ctypes.TrustedCheckpoint{
		SectionIndex: 43,
		SectionHead:  common.HexToHash("0xade01e713d874b87dc6de44db12fda26963b38ca9b83cc4dc764fb7c8548d762"),
	},
}
