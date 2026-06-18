# Soft Spoon —— core-geth 在 theDAO 合约部署前分叉 · 设计文档

- 日期：2026-06-18
- 项目：Soft Spoon（基于以太坊公链的艺术项目）
- 目标：改造 core-geth，使其在区块 **1428756**（theDAO 合约部署前最后一块）后从 **1428757** 起自出块、永久 PoW、CPU/单卡可挖，并可分发给其他节点快速同步。

---

## 1. 链语义与总体架构

在区块 **1428756**（theDAO 合约部署前的最后一块）截断以太坊历史，从 **1428757** 起由作品自己出块。

- **0 ～ 1428756**：完整保留真实主网/ETC 历史（两条链在此高度前完全相同），用原始 Frontier→Homestead 规则验证。genesis 必须沿用主网 Frontier alloc，保证 genesis 哈希 = `0xd4e5…`（与主网一致），截断后的历史才能对接。
- **1428757（分叉首块）**：难度一次性重置为单卡/CPU 可挖值；启用 EIP155（chainid=2517）做重放保护；移除难度炸弹。
- **1428757 之后**：标准 Homestead 动态难度调整、无炸弹、无上限、永远 PoW。不进入 DAO/EIP150/Byzantium 等任何后续分叉，无 TTD、无 Merge。

### 数据引导（截断法）

在 archive 库**副本**上操作，顺序至关重要：

1. 先用**原版 `--classic` 二进制**执行 `debug.setHead(1428756)`，删除其后的真实 ETC 区块。
2. 再换成本次改造的 **`--softspoon` 二进制**启动挖矿。

> 顺序原因：若直接用新配置二进制启库，core-geth 的配置兼容性检查会在更高的分叉点（如 EIP150 的 2500000）触发强制回滚，绕路。先用原配置回滚到 1428756，再切换配置即可避免（在 1428756 高度，新旧配置对 ≤head 的区块规则无差异：Homestead 同为 1150000，DAO 同为 nil，EIP155 两者均 > head）。

截断后，作品挖出的 1428757 成为该库唯一的规范延续（无更高 TD 的竞争链）。

---

## 2. `PreDAOForkChainConfig`（core-geth 结构）

> 注意：claude.md 草案使用的是 go-ethereum 的 `ChainConfig` 结构；core-geth 实际用 `coregeth.CoreGethChainConfig`（细粒度 EIP 开关），需翻译如下。

新建 `params/config_softspoon.go`：

```go
var PreDAOForkChainConfig = &coregeth.CoreGethChainConfig{
    NetworkID:                 2517,
    ChainID:                   big.NewInt(2517),   // 防重放，并用于 gate 难度重置
    Ethash:                    new(ctypes.EthashConfig),
    SupportedProtocolVersions: vars.DefaultProtocolVersions,

    // —— 分叉点之前的历史规则，必须与主网一致 ——
    EIP2FBlock: big.NewInt(1150000),  // Homestead 难度调整
    EIP7FBlock: big.NewInt(1150000),  // Homestead DELEGATECALL

    // DAOForkBlock: nil               // 关键：不执行 DAO fork（DAO 实际在 1920000，> 分叉点）
    // EIP150 / EIP155(主网原值) / Byzantium… 全部留 nil

    // —— 作品自己的分叉规则 ——
    EIP155Block:   big.NewInt(1428757), // 从分叉首块起用 chainid=2517 做重放保护
    DisposalBlock: big.NewInt(1428757), // ECIP1041：分叉后移除难度炸弹

    // TrustedCheckpoint: nil           // 阶段 A 留 nil，阶段 B 回填（见 §5）
}
```

要点：

- **EIP155 单独启用、不带 EIP158/EIP161**：core-geth 中 `EIP155Block` 与 `EIP160FBlock`/`EIP161FBlock` 是独立开关，只开 EIP155（重放保护，不做状态清理）。历史块（<1428757）无 chainid 签名照常验证。
- **`DisposalBlock=1428757`** 同时承担"分叉后去炸弹"：历史块仍走原炸弹逻辑，能过验证；分叉后 `IsEnabled(ECIP1041)` 为 true，在加炸弹项前提前 return。
- 其余所有 EIP/ECIP 字段全部 `nil`，规则集永久停在 Homestead+EIP155，永远 PoW。

