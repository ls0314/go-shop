// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"net/http"

	"demo-shop/services/bff/internal/logic/product"
	"demo-shop/services/bff/internal/response"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func WithdrawProductHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProductIdReq
		// 解析/校验失败一律回 400 + 通用文案。
		//
		// **不透传 httpx.Parse 的原始错误**:它的文案形如
		// "field receiver_name is not set" / "type mismatch" ——
		// 那是给开发者看的,不是给前端的。单体在这一层也回通用文案
		// (model.StatusBadRequest)。原始错误进日志,便于排查
		// "前端说参数没错"这类争议。
		//
		// **不用 httpx.ErrorCtx 写响应**:它在配了 error handler 时会
		// 自己写一次,导致后面的 BadRequest 因 WriteHeader 重复而失效。
		if err := httpx.Parse(r, &req); err != nil {
			logx.WithContext(r.Context()).Errorf("请求参数解析失败: %v", err)
			response.BadRequest(w, http.StatusBadRequest, response.InvalidParamMessage)
			return
		}

		l := product.NewWithdrawProductLogic(r.Context(), svcCtx)
		resp, err := l.WithdrawProduct(&req)
		if err != nil {
			// 状态码与文案由 response.Failure 决定:
			// 下游不可用 → 503,凭据无效 → 401,业务失败 → 400,其余 → 500。
			//
			// 不用 httpx.ErrorCtx —— go-zero 把它一律翻成 500,
			// 与单体契约(业务失败回 400)不符。
			response.Failure(w, err)
		} else {
			// 信封 {code,message,data} 由 response.OK 封。
			//
			// 无返回值(returns (Empty))的路由传 nil —— 得到
			// "data": null,与单体 utils.Success(c, nil) 一致。
			// 生成物默认走 httpx.Ok(w),那写的是 200 + 空 body,
			// 前端会因 res.data 为 undefined 而报错。
			response.OK(w, resp)
		}
	}
}
