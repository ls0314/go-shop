module demo-shop/pkg

go 1.25.0

// 版本**必须与 demo-shop-back 一致**:同一个令牌要在 BFF 与单体两处验签
// (过渡期两边并存),两个版本意味着两套解析行为 —— 而差异只在边界情况
// (如时钟偏移容忍、协议头校验)暴露,表现为"某些令牌在这里过、在那里被拒"。
require github.com/golang-jwt/jwt/v5 v5.3.1
