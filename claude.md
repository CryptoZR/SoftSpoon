### 项目介绍
Soft Spoon是一个基于以太坊公链的艺术项目，作品将以太坊在theDAO合约部署前的一个区块高度（1428756）上进行分叉，分叉时所有账户状态需要与当时主网保持一致。刚分叉的初始状态以单一显卡级的算力即可启动挖矿，后续沿用动态难度调整规则。因ETH主网已转向POS，早期数据难以获得，故使用ETC Archive Node来获取当时的主网状态。

### 相关资源
1. 本地已安装core-geth并完成同步 ./build/bin/geth --classic --syncmode full --gcmode archive
2. Archive Node区块高度已同步至1428756之后，数据保存在~/Library/Ethereum/classic/geth文件夹内
3. 参考项目 https://github.com/ethereumpow

### 项目需求
1. 基于core-geth代码进行改造，使其符合作品要求
2. 区块高度1428757处启用EIP155，chainid修改为2517

### 链环境配置
// params/config.go

var PreDAOForkChainConfig = &ChainConfig{
    ChainID:             big.NewInt(2517),        // 自定义 chainId，防重放
    HomesteadBlock:      big.NewInt(1150000),       // 保持与主网一致
    DAOForkBlock:        nil,                       // 关键：设为 nil，不执行 DAO fork
    DAOForkSupport:      false,                     // 不支持 DAO fork
    EIP150Block:         nil,                       // 不启用后续硬分叉
    EIP155Block:         big.NewInt(1428757),       // 从分叉点开始启用 EIP155（用新 chainId 做重放保护）
    EIP158Block:         nil,
    ByzantiumBlock:      nil,                       // 不进入 Byzantium
    ConstantinopleBlock: nil,
    PetersburgBlock:     nil,
    IstanbulBlock:       nil,
    BerlinBlock:         nil,
    LondonBlock:         nil,
    // 没有 MergeNetsplitBlock，没有 TerminalTotalDifficulty
    // 永远保持 PoW
    Ethash: new(EthashConfig),
}

// consensus/ethash/consensus.go

var (
    ForkBlock          = big.NewInt(1428757)      // 第一个由我们出的块
    ForkInitDifficulty = big.NewInt(0x20000000)   // 536,870,912 ≈ 5.4e8，单卡可挖
)

func CalcDifficulty(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
    next := new(big.Int).Add(parent.Number, big1)

    switch {
    case next.Cmp(ForkBlock) == 0:
        // 分叉首块：丢弃父块的天量历史难度，重置为单卡可挖的低值
        return new(big.Int).Set(ForkInitDifficulty)

    case next.Cmp(ForkBlock) > 0:
        // 分叉之后：标准 Homestead 动态调整，去掉炸弹，不设上限
        return calcDifficultyForkHomestead(time, parent)

    default:
        // 分叉前的历史块：保持原算法，否则历史无法通过验证
        return calcDifficultyHomestead(time, parent)
    }
}

// calcDifficultyForkHomestead: 去掉难度炸弹的标准 Homestead 调整
// diff = parent_diff + parent_diff/2048 * max(1 - (t - t_parent)/10, -99)
// 不含指数炸弹项；不设上限；下限为 params.MinimumDifficulty
func calcDifficultyForkHomestead(time uint64, parent *types.Header) *big.Int {
    bigTime := new(big.Int).SetUint64(time)
    bigParentTime := new(big.Int).SetUint64(parent.Time)

    x := new(big.Int)
    y := new(big.Int)

    // 1 - (t - t_parent) / 10
    x.Sub(bigTime, bigParentTime)
    x.Div(x, big10)
    x.Sub(big1, x)

    // max(..., -99)
    if x.Cmp(bigMinus99) < 0 {
        x.Set(bigMinus99)
    }

    // parent_diff + parent_diff/2048 * x
    y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
    x.Mul(y, x)
    x.Add(parent.Difficulty, x)

    // 下限保护（无上限）
    if x.Cmp(params.MinimumDifficulty) < 0 {
        x.Set(params.MinimumDifficulty)
    }

    // 刻意不加 2^(periodCount-2) 难度炸弹项
    return x
}