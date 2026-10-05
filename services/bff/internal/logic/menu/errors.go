package menu

import "errors"

// errNoIdentity 取不到调用方身份。
//
// 只在一种情况下出现:**该路由没挂 Auth 中间件**(配置错误),
// 而不是用户的问题。故它是普通 error,经 response.Failure 落到 500 ——
// 让运维看见配置漏了,而不是静默返回空数据。
//
// Auth 中间件自己拒绝无令牌/坏令牌时回的是 401,走不到这里。
//
// **每个 logic 包各有一份同样的定义**。Go 的包内标识符不跨包,
// 而这些包都要判身份;放进 converter 包不合适(那是映射层),
// 抽共享 errors 包又会让"哪个包用哪个错误"变得不明显。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errNothingToUpdate 局部更新请求里没有任何可更新的字段。
//
// 见 role 包 helpers.go 的详细说明(两种情况:一个字段都没传,
// 或传的字段名都不认识 —— go-zero 的 httpx.Parse 会静默丢弃
// 未声明的字段)。
var errNothingToUpdate = errors.New("请求参数错误")
