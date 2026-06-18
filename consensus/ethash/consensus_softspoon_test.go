package ethash

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func headerAt(n int64, diff *big.Int) *types.Header {
	return &types.Header{Number: big.NewInt(n), Time: 1000, Difficulty: diff}
}

// 分叉首块（parent=1428756）在 chainID 2517 下应重置为单卡初始难度。
func TestSoftSpoonForkResetsDifficulty(t *testing.T) {
	parent := headerAt(1_428_756, big.NewInt(0x7fffffffffff)) // 父块天量历史难度
	got := CalcDifficulty(params.PreDAOForkChainConfig, parent.Time+13, parent)
	if got.Cmp(softSpoonForkInitDifficulty) != 0 {
		t.Fatalf("fork block difficulty: want %v, got %v", softSpoonForkInitDifficulty, got)
	}
}

// 同样的高度，但在 Classic（chainID 61）下绝不能触发重置。
func TestSoftSpoonGateDoesNotAffectClassic(t *testing.T) {
	parent := headerAt(1_428_756, big.NewInt(0x7fffffffffff))
	got := CalcDifficulty(params.ClassicChainConfig, parent.Time+13, parent)
	if got.Cmp(softSpoonForkInitDifficulty) == 0 {
		t.Fatal("classic must not be reset at 1428757")
	}
}

// 分叉之后的块（parent=1428757）不再重置，走正常调整逻辑。
func TestSoftSpoonNoResetAfterFork(t *testing.T) {
	parent := headerAt(1_428_757, big.NewInt(0x25000000))
	got := CalcDifficulty(params.PreDAOForkChainConfig, parent.Time+13, parent)
	if got.Cmp(softSpoonForkInitDifficulty) == 0 {
		t.Fatal("block 1428758 should not be force-reset")
	}
}
