# Soft Spoon pre-DAO-fork 改造 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 改造 core-geth，新增 `--softspoon` 网络，使其在区块 1428756（theDAO 合约部署前）后从 1428757 起自出块、永久 PoW、CPU/单卡可挖，并能分发给其他节点快速同步。

**Architecture:** 新增 `PreDAOForkChainConfig`（`coregeth.CoreGethChainConfig`），复用主网 Frontier alloc 保证 genesis 哈希与主网一致；难度炸弹通过 `DisposalBlock=1428757`(ECIP1041) 在分叉后移除；分叉首块一次性难度重置无法用配置表达，在 `ethash.CalcDifficulty` 开头注入一段以 `ChainID==2517` gate 的代码；通过镜像 `ClassicFlag` 的接线新增 `--softspoon` flag。

**Tech Stack:** Go 1.21+，core-geth（multi-geth 衍生），go test，ethash PoW。

## Global Constraints

- 分叉点区块号：**1428757**（第一个自出块）；截断点：**1428756**。
- ChainID / NetworkID：**2517**。
- genesis 必须沿用主网 Frontier alloc，genesis 哈希必须等于 `MainnetGenesisHash`（`0xd4e56740…`）。
- 永久 PoW：不启用 EIP150/EIP155 之外的任何 EIP/ECIP，不设 TTD/Merge。
- 难度重置注入必须以 `ChainID==2517` gate，禁止影响 Classic/mainnet/Mordor 等现有网络。
- 不启用 ECIP1099(etchash)：使用标准 ethash DAG 目录，不要把 `--softspoon` 加入 etchash 目录分支。
- 所有改动遵循 core-geth 现有代码风格与 LGPL 版权头。
- 提交信息结尾加 `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`。

---

## File Structure

| 文件 | 责任 |
|---|---|
| `params/config_softspoon.go`（新建） | 定义 `PreDAOForkChainConfig` |
| `params/genesis_softspoon.go`（新建） | 定义 `DefaultPreDAOForkGenesisBlock()` |
| `params/config_softspoon_test.go`（新建） | 配置激活点 + genesis 哈希测试 |
| `consensus/ethash/consensus.go`（修改） | `CalcDifficulty` 开头注入分叉首块难度重置 + Soft Spoon 常量 |
| `consensus/ethash/consensus_softspoon_test.go`（新建） | 难度重置行为测试 |
| `cmd/utils/flags.go`（修改） | `SoftSpoonFlag` 定义 + 网络接线 |
| `cmd/geth/main.go`（修改） | 启动日志分支 |
| `docs/RUNBOOK-softspoon.md`（新建） | 截断 → 铸造 → 验证 → 镜像 → checkpoint 回填手册 |

---

## Task 1: PreDAOForkChainConfig 与 genesis

**Files:**
- Create: `params/config_softspoon.go`
- Create: `params/genesis_softspoon.go`
- Test: `params/config_softspoon_test.go`

**Interfaces:**
- Produces:
  - `params.PreDAOForkChainConfig *coregeth.CoreGethChainConfig`（`ChainID=2517`，`EIP2FBlock=EIP7FBlock=1150000`，`EIP155Block=DisposalBlock=1428757`，其余 EIP/ECIP 为 nil，`NetworkID=2517`，`Ethash` 非 nil）
  - `params.DefaultPreDAOForkGenesisBlock() *genesisT.Genesis`

- [ ] **Step 1: 写失败测试 `params/config_softspoon_test.go`**

```go
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
```

- [ ] **Step 2: 运行测试确认失败（未定义）**

Run: `cd /Users/a/core-geth && go test ./params/ -run 'TestPreDAOFork' -v`
Expected: 编译失败 `undefined: PreDAOForkChainConfig` / `undefined: DefaultPreDAOForkGenesisBlock`

- [ ] **Step 3: 创建 `params/config_softspoon.go`**

```go
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

	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

// PreDAOForkChainConfig is the chain config for the "Soft Spoon" art-project network:
// Ethereum forked at block 1428757 (the block right after the last pre-theDAO-contract
// block, 1428756). History 0..1428756 validates with original Frontier->Homestead rules;
// from 1428757 the chain stays PoW forever with EIP155 replay protection (chainID 2517)
// and no difficulty bomb. The one-time difficulty reset at 1428757 lives in
// consensus/ethash (gated by chainID 2517) since it is not expressible via config.
var PreDAOForkChainConfig = &coregeth.CoreGethChainConfig{
	NetworkID:                 2517,
	ChainID:                   big.NewInt(2517),
	Ethash:                    new(ctypes.EthashConfig),
	SupportedProtocolVersions: vars.DefaultProtocolVersions,

	// Pre-fork history rules — must match Ethereum mainnet history.
	EIP2FBlock: big.NewInt(1_150_000), // Homestead difficulty adjustment
	EIP7FBlock: big.NewInt(1_150_000), // Homestead DELEGATECALL

	// DAOForkBlock left nil: never execute the DAO fork.

	// Our fork rules, effective from the first self-mined block.
	EIP155Block:   big.NewInt(1_428_757), // replay protection with chainID 2517
	DisposalBlock: big.NewInt(1_428_757), // ECIP1041: remove difficulty bomb after the fork

	// TrustedCheckpoint left nil; backfilled in release phase B (see RUNBOOK).
}
```

