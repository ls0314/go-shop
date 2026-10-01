package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshToken 用 refresh 令牌换新的 access 令牌,refresh 令牌不轮换。
func (l *RefreshTokenLogic) RefreshToken(in *v1_userv1.RefreshTokenReq) (*v1_userv1.RefreshTokenResp, error) {
	// ParseRefreshToken 已包含验签与"类型必须是 refresh"两项校验
	claims, err := l.svcCtx.JWT.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		return &v1_userv1.RefreshTokenResp{ErrorMsg: err.Error()}, nil
	}

	accessToken, err := l.svcCtx.JWT.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		return nil, err
	}

	return &v1_userv1.RefreshTokenResp{AccessToken: accessToken}, nil
}
