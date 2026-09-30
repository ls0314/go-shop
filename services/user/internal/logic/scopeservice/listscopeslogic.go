package scopeservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListScopesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListScopesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScopesLogic {
	return &ListScopesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListScopesLogic) ListScopes(in *v1_userv1.ListScopesReq) (*v1_userv1.ListScopesResp, error) {
	page := int(in.Page)
	if page <= 0 {
		page = 1
	}

	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	scopes, total, err := l.svcCtx.ScopeRepo.GetScopeList(page, pageSize, in.ResourceType)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Scope, 0, len(scopes))
	for _, s := range scopes {
		items = append(items, converter.ToProtoScope(s))
	}

	return &v1_userv1.ListScopesResp{
		Items: items,
		Total: total,
	}, nil
}
