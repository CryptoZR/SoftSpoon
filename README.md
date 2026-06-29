<p align="right"><strong>English</strong> | <a href="#zh">中文</a></p>

<a name="en"></a>

# Soft Spoon — Node Deployment Guide

> Soft Spoon is an art project: the Soft Spoon of Ethereum at the block right before the
> theDAO contract was deployed (Soft Spoon block **1428757**).
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
| Consensus | Ethash PoW |
| Genesis hash | `0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3` |
| Soft Spoon block (first self-mined) | `1428757` |
| Soft Spoon block `1428757` hash | `0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564` |
| Soft Spoon block `1428757` difficulty | `1048576` (`0x100000`) |
| Trusted checkpoint | section `43`, head `0xade01e713d874b87dc6de44db12fda26963b38ca9b83cc4dc764fb7c8548d762` (block `1441791`) |
| Default data directory | macOS `~/Library/Ethereum/softspoon`, Linux `~/.ethereum/softspoon` |
| Bootnodes | `enode://ef794a99…bb4564@95.217.104.247:30304`<br>`enode://c4d03b5f…c0b5f@188.40.138.215:30304` |

## 2. Build

Requires Go 1.21+ and a C toolchain (gcc/clang), git, make.

```bash
git clone https://github.com/CryptoZR/SoftSpoon.git
cd SoftSpoon
make geth
# binary at ./build/bin/geth
```

## 3. Obtain the chain

You need the chain data up to and beyond the Soft Spoon block `1428757`. Two ways:

### Option A — Restore from the published chain image (recommended, fastest)

Download `softspoon.tar.gz`:

- Google Drive: https://drive.google.com/file/d/1CSgA-Qf_QUSfUmDCSJaXLtMvzDslojay/view?usp=sharing
- Baidu Netdisk: https://pan.baidu.com/s/1x9wlD09ymku5w-Abs6u-bA?pwd=2517 (extraction code `2517`)

The archive contains a `softspoon/` directory (`geth/`), which is exactly the
default datadir that `geth --softspoon` uses. Extract it into the data root for
your OS — afterwards you do **not** need a `--datadir` flag.

```bash
# macOS — data root ~/Library/Ethereum
tar -xzf softspoon.tar.gz -C ~/Library/Ethereum

# Linux — data root ~/.ethereum
tar -xzf softspoon.tar.gz -C ~/.ethereum
```

Result: `~/Library/Ethereum/softspoon/geth` (macOS) or
`~/.ethereum/softspoon/geth` (Linux).

### Option B — Sync from the network

Sync from a project bootnode. Trust is anchored by the hardcoded
`TrustedCheckpoint` baked into the binary, so snap sync is safe.

```bash
./build/bin/geth --softspoon \
  --bootnodes "enode://ef794a991152c3bb9f6f659a631b7b244898a196daf63304c8a863168e6be80d480a4cb0337575e754fd461b26b191b0b444dd094be212085bfd296235bb4564@95.217.104.247:30304,enode://c4d03b5f58fea266160bec11507f8b03e71523ea5dcb8b769b8a93b6010894f9821082cf4b7d638d94d6a50718673e1c06bc8c022ba682d9020858b2bc0c0b5f@188.40.138.215:30304,enode://54ca4dd12a7ebefcc34c1e35ee5acd213a004130b9239f38b3fcbb2ca2493d5b88616b53fb22f2afe0989474ca1f5c8af939a5b1d10857e47221fc7649e4ec3e@95.217.201.92:30303" \
  --syncmode snap
```

## 4. Run a node