---

## 3. 难度一次性重置（代码注入）

唯一无法用配置表达的是"分叉首块丢弃父块天量难度、重置为 CPU/单卡值"。在 `consensus/ethash/consensus.go` 的 `CalcDifficulty`（约 line 364）开头注入，**用 ChainID==2517 gate**，避免误伤真实 Classic/mainnet 在 1428757 的历史块验证：

```go
func CalcDifficulty(config ctypes.ChainConfigurator, time uint64, parent *types.Header) *big.Int {
    next := new(big.Int).Add(parent.Number, big1)

    // Soft Spoon: 分叉首块一次性重置难度（仅本网络 chainid=2517）
    if cid := config.GetChainID(); cid != nil && cid.Cmp(softSpoonChainID) == 0 &&
        next.Cmp(softSpoonForkBlock) == 0 {
        return new(big.Int).Set(softSpoonForkInitDifficulty)
    }
    // ... 原有逻辑保持不动
}
```

包级常量（就近放在 `consensus.go` 或 `config_softspoon.go`）：

```go
var (
    softSpoonChainID            = big.NewInt(2517)
    softSpoonForkBlock          = big.NewInt(1428757)
    softSpoonForkInitDifficulty = big.NewInt(0x20000000) // 可调，见 §5；下限 vars.MinimumDifficulty(131072)
)
```

设计理由：

- 不新增 config 字段——给 `ChainConfigurator` 接口加字段需在 classic/multigeth/coregeth 等所有实现里铺一遍，过重。ChainID gate 是最小、隔离、不影响任何现有网络的方案（2517 为 Soft Spoon 独有）。
- 分叉之后（next > 1428757）自动走 `else if IsEnabled(EIP2)` 的 Homestead 调整分支；因 `DisposalBlock=1428757` 已启用 ECIP1041，在加炸弹前 return，天然无炸弹、有下限保护（`vars.MinimumDifficulty`）、无上限。**无需另写 `calcDifficultyForkHomestead`**——core-geth 现有逻辑已覆盖，比 claude.md 草案更省。

---

## 4. `--softspoon` flag 接入与 genesis

### genesis（`params/genesis_softspoon.go`）

复用主网 Frontier alloc，保证 genesis 哈希 `0xd4e5…` 与主网一致：

```go
func DefaultPreDAOForkGenesisBlock() *genesisT.Genesis {
    return &genesisT.Genesis{
        Config:     params.PreDAOForkChainConfig,
        Nonce:      66,
        ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
        GasLimit:   5000,
        Difficulty: big.NewInt(17179869184),
        Alloc:      genesisT.DecodePreAlloc(mainnetAllocData), // 同 DefaultClassicGenesisBlock
    }
}
```

（Nonce/ExtraData/GasLimit/Difficulty 全部照搬 Frontier 创世，与 `DefaultClassicGenesisBlock` 一致，仅 `Config` 换成 `PreDAOForkChainConfig`。）

### CLI flag（`cmd/utils/flags.go` + `cmd/geth/main.go`）

镜像 `ClassicFlag` 的约 10 处接线：

| 位置（约） | 改动 |
|---|---|
| flags.go ~165 | 定义 `SoftSpoonFlag = &cli.BoolFlag{Name:"softspoon", ...}` |
| flags.go ~1126 | 加入网络 flag 组 |
| flags.go ~1208 / ~1245 | bootnodes / NetworkId 分支 |
| flags.go ~1688 / ~1788 / ~1812 / ~1858 | genesis、ethash cache/dataset 目录分支（PoW 挖矿需 DAG 目录） |
| flags.go ~1948 | `CheckExclusive` 加入互斥组 |
| flags.go ~2505 | `genesis = params.DefaultPreDAOForkGenesisBlock()` |
| main.go ~350 | 启动分支 |

bootnodes 留空（作品链/私链，无需公网种子节点）。

---

## 5. CPU 出块、链镜像分发、checkpoint 固化（两阶段发布）

