package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSelfPermCodesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSelfPermCodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSelfPermCodesLogic {
	return &ListSelfPermCodesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListSelfPermCodes 取用户全部权限码,供前端按钮级权限控制。
// 直接查库不缓存:登录后调一次,不在请求热路径上。
func (l *ListSelfPermCodesLogic) ListSelfPermCodes(in *v1_userv1.ListSelfPermCodesReq) (*v1_userv1.ListSelfPermCodesResp, error) {
	codes, err := l.svcCtx.PermRepo.GetPermCodesByUserId(in.UserId)
	if err != nil {
		return nil, err
	}
	if codes == nil {
		codes = []string{}
	}
	return &v1_userv1.ListSelfPermCodesResp{PermCodes: codes}, nil
}