- [ ] **Step 4: 创建 `params/genesis_softspoon.go`**

```go
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
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd /Users/a/core-geth && go test ./params/ -run 'TestPreDAOFork' -v`
Expected: `PASS`（`TestPreDAOForkConfig` 与 `TestPreDAOForkGenesisHash` 均通过）

- [ ] **Step 6: 提交**

```bash
cd /Users/a/core-geth
git add params/config_softspoon.go params/genesis_softspoon.go params/config_softspoon_test.go
git commit -m "params: 新增 Soft Spoon PreDAOForkChainConfig 与 genesis

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 2: 分叉首块难度重置注入

**Files:**
- Modify: `consensus/ethash/consensus.go`（`CalcDifficulty` 约 line 364；常量加到文件内 `var` 区）
- Test: `consensus/ethash/consensus_softspoon_test.go`

**Interfaces:**
- Consumes: `params.PreDAOForkChainConfig`、`params.ClassicChainConfig`（Task 1 / 既有）
- Produces:
  - 包级变量 `softSpoonChainID = big.NewInt(2517)`、`softSpoonForkBlock = big.NewInt(1428757)`、`softSpoonForkInitDifficulty = big.NewInt(0x20000000)`（可调）
  - `CalcDifficulty` 在 `next==1428757 && chainID==2517` 时返回 `softSpoonForkInitDifficulty` 的副本

- [ ] **Step 1: 写失败测试 `consensus/ethash/consensus_softspoon_test.go`**

```go
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
	parent := headerAt(1_428_757, big.NewInt(0x20000000))
	got := CalcDifficulty(params.PreDAOForkChainConfig, parent.Time+13, parent)
	if got.Cmp(softSpoonForkInitDifficulty) == 0 {
		t.Fatal("block 1428758 should not be force-reset")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/a/core-geth && go test ./consensus/ethash/ -run 'TestSoftSpoon' -v`
Expected: 编译失败 `undefined: softSpoonForkInitDifficulty`

- [ ] **Step 3: 在 `consensus/ethash/consensus.go` 的 `var (` 区（与 `big1` 同块，约 line 530）后新增常量**

在 line 534 的 `)` 之后新增：

```go
// Soft Spoon (pre-theDAO art fork) difficulty parameters. The one-time reset at the
// fork block cannot be expressed via chain config, so it is applied here and gated by
// ChainID so no other network (Classic=61, mainnet=1, Mordor, ...) is affected.
// Keep softSpoonForkBlock in sync with params.PreDAOForkChainConfig (EIP155Block/DisposalBlock).
var (
	softSpoonChainID            = big.NewInt(2517)
	softSpoonForkBlock          = big.NewInt(1_428_757)
	softSpoonForkInitDifficulty = big.NewInt(0x20000000) // ≈5.4e8; tune down for CPU minting (floor: vars.MinimumDifficulty)
)
```

- [ ] **Step 4: 在 `CalcDifficulty`（line 364）开头注入重置逻辑**

把：

```go
func CalcDifficulty(config ctypes.ChainConfigurator, time uint64, parent *types.Header) *big.Int {
	next := new(big.Int).Add(parent.Number, big1)
	out := new(big.Int)
```

改为：

```go
func CalcDifficulty(config ctypes.ChainConfigurator, time uint64, parent *types.Header) *big.Int {
	next := new(big.Int).Add(parent.Number, big1)
	out := new(big.Int)

	// Soft Spoon: one-time difficulty reset at the fork block, only for chainID 2517.
	if cid := config.GetChainID(); cid != nil && cid.Cmp(softSpoonChainID) == 0 &&
		next.Cmp(softSpoonForkBlock) == 0 {
		return new(big.Int).Set(softSpoonForkInitDifficulty)
	}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd /Users/a/core-geth && go test ./consensus/ethash/ -run 'TestSoftSpoon' -v`
Expected: 三个用例全部 `PASS`

- [ ] **Step 6: 跑既有难度测试确保无回归**

Run: `cd /Users/a/core-geth && go test ./consensus/ethash/ -run 'Difficulty' -v`
Expected: `PASS`（既有 Classic/mainnet 难度用例不受影响）

- [ ] **Step 7: 提交**

```bash
cd /Users/a/core-geth
git add consensus/ethash/consensus.go consensus/ethash/consensus_softspoon_test.go
git commit -m "consensus/ethash: 分叉首块难度重置（chainID 2517 gate）

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 3: --softspoon CLI flag 接线

**Files:**
- Modify: `cmd/utils/flags.go`（多处，镜像 `ClassicFlag`）
- Modify: `cmd/geth/main.go`（约 line 350，启动日志）

**Interfaces:**
- Consumes: `params.DefaultPreDAOForkGenesisBlock()`（Task 1）
- Produces: `utils.SoftSpoonFlag *cli.BoolFlag`（Name: `softspoon`）；`geth --softspoon` 选用 PreDAOFork genesis，NetworkId 自动取 2517

本任务无单元测试（CLI 接线），以 `go build` + `dumpgenesis` 烟测验证。

- [ ] **Step 1: 定义 flag（`cmd/utils/flags.go` 约 line 169，`ClassicFlag` 之后）**

```go
	SoftSpoonFlag = &cli.BoolFlag{
		Name:     "softspoon",
		Usage:    "Soft Spoon network: Ethereum forked just before the theDAO contract (block 1428757), permanent PoW",
		Category: flags.EthCategory,
	}
```

- [ ] **Step 2: 注册到网络 flag 列表（约 line 1132，`ClassicFlag,` 之后加一行）**

```go
		ClassicFlag,
		SoftSpoonFlag,
```

- [ ] **Step 3: bootnodes 留空（约 line 1208，`setBootstrapNodes` 的 switch 内新增分支）**

```go
		case ctx.Bool(ClassicFlag.Name):
			urls = params.ClassicBootnodes
		case ctx.Bool(SoftSpoonFlag.Name):
			urls = nil // private art-chain: no public bootnodes
```

- [ ] **Step 4: datadir 子目录（约 line 1688，`dataDirPathForCtxChainConfig` 内新增分支）**

```go
		case ctx.Bool(ClassicFlag.Name):
			return filepath.Join(baseDataDirPath, "classic")
		case ctx.Bool(SoftSpoonFlag.Name):
			return filepath.Join(baseDataDirPath, "softspoon")
```

- [ ] **Step 5: 互斥检查（约 line 1948，把 `SoftSpoonFlag` 加入 `CheckExclusive` 参数列表）**

```go
	CheckExclusive(ctx, MainnetFlag, DeveloperFlag, DeveloperPoWFlag, SepoliaFlag, ClassicFlag, MordorFlag, MintMeFlag, HoleskyFlag, SoftSpoonFlag)
```

- [ ] **Step 6: genesis 选择（约 line 2505，`genesisForCtxChainConfig` 内 `ClassicFlag` 分支后新增）**

```go
		case ctx.Bool(ClassicFlag.Name):
			genesis = params.DefaultClassicGenesisBlock()
		case ctx.Bool(SoftSpoonFlag.Name):
			genesis = params.DefaultPreDAOForkGenesisBlock()
```

- [ ] **Step 7: 启动日志（`cmd/geth/main.go` 约 line 350，`ClassicFlag` 分支后新增）**

```go
	case ctx.IsSet(utils.ClassicFlag.Name):
		log.Info("Starting Geth on Ethereum Classic...")

	case ctx.IsSet(utils.SoftSpoonFlag.Name):
		log.Info("Starting Geth on Soft Spoon (pre-theDAO fork)...")
```

- [ ] **Step 8: 编译**

Run: `cd /Users/a/core-geth && make geth`
Expected: 编译成功，生成 `build/bin/geth`

- [ ] **Step 9: 烟测 —— flag 出现在帮助里**

Run: `cd /Users/a/core-geth && ./build/bin/geth --help 2>&1 | grep -A1 softspoon`
Expected: 输出含 `--softspoon` 及其 Usage 描述

- [ ] **Step 10: 烟测 —— genesis 正确、chainId=2517**

Run: `cd /Users/a/core-geth && ./build/bin/geth --softspoon dumpgenesis 2>/dev/null | grep -o '"chainId":[0-9]*'`
Expected: `"chainId":2517`

- [ ] **Step 11: 提交**

```bash
cd /Users/a/core-geth
git add cmd/utils/flags.go cmd/geth/main.go
git commit -m "cmd: 新增 --softspoon 网络 flag

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 4: 运行手册（截断 → 铸造 → 验证 → 镜像 → checkpoint）

**Files:**
- Create: `docs/RUNBOOK-softspoon.md`

无测试；交付一份可照做的手册。完成后人工通读校验命令完整、无占位符。

- [ ] **Step 1: 创建 `docs/RUNBOOK-softspoon.md`**

````markdown
# Soft Spoon 运行手册

> 所有操作在 ETC archive 库的**副本**上进行，切勿改动已同步的原库。
> 假设已编译出含 `--softspoon` 的新二进制 `build/bin/geth`，以及一份**原版** core-geth 二进制（用于截断），记作 `geth-classic`。

## 0. 前提
- ETC archive node 已同步至 1428756 之后，数据在 `~/Library/Ethereum/classic/geth`。
- 1428756 的区块哈希可用原库查得（见步骤 2）。

## 1. 复制数据库
```bash
cp -a ~/Library/Ethereum/classic ~/Library/Ethereum/softspoon-work
```

## 2. 用原版 classic 二进制截断到 1428756
启动控制台：
```bash
geth-classic --classic --datadir ~/Library/Ethereum/softspoon-work \
  --syncmode full --gcmode archive --maxpeers 0 --nodiscover console
```
在控制台中：
```javascript
// 确认 1428756 的哈希（信任锚的父块）
eth.getBlock(1428756).hash
// 回滚链头到 1428756，删除其后的真实 ETC 区块
debug.setHead("0x15CF94")   // 0x15CF94 == 1428756
eth.blockNumber              // 应为 1428756
exit
```

## 3. 用 softspoon 二进制铸造分叉首块
> 如 CPU 出块过慢，先在 `consensus/ethash/consensus.go` 把 `softSpoonForkInitDifficulty`
> 调低（如 `0x100000`），`make geth` 重新编译。下限为 `vars.MinimumDifficulty`(131072)。
```bash
build/bin/geth --softspoon --datadir ~/Library/Ethereum/softspoon-work \
  --syncmode full --gcmode archive --maxpeers 0 --nodiscover \
  --mine --miner.threads 1 --miner.etherbase 0xYOUR_ADDRESS console
```
等待挖出 1428757，记录其哈希：
```javascript
eth.getBlock(1428757)
// 记录 .hash 与 .difficulty（应等于 softSpoonForkInitDifficulty）
```

## 4. 验证
```javascript
eth.getBlock(1428756).hash === <步骤2记录值>   // 历史对接正确
eth.getBlock(1428757).number                    // 1428757
eth.getBlock(1428757).difficulty                 // == softSpoonForkInitDifficulty
admin.nodeInfo.protocols.eth.network             // 2517
// 发一笔 EIP155（chainId 2517）交易，确认重放保护生效
```

## 5. 制作链镜像（供他人快速同步）
任选其一：
- **整库打包**（最快还原）：先停掉节点，再打包整个工作目录。
  ```bash
  # 确保 geth 已退出，避免打包到半写入状态的库
  tar -C ~/Library/Ethereum -czf softspoon-chain.tar.gz softspoon-work
  ```
  他人：解包到自己的 datadir 后直接 `--softspoon` 启动。
- **链文件导出/导入**（跨版本更稳）：
  ```bash
  build/bin/geth --softspoon --datadir ~/Library/Ethereum/softspoon-work export softspoon-0-1428757.rlp 0 1428757
  # 他人：
  build/bin/geth --softspoon --datadir <新库> import softspoon-0-1428757.rlp
  ```

## 6. 固化 TrustedCheckpoint（阶段 B，可选但推荐）
当链增长越过分叉点之后的一个 CHT section 边界（每 32768 块一段）后，用 core-geth
自带的 checkpoint 工具算出该段的 `{SectionIndex, SectionHead, CHTRoot, BloomRoot}`，
填入 `params/config_softspoon.go` 的 `PreDAOForkChainConfig.TrustedCheckpoint`，
然后 `make geth` 重新编译并发布。CHTRoot 在密码学上承诺了 0…SectionHead 间每个区块哈希
（含 1428757），因此该 checkpoint 即作为全网信任根钉死了分叉链。

## 发布清单
- [ ] 1428757 规范哈希：__________
- [ ] softSpoonForkInitDifficulty 实际取值：__________
- [ ] 链镜像文件：__________
- [ ] TrustedCheckpoint 已回填并重新编译：是 / 否
````

- [ ] **Step 2: 通读校验**（确认每条命令完整、无 `TODO`/占位符；`0x15CF94` 等十六进制与十进制对应正确）

- [ ] **Step 3: 提交**

```bash
cd /Users/a/core-geth
git add docs/RUNBOOK-softspoon.md
git commit -m "docs: Soft Spoon 运行手册（截断/铸造/镜像/checkpoint）

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## 验收（全部任务完成后）

- [ ] `go test ./params/ ./consensus/ethash/ -run 'PreDAOFork|SoftSpoon|Difficulty'` 全绿
- [ ] `make geth` 编译通过
- [ ] `./build/bin/geth --softspoon dumpgenesis` 输出 chainId 2517、genesis 哈希 `0xd4e5…`
- [ ] `docs/RUNBOOK-softspoon.md` 可照做，无占位符
- [ ] 现有网络（classic/mainnet/mordor）行为无回归
