// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
// Enhanced blockchain implementation by Circle Layer <https://circlelayer.com>

package tests

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	mainnetChainConfig = params.ChainConfig{
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(1150000),
		DAOForkBlock:   big.NewInt(1920000),
		DAOForkSupport: true,
		EIP150Block:    big.NewInt(2463000),
		EIP150Hash:     common.HexToHash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0"),
		EIP155Block:    big.NewInt(2675000),
		EIP158Block:    big.NewInt(2675000),
		ByzantiumBlock: big.NewInt(4370000),
	}
)

func TestDifficulty(t *testing.T) {
	t.Parallel()

	dt := new(testMatcher)
	// Not difficulty-tests
	dt.skipLoad("hexencodetest.*")
	dt.skipLoad("crypto.*")
	dt.skipLoad("blockgenesistest\\.json")
	dt.skipLoad("genesishashestest\\.json")
	dt.skipLoad("keyaddrtest\\.json")
	dt.skipLoad("txtest\\.json")

	// files are 2 years old, contains strange values
	dt.skipLoad("difficultyCustomHomestead\\.json")
	dt.skipLoad("difficultyMorden\\.json")
	dt.skipLoad("difficultyOlimpic\\.json")

	// dt.config("Ropsten", *params.RopstenChainConfig)
	// dt.config("Morden", *params.RopstenChainConfig)
	dt.config("Frontier", params.ChainConfig{})

	dt.config("Homestead", params.ChainConfig{
		HomesteadBlock: big.NewInt(0),
	})

	dt.config("Byzantium", params.ChainConfig{
		ByzantiumBlock: big.NewInt(0),
	})

	// dt.config("Frontier", *params.RopstenChainConfig)
	dt.config("MainNetwork", mainnetChainConfig)
	dt.config("CustomMainNetwork", mainnetChainConfig)
	dt.config("Constantinople", params.ChainConfig{
		ConstantinopleBlock: big.NewInt(0),
	})
	dt.config("EIP2384", params.ChainConfig{
		MuirGlacierBlock: big.NewInt(0),
	})
	dt.config("EIP4345", params.ChainConfig{
		ArrowGlacierBlock: big.NewInt(0),
	})
	dt.config("difficulty.json", mainnetChainConfig)

	dt.walk(t, difficultyTestDir, func(t *testing.T, name string, test *DifficultyTest) {
		cfg := dt.findConfig(t)
		if test.ParentDifficulty.Cmp(params.MinimumDifficulty) < 0 {
			t.Skip("difficulty below minimum")
			return
		}
		if err := dt.checkFailure(t, test.Run(cfg)); err != nil {
			t.Error(err)
		}
	})
}
