<p align="right"><strong>English</strong> | <a href="#zh">中文</a></p>

<a name="en"></a>

# Soft Spoon — Node Deployment Guide

> Soft Spoon is an art project: a fork of Ethereum at the block right before the
> theDAO contract was deployed (fork block **1428757**), kept permanent PoW.
> This guide is for **operators who want to run / mine a node** on the existing
> chain. It does **not** cover one-time chain creation (truncation / minting) —
> that has already been done by the project; you only need to obtain the chain
> and run a node.

## 1. Network parameters

| Item | Value |
|------|-------|
| Network name (flag) | `--softspoon` |
| Chain ID | `2517` |
| Network ID | `2517` |
| Consensus | Ethash PoW (permanent, no Merge) |
| Genesis hash | `0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3` |
| Fork block (first self-mined) | `1428757` |
| Fork block `1428757` hash | `0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564` |
| Fork block `1428757` difficulty | `1048576` (`0x100000`) |
| Trusted checkpoint | `<TBD — to be filled in>` |

## 2. Build

Requires Go 1.21+ and a C toolchain (gcc/clang), git, make.

```bash
git clone https://github.com/CryptoZR/SoftSpoon.git
cd SoftSpoon
make geth
# binary at ./build/bin/geth
```

## 3. Obtain the chain

You need the chain data up to and beyond the fork block `1428757`. Two ways:

### Option A — Restore from the published chain image (recommended, fastest)

```bash
# Download the image (URL TBD)
curl -L -o softspoon-chain.tar.gz "<IMAGE_URL — to be filled in>"

# Extract into your Ethereum data root (creates the softspoon datadir)
tar -C ~/Library/Ethereum -xzf softspoon-chain.tar.gz
# Linux default root: ~/.ethereum
```

### Option B — Sync from the network

Sync from a project bootnode. Trust is anchored by the hardcoded
`TrustedCheckpoint` baked into the binary, so snap sync is safe.

```bash
./build/bin/geth --softspoon \
  --bootnodes "<BOOTNODE_ENODE — to be filled in>" \
  --syncmode snap \
  --datadir <your-datadir>
```

## 4. Run a node

```bash
./build/bin/geth --softspoon \
  --datadir <your-datadir> \
  --bootnodes "<BOOTNODE_ENODE — to be filled in>" \
  --http --http.api eth,net,web3
```

Verify you are on the right chain:

```bash
./build/bin/geth attach <your-datadir>/geth.ipc
> eth.chainId()                 // 2517
> eth.getBlock(1428757).hash    // 0xd4f997...6564
```

## 5. Mining

Soft Spoon stays PoW and is CPU/single-GPU mineable.

```bash
./build/bin/geth --softspoon \
  --datadir <your-datadir> \
  --bootnodes "<BOOTNODE_ENODE — to be filled in>" \
  --mine --miner.threads 1 \
  --miner.etherbase 0xYOUR_REWARD_ADDRESS
```

Difficulty after the fork follows the standard Homestead dynamic adjustment
(no difficulty bomb), so it tracks the real network hashrate automatically.

---

<a name="zh"></a>

<p align="right"><a href="#en">English</a> | <strong>中文</strong></p>

# Soft Spoon — 节点部署指南

> Soft Spoon 是一个艺术项目：把以太坊在 theDAO 合约部署前的区块处分叉（分叉首块
> **1428757**），并永久保持 PoW。本指南面向**希望在现有链上运行 / 挖矿的节点运营者**，
> **不**涉及一次性的建链（截断 / 铸造）——那部分已由项目方完成，你只需获取链数据并
> 运行节点即可。

## 1. 网络参数

| 项目 | 取值 |
|------|------|
| 网络名（flag） | `--softspoon` |
| Chain ID | `2517` |
| Network ID | `2517` |
| 共识 | Ethash PoW（永久，无 Merge） |
| Genesis 哈希 | `0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3` |
| 分叉首块（首个自出块） | `1428757` |
| 分叉首块 `1428757` 哈希 | `0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564` |
| 分叉首块 `1428757` 难度 | `1048576`（`0x100000`） |
| 可信检查点 | `<待填充>` |

## 2. 编译

需要 Go 1.21+、C 工具链（gcc/clang）、git、make。

```bash
git clone https://github.com/CryptoZR/SoftSpoon.git
cd SoftSpoon
make geth
# 二进制位于 ./build/bin/geth
```

## 3. 获取链数据

你需要拿到包含分叉首块 `1428757` 及之后的链数据，两种方式：

### 方式 A — 从发布的链镜像还原（推荐，最快）

```bash
# 下载镜像（地址待填充）
curl -L -o softspoon-chain.tar.gz "<镜像地址 — 待填充>"

# 解包到你的 Ethereum 数据根目录（会生成 softspoon 数据目录）
tar -C ~/Library/Ethereum -xzf softspoon-chain.tar.gz
# Linux 默认根目录：~/.ethereum
```

### 方式 B — 从网络同步

通过项目 bootnode 同步。信任由编译进二进制的硬编码 `TrustedCheckpoint` 锚定，
因此 snap 同步是安全的。

```bash
./build/bin/geth --softspoon \
  --bootnodes "<bootnode enode — 待填充>" \
  --syncmode snap \
  --datadir <你的数据目录>
```

## 4. 运行节点

```bash
./build/bin/geth --softspoon \
  --datadir <你的数据目录> \
  --bootnodes "<bootnode enode — 待填充>" \
  --http --http.api eth,net,web3
```

验证你在正确的链上：

```bash
./build/bin/geth attach <你的数据目录>/geth.ipc
> eth.chainId()                 // 2517
> eth.getBlock(1428757).hash    // 0xd4f997...6564
```

## 5. 挖矿

Soft Spoon 保持 PoW，CPU / 单卡即可挖。

```bash
./build/bin/geth --softspoon \
  --datadir <你的数据目录> \
  --bootnodes "<bootnode enode — 待填充>" \
  --mine --miner.threads 1 \
  --miner.etherbase 0x你的收款地址
```

分叉之后的难度采用标准 Homestead 动态调整（无难度炸弹），会自动跟随网络真实算力。