[![asciicast](https://asciinema.org/a/8up4tdJ0VWMOMzuJ.svg)](https://asciinema.org/a/8up4tdJ0VWMOMzuJ)

```bash
./build/bin/geth --softspoon \
  --http --http.api eth,net,web3 --bootnodes "enode://ef794a991152c3bb9f6f659a631b7b244898a196daf63304c8a863168e6be80d480a4cb0337575e754fd461b26b191b0b444dd094be212085bfd296235bb4564@95.217.104.247:30304,enode://c4d03b5f58fea266160bec11507f8b03e71523ea5dcb8b769b8a93b6010894f9821082cf4b7d638d94d6a50718673e1c06bc8c022ba682d9020858b2bc0c0b5f@188.40.138.215:30304,enode://54ca4dd12a7ebefcc34c1e35ee5acd213a004130b9239f38b3fcbb2ca2493d5b88616b53fb22f2afe0989474ca1f5c8af939a5b1d10857e47221fc7649e4ec3e@95.217.201.92:30303"
```

Verify you are on the right chain (default IPC path shown):

```bash
# macOS
./build/bin/geth attach ~/Library/Ethereum/softspoon/geth.ipc
# Linux: ~/.ethereum/softspoon/geth.ipc
> eth.chainId()                 // 2517
> eth.getBlock(1428757).hash    // 0xd4f997...6564
```

## 5. Mining

Soft Spoon is CPU/single-GPU mineable.

```bash
./build/bin/geth --softspoon \
  --mine --miner.threads 1 \
  --miner.etherbase 0xYOUR_REWARD_ADDRESS
```

Difficulty after the Soft Spoon follows the standard Homestead dynamic adjustment
(no difficulty bomb), so it tracks the real network hashrate automatically.

---

<a name="zh"></a>

<p align="right"><a href="#en">English</a> | <strong>中文</strong></p>

# Soft Spoon — 节点部署指南

> Soft Spoon 是一个艺术项目：把以太坊在 theDAO 合约部署前的区块处做 Soft Spoon（Soft Spoon 首块
> **1428757**）。本指南面向**希望在现有链上运行 / 挖矿的节点运营者**，
> **不**涉及一次性的建链（截断 / 铸造）——那部分已由项目方完成，你只需获取链数据并
> 运行节点即可。

## 1. 网络参数

| 项目 | 取值 |
|------|------|
| 网络名（flag） | `--softspoon` |
| Chain ID | `2517` |
| Network ID | `2517` |
| 共识 | Ethash PoW |
| Genesis 哈希 | `0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3` |
| Soft Spoon 首块（首个自出块） | `1428757` |
| Soft Spoon 首块 `1428757` 哈希 | `0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564` |
| Soft Spoon 首块 `1428757` 难度 | `1048576`（`0x100000`） |
| 可信检查点 | section `43`，head `0xade01e713d874b87dc6de44db12fda26963b38ca9b83cc4dc764fb7c8548d762`（区块 `1441791`） |
| 默认数据目录 | macOS `~/Library/Ethereum/softspoon`，Linux `~/.ethereum/softspoon` |
| Bootnodes | `enode://ef794a99…bb4564@95.217.104.247:30304`<br>`enode://c4d03b5f…c0b5f@188.40.138.215:30304` |

## 2. 编译

需要 Go 1.21+、C 工具链（gcc/clang）、git、make。

```bash
git clone https://github.com/CryptoZR/SoftSpoon.git
cd SoftSpoon
make geth
# 二进制位于 ./build/bin/geth
```

## 3. 获取链数据

你需要拿到包含 Soft Spoon 首块 `1428757` 及之后的链数据，两种方式：

### 方式 A — 从发布的链镜像还原（推荐，最快）

下载 `softspoon.tar.gz`：

- Google Drive：https://drive.google.com/file/d/1CSgA-Qf_QUSfUmDCSJaXLtMvzDslojay/view?usp=sharing
- 百度网盘：https://pan.baidu.com/s/1x9wlD09ymku5w-Abs6u-bA?pwd=2517 （提取码 `2517`）

压缩包内含一个 `softspoon/` 目录（`geth/`），它正是 `geth --softspoon` 默认使用的数据目录。
按你的操作系统解压到对应的数据根目录即可——之后**无需** `--datadir`。

```bash
# macOS —— 数据根目录 ~/Library/Ethereum
tar -xzf softspoon.tar.gz -C ~/Library/Ethereum

# Linux —— 数据根目录 ~/.ethereum
tar -xzf softspoon.tar.gz -C ~/.ethereum
```

解压后得到：`~/Library/Ethereum/softspoon/geth`（macOS）或
`~/.ethereum/softspoon/geth`（Linux）。

### 方式 B — 从网络同步

通过项目 bootnode 同步。信任由编译进二进制的硬编码 `TrustedCheckpoint` 锚定，
因此 snap 同步是安全的。

```bash
./build/bin/geth --softspoon \
  --bootnodes "enode://ef794a991152c3bb9f6f659a631b7b244898a196daf63304c8a863168e6be80d480a4cb0337575e754fd461b26b191b0b444dd094be212085bfd296235bb4564@95.217.104.247:30304,enode://c4d03b5f58fea266160bec11507f8b03e71523ea5dcb8b769b8a93b6010894f9821082cf4b7d638d94d6a50718673e1c06bc8c022ba682d9020858b2bc0c0b5f@188.40.138.215:30304,enode://54ca4dd12a7ebefcc34c1e35ee5acd213a004130b9239f38b3fcbb2ca2493d5b88616b53fb22f2afe0989474ca1f5c8af939a5b1d10857e47221fc7649e4ec3e@95.217.201.92:30303" \
  --syncmode snap
```

## 4. 运行节点

[![asciicast](https://asciinema.org/a/8up4tdJ0VWMOMzuJ.svg)](https://asciinema.org/a/8up4tdJ0VWMOMzuJ)

```bash
./build/bin/geth --softspoon \
  --http --http.api eth,net,web3 --bootnodes "enode://ef794a991152c3bb9f6f659a631b7b244898a196daf63304c8a863168e6be80d480a4cb0337575e754fd461b26b191b0b444dd094be212085bfd296235bb4564@95.217.104.247:30304,enode://c4d03b5f58fea266160bec11507f8b03e71523ea5dcb8b769b8a93b6010894f9821082cf4b7d638d94d6a50718673e1c06bc8c022ba682d9020858b2bc0c0b5f@188.40.138.215:30304,enode://54ca4dd12a7ebefcc34c1e35ee5acd213a004130b9239f38b3fcbb2ca2493d5b88616b53fb22f2afe0989474ca1f5c8af939a5b1d10857e47221fc7649e4ec3e@95.217.201.92:30303"
```

验证你在正确的链上（下方为默认 IPC 路径）：

```bash
# macOS
./build/bin/geth attach ~/Library/Ethereum/softspoon/geth.ipc
# Linux：~/.ethereum/softspoon/geth.ipc
> eth.chainId()                 // 2517
> eth.getBlock(1428757).hash    // 0xd4f997...6564
```

## 5. 挖矿

Soft Spoon 支持 CPU / 单卡挖矿。

```bash
./build/bin/geth --softspoon \
  --mine --miner.threads 1 \
  --miner.etherbase 0x你的收款地址
```

Soft Spoon 之后的难度采用标准 Homestead 动态调整（无难度炸弹），会自动跟随网络真实算力。
