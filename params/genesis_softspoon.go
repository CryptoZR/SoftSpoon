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

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
)

// DefaultPreDAOForkGenesisBlock returns the Soft Spoon genesis block. It reuses the
// Ethereum mainnet Frontier allocation so the genesis hash equals MainnetGenesisHash,
// allowing the real history 0..1428756 to attach to this config.
func DefaultPreDAOForkGenesisBlock() *genesisT.Genesis {
	return &genesisT.Genesis{
		Config:     PreDAOForkChainConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		GasLimit:   5000,
		Difficulty: big.NewInt(17179869184),
		Alloc:      genesisT.DecodePreAlloc(mainnetAllocData),
	}
}
