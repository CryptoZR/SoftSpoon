package params

import (
	"math/big"
	"testing"
)

func TestPreDAOForkConfig(t *testing.T) {
	c := PreDAOForkChainConfig

	if c.GetChainID().Cmp(big.NewInt(2517)) != 0 {
		t.Fatalf("chainID: want 2517, got %v", c.GetChainID())
	}

	// Homestead (EIP2) 在 1150000 之前未启用、之后启用
	if c.IsEnabled(c.GetEIP2Transition, big.NewInt(1_149_999)) {
		t.Fatal("EIP2 should be disabled before 1150000")
	}
	if !c.IsEnabled(c.GetEIP2Transition, big.NewInt(1_150_000)) {
		t.Fatal("EIP2 should be enabled at 1150000")
	}

	// DAO fork (EIP779) 永不启用
	for _, bn := range []*big.Int{big.NewInt(0), big.NewInt(1_428_757), big.NewInt(1_920_000), big.NewInt(10_000_000)} {
		if c.IsEnabled(c.GetEthashEIP779Transition, bn) {
			t.Fatalf("DAO fork must never be enabled, got enabled at %v", bn)
		}
	}

	// EIP155 从 1428757 起启用，之前不启用
	if c.IsEnabled(c.GetEIP155Transition, big.NewInt(1_428_756)) {
		t.Fatal("EIP155 should be disabled at 1428756")
	}
	if !c.IsEnabled(c.GetEIP155Transition, big.NewInt(1_428_757)) {
		t.Fatal("EIP155 should be enabled at 1428757")
	}

	// ECIP1041 (DisposalBlock，去炸弹) 从 1428757 起启用，之前不启用
	if c.IsEnabled(c.GetEthashECIP1041Transition, big.NewInt(1_428_756)) {
		t.Fatal("ECIP1041 should be disabled at 1428756")
	}
	if !c.IsEnabled(c.GetEthashECIP1041Transition, big.NewInt(1_428_757)) {
		t.Fatal("ECIP1041 should be enabled at 1428757")
	}

	// 后续硬分叉永不启用（抽查 Byzantium 的 EIP100）
	if c.IsEnabled(c.GetEthashEIP100BTransition, big.NewInt(20_000_000)) {
		t.Fatal("EIP100 (Byzantium) must never be enabled")
	}
}

func TestPreDAOForkGenesisHash(t *testing.T) {
	genesis := DefaultPreDAOForkGenesisBlock()
	block := genesisToBlock(genesis, nil)
	if block.Hash() != MainnetGenesisHash {
		t.Errorf("genesis hash want %s, got %s", MainnetGenesisHash.Hex(), block.Hash().Hex())
	}
}
