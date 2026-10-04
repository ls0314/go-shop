package addressservicelogic

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetAddressSnapshotLogic 取下单用的地址快照。
//
// 与 GetAddress 分开的**关键理由是失败语义**:下单时把"地址不存在"
// 与"user-service 不可用"混在一起,会得到"用户地址填错了"这种误导性提示,
// 而实际上只是下游抖了一下(DS-A-25 §4.5.2 第 1 条专门点了这一条)。
//
// 所以这里:地址不存在 → ErrorMsg(业务失败,别重试);
// user-service 本身不可用 → gRPC error(调用方该重试)。
// 这个区分由"谁返回什么"承载 —— 本方法只负责前者。
type GetAddressSnapshotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressSnapshotLogic {
	return &GetAddressSnapshotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAddressSnapshotLogic) GetAddressSnapshot(in *v1_userv1.GetAddressSnapshotReq) (*v1_userv1.GetAddressSnapshotResp, error) {
	// 走与其它操作同一个归属校验:下单取快照同样是"读自己的地址",
	// 不校验就等于允许拿别人的地址下单
	addr, err := loadOwnedAddress(l.svcCtx.AddressRepo, in.GetUserId(), in.GetAddressId())
	if err != nil {
		if isBizError(err) {
			return &v1_userv1.GetAddressSnapshotResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	return &v1_userv1.GetAddressSnapshotResp{Snapshot: toProtoSnapshot(addr)}, nil
}
