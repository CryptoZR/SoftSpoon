# Soft Spoon 运行手册

> 所有操作在 ETC archive 库的**副本**上进行，切勿改动已同步的原库。
> 只需一个二进制：`make geth` 编译出的 `build/bin/geth`。它**同时支持** `--classic`（沿用未改动的 ClassicChainConfig）和 `--softspoon`。截断步骤用 `--classic`、铸造步骤用 `--softspoon`，无需另外准备"原版"二进制。
>
> 为什么截断用 `--classic` 而非 `--softspoon`：截断时库里存的链配置就是 classic，用 `--classic` 运行配置完全一致、不触发任何兼容性回滚；之后切到 `--softspoon` 时，两套配置的首个差异点在 1428757（> 截断后的 head 1428756），故同样不会强制回滚。

## 0. 前提
- ETC archive node 已同步至 1428756 之后，数据在 `~/Library/Ethereum/classic/geth`。
- 1428756 的区块哈希可用原库查得（见步骤 2）。

## 1. 复制数据库
```bash
cp -a ~/Library/Ethereum/classic ~/Library/Ethereum/softspoon-work
```

## 2. 用 `--classic` 截断到 1428756
启动控制台（同一个 `build/bin/geth`，这一步用 `--classic`）：
```bash
build/bin/geth --classic --datadir ~/Library/Ethereum/softspoon-work \
  --syncmode full --gcmode archive --maxpeers 0 --nodiscover console
```
在控制台中：
```javascript
// 确认 1428756 的哈希（信任锚的父块）
eth.getBlock(1428756).hash
// 回滚链头到 1428756，删除其后的真实 ETC 区块
debug.setHead("0x15CD14")   // 0x15CD14 == 1428756
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

> 重要：本版 core-geth 已移除 `light/`、`les/` 轻客户端栈，**不存在** CHT 生成工具
> （`ChtIndexer` / `checkpoint-admin` 都没有），也**无法**计算 `CHTRoot` / `BloomRoot`。
> 全节点的 checkpoint 校验（`eth/handler.go`）只使用 `SectionIndex` 与 `SectionHead`
> 两个字段——它在 snap sync 时向对端索要 `SectionHead` 高度的 header 并校验哈希一致，
> `CHTRoot`/`BloomRoot` 从不被读取，留零值即可。钉死 `SectionHead` 的哈希即沿
> parent-hash 链条锚定了包含 1428757 在内的整条历史。

CHT 每段 `CHTFrequency = 32768`。包含 1428757 的是 **section 43**（覆盖 `[1409024, 1441791]`），
其 SectionHead 高度 = `(43+1)×32768 − 1` = **1441791**。

步骤：
1. 挖到区块 **1441791**（单矿工私链无重排，head 一到即可，无需额外确认数）。
2. 取该块哈希：
   ```javascript
   eth.getBlock(1441791).hash   // 记为 <SECTIONHEAD_HASH>
   ```
3. 回填 `params/config_softspoon.go` 的 `PreDAOForkChainConfig`：
   ```go
   TrustedCheckpoint: &ctypes.TrustedCheckpoint{
       SectionIndex: 43,
       SectionHead:  common.HexToHash("0x<SECTIONHEAD_HASH>"),
       // CHTRoot / BloomRoot: 全节点不使用，留零值
   },
   ```
   （需在 `config_softspoon.go` 导入 `common` 与 `ctypes`。）
4. `make geth` 重新编译并发布。

## 发布清单
- [x] 1428757 规范哈希：`0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564`
- [x] 1428757 难度（== `softSpoonForkInitDifficulty`）：`1048576`（`0x100000`）
- [ ] 链镜像文件：__________
- [ ] TrustedCheckpoint 已回填（SectionIndex=43, SectionHead=block 1441791 hash）并重新编译：是 / 否
