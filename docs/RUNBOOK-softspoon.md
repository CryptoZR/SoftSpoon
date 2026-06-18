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