**本质时序问题**：1428757 的 blockhash 是**非确定性的**（取决于矿工 nonce、timestamp、etherbase），只有挖出来那一刻才确定。整个作品因此是**两阶段发布**。

### 阶段 A · 铸造（一次性）

- `softSpoonForkInitDifficulty` 设成 CPU 可在可接受时间内出块的值。`0x20000000`(≈5.4e8) 对 CPU 单线程偏高（可能数十分钟）；下限 `vars.MinimumDifficulty`(131072)。作为可调常量，按实测机器下调（如 `0x100000`，CPU 秒级出块）。出块后动态调整自动跟上真实算力，不影响艺术设定。
- `geth --softspoon --mine --miner.threads N`（CPU）挖出 **1428757**，记录其**规范 blockhash**——全网唯一信任锚。

### 阶段 B · 固化与分发

1. **链镜像（快速同步）**：挖出首块后，把截断+出块后的 datadir 打包（`tar` 整库，或 `geth export` 成链文件）作为交付物。其他节点 `geth import` / 解包还原即可秒级到达规范链头，无需各自重做截断。
2. **TrustedCheckpoint 硬编码**：当链越过分叉点之后的一个 CHT section 边界（每 32768 块一段）时，用 core-geth 自带 checkpoint 工具算出该段 `{SectionIndex, SectionHead, CHTRoot, BloomRoot}`，硬编码进 `PreDAOForkChainConfig.TrustedCheckpoint`。

> 关于"把 1428757 的 blockhash 硬编码到 checkpoint"：core-geth 的 checkpoint 是 **CHT 段承诺**，非单块哈希字段。但 **CHTRoot 在密码学上已承诺 0…SectionHead 间每一个区块哈希，包含 1428757**。因此硬编码该 TrustedCheckpoint = 间接但密码学强绑定地钉死 1428757，达到所需效果，且为 core-geth 原生机制（无需自造校验）。新节点 snap/light sync 以此为信任根。

`TrustedCheckpoint` 字段初始留 `nil`（阶段 A），阶段 B 回填——写进发布清单。

---

## 6. 运行手册（交付物，`docs/` 下）

1. **复制库**：复制 archive 数据库副本（保护已同步原库），仅在副本上操作。
2. **截断**：原版 `--classic` 二进制启动，控制台 `debug.setHead("0x…1428756 的高度十六进制")`，确认 head=1428756。
3. **编译**：`make geth`，得到含 `--softspoon` 的新二进制。
4. **铸造首块**：`geth --softspoon --datadir <副本> --mine --miner.threads N --miner.etherbase <地址>`，挖出 1428757。
5. **验证**：检查 1428757 出块、`difficulty == softSpoonForkInitDifficulty`、用 chainid 2517 发交易（EIP155 重放保护生效）。
6. **制作镜像**：打包 datadir 或 `geth export` 链文件，提供他人 `geth import`/还原。
7. **回填 checkpoint**：到达 section 边界后算出 CHT checkpoint，硬编码进 `PreDAOForkChainConfig.TrustedCheckpoint`，重新编译并发布。

---

## 7. 交付边界

- ✅ core-geth 源码改造（config / 难度重置 / `--softspoon` flag / genesis）并编译通过。
- ✅ 运行手册（截断 → 铸造 → 验证 → 镜像 → checkpoint 回填）。
- 截断、挖矿、镜像制作由作者按手册在本机执行（涉及数十 GB archive 库）。

---

## 8. 文件改动清单

| 文件 | 改动 |
|---|---|
| `params/config_softspoon.go`（新） | `PreDAOForkChainConfig` + Soft Spoon 常量 |
| `params/genesis_softspoon.go`（新） | `DefaultPreDAOForkGenesisBlock()`（复用 `mainnetAllocData`） |
| `consensus/ethash/consensus.go` | `CalcDifficulty` 开头注入分叉首块难度重置（chainid gate） |
| `cmd/utils/flags.go` | `SoftSpoonFlag` 定义 + 约 10 处网络接线 |
| `cmd/geth/main.go` | 启动分支 |
| `docs/…/RUNBOOK.md`（新） | 运行手册 |
