// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"
	"encoding/json"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateProduct 局部更新商品(只改 SPU 的基本字段,**不碰 SKU 与图片**)。
//
// ============================================================
// 契约是"逐字段合并",不是"整体替换"
// ============================================================
//
// proto 的 UpdateProductReq 只有 {spu_id, updates_json}:
//
//	updates_json 是一个 JSON 对象文本,键是**实体的 json tag 名**
//	服务端 DecodeUpdateJSON 解成 map 后,用 mapstructure 按 json tag
//	合并到 SysProductSpu 上(与类目域共用同一份实现)
//
// 所以 BFF 这一层的唯一职责是:**把非零值字段拼成那个 JSON 对象**。
// 拼装用 productUpdates 结构体(键名来自 json tag,不可能打错),
// 取舍见 productUpdatesFrom 的说明 —— 核心代价是"零值改不了"
// (brand 清不成空串、priority 改不成 0)。
//
// 想改 SKU / 图片 / 规格模板的,走 UpdateProductFull(见那边)。
//
// ============================================================
// 一个字段都没传 → 400(与 role / user / menu 域一致)
// ============================================================
//
// 见 errNothingToUpdate 的说明:单体对这种请求回 200 空操作,而
// go-zero 的 httpx.Parse **静默丢弃**未声明的字段,于是"前端字段名
// 打错"会退化成"空更新 + 200",用户以为改成功了。报 400 让它当场可见。
//
// ============================================================
// spu_status 只能在 withdrawn → draft 之间改
// ============================================================
//
// 服务端把状态机的判定握在手里:除 withdrawn → draft 之外的任何变更
// 都回 error_msg("状态转换不合法")。发布/下架请走 publish / withdraw
// 接口 —— 那两条有自己的前置校验(类目是启用的叶子、至少一个 active
// SKU、库存 > 0、价格 > 0),从"改商品"这里改状态会把校验整个跳过。
// BFF 不预检状态,也不替前端猜(预检要先查商品,那是服务端的库,
// 且查完到改之间是 TOCTOU)。
//
// ============================================================
// 服务端返回合并后的完整对象,但本接口只回 Empty —— 这是 .api 的选择
// ============================================================
//
// proto 的 UpdateProductResp 带着合并后的 ProductDetail(注释:"调用方
// 无需再查一次"),而 .api 里这条路由是 `returns (Empty)` → 前端拿到
// "data": {}。单体当年回的是合并后的对象,故这是**一处契约收紧**。
//
// 影响有限(前端更新后自己会重新拉详情/返回列表页),但若将来有调用方
// 直接读 res.data.data.spu_name,拿到的是 undefined。要"顺手"把它改成
// 返回详情是**契约增强**,应单独一次改动 + 通知前端,不要夹在迁移里做。
func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductReq) (*types.Empty, error) {
	updates := productUpdatesFrom(req)
	if updates.empty() {
		return nil, errNothingToUpdate
	}

	updatesJSON, err := json.Marshal(updates)
	if err != nil {
		// 全是 string/int64 字段,序列化不可能失败;留着是为了不吞错误。
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.UpdateProduct(ctx, &v1_productv1.UpdateProductReq{
		SpuId:       req.Id,
		UpdatesJson: string(updatesJSON),
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 状态转换不合法 / 已上架不许改规格模板 /
		// 目标类目不是启用的叶子 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
