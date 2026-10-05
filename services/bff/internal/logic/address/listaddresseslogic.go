// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package address

import (
	"context"

	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListAddresses 取当前登录用户的收货地址列表。
//
// ============================================================
// 响应是**裸数组**,不是 {list: [...], total: ...}
// ============================================================
//
// 单体 address_handler.go 是 utils.Success(c, addressList) ——
// data 直接就是数组:
//
//	{ "code":200, "message":"Success", "data": [ {...}, {...} ] }
//
// **地址域与 RBAC 域的列表形状不同**:RBAC 那边 handler 手工包了
// gin.H{"list": ..., "total": ...},而地址域没有。两者不能互相类推
// (这一点我先前推断错过一次,已按 handler 核实)。
//
// 故 .api 里这条是 `returns ([]AddressItem)`,生成
// (resp []types.AddressItem, err error)。
//
// **注意没有 total** —— 地址没有分页,一次返回全部
// (服务端限制每用户最多 20 条,见 AddressMaxCount)。
func (l *ListAddressesLogic) ListAddresses(req *types.Empty) (resp []types.AddressItem, err error) {
	// TODO: 依赖 AddressRPC.ListAddresses 与一个 AddressItem 的映射,
	// 两者都在第 6 步(地址域)写。这里先留桩。
	//
	// 之所以现在不写:地址域的 logic 属于第 6 步,而这一步的目标
	// 是修复 .api 并让整个模块编译通过 —— 故只需把签名对齐。
	//
	// 替换本方法体时用:
	//
	//	userId, ok := middleware.UserID(l.ctx)
	//	if !ok { return nil, errNoIdentity }
	//	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	//	defer cancel()
	//	resp, grpcErr := l.svcCtx.AddressRPC.ListAddresses(ctx,
	//		&v1_userv1.ListAddressesReq{UserId: userId})
	//	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	//	switch kind {
	//	case rpc.ResultInfra, rpc.ResultBiz:
	//		return nil, err
	//	}
	//	return converter.AddressItems(resp.GetItems()), nil
	return
}
