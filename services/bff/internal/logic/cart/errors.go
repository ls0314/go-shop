package cart

import "errors"

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
// 每个 logic 包各有一份,理由见 role 包 helpers.go 的说明。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// ============================================================
// 购物车域没有"局部更新"需要的 FieldUpdate
// ============================================================
//
// 与 RBAC 域不同,购物车的更新是一条专门的 RPC:
//
//	UpdateCartItemReq{cart_item_id, user_id, quantity, is_selected}
//
// 它是**整条更新**而不是逐字段更新 —— 故不需要 strField/compactFields
// 那一套。这也意味着"只想改数量"时也必须把 is_selected 一起传:
//
//	若传零值 false,会把已勾选的商品**取消勾选**。
//
// 这是 .api 里 UpdateCartItemReq 两个字段都 optional 的原因 ——
// 但 optional 在 goctl 里只影响绑定(不传时为零值),**无法区分
// "没传"与"传了 false"**。故 logic 里必须做取舍,见
// updatecartitemlogic.go 的说明。
