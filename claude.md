### 项目介绍
Soft Spoon是一个基于以太坊公链的艺术项目，作品将以太坊在theDAO合约部署前的一个区块高度（1428756）上进行分叉，分叉时所有账户状态需要与当时主网保持一致。刚分叉的初始状态以CPU级的算力即可启动挖矿，后续沿用动态难度调整规则。因ETH主网已转向POS，早期数据难以获得，故使用ETC Archive Node来获取当时的主网状态。

### 相关资源
1. 本地已安装core-geth并完成同步 ./build/bin/geth --classic --syncmode full --gcmode archive
2. Archive Node区块高度已同步至1428756之后，数据保存在~/Library/Ethereum/classic/geth文件夹内
3. 参考项目 https://github.com/ethereumpow

### 项目需求
1. 基于core-geth代码进行改造，使其符合作品要求
2. 区块高度1428757处启用EIP155，chainid修改为2517
3. 先使用CPU级别算力出第一个块，后续需要将包含这个1428757高度的链做成镜像，以便其他节点进行快速同步

### 1428757区块数据
difficulty: 1048576
hash: "0xd4f997aca084bd361480b034adea2db292f079f542d52a718a04e71d671d6564"