// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package address

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAddressLogic {
	return &CreateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateAddress 新增收货地址。
//
// ============================================================
// 响应是**裸数字** —— 这是全项目唯一一个这样的接口
// ============================================================
//
// 单体: utils.Success(c, addressId)
//
// data 直接就是那个 int64:
//
//	{ "code":200, "message":"Success", "data": 7 }
//
// 故 .api 里这条是 `returns (int64)`,生成
// `(resp int64, err error)` —— **不是** *int64,也**不是**包一层
// {address_id: 7}。
//
// 前端解的是 `res.data`(数字本身)。若这里包一层,前端会拿到 NaN
// 或 undefined,新增地址后的跳转/选中逻辑会失败。
//
// ============================================================
// user_id 从 JWT 取,不接受请求体
// ============================================================
//
// proto 的 CreateAddressReq 有 user_id 字段,但那**不是**给客户端传的:
// 服务端拿它做 `WHERE user_id = ?` 的数据隔离。
//
// 若用请求体里的 user_id,任何登录用户都能往别人的地址簿里加地址。
//
// ============================================================
// is_default 的语义(proto 注释)
// ============================================================
//
//	"is_default 指定为默认地址;
//	 若该用户还没有任何地址,服务端**强制**设为默认(忽略此字段)——
//	 否则会得到一个'有地址但没有默认地址'的状态,而结算页要预选默认地址"
//
// 即:**第一条地址一定是默认**,客户端传 false 也无效。
//
// BFF 不做这个判断(要知道"是不是第一条"得先查,那是服务端的事)。
// 直接透传客户端的值,由服务端覆盖。
func (l *CreateAddressLogic) CreateAddress(req *types.CreateAddressReq) (int64, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return 0, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.AddressRPC.CreateAddress(ctx, &v1_userv1.CreateAddressReq{
		// 归属由服务端决定,不信客户端
		UserId: userId,

		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		AddressTag:    req.AddressTag,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return 0, err
	case rpc.ResultBiz:
		// 手机号格式错 / 姓名或手机号为空 / 已达地址数上限(20 条)→ 400
		return 0, err
	}

	return resp.GetAddressId(), nil
}
