// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package scope

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScopeLogic {
	return &CreateScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateScope 创建数据权限规则(管理端)。
//
// ============================================================
// 数据权限是什么 —— 决定了这个域的字段语义
// ============================================================
//
// 它定义的是"某个角色能看到**哪些行**的数据",即 SQL 层面的过滤条件:
//
//	resource_type    作用在哪个资源上(如 "order" / "product")
//	field_name       按哪个字段过滤(如 "dept_id" / "creator_id")
//	condition_type   条件的种类(如 "in" / "eq" / "self")
//	condition_value  条件的值(如 "1,2,3" 或 "current_user")
//
// 拼出来的效果类似 `WHERE dept_id IN (1,2,3)`。
//
// **所以这几个字段是"能被拼进 SQL 的输入"** —— 服务端必须对
// condition_value 做白名单或参数化,否则这是注入面。
//
// BFF **不做校验**:它不知道服务端的拼接方式(白名单?参数化?
// 还是字符串拼接?),自己加一套规则只会得到"BFF 放过了、
// 服务端拒绝"或反过来的不一致。
//
// 但值得知道:**这个接口是安全敏感面**。若前端有"自定义条件值"
// 的自由输入框,那是一个需要服务端严格校验的入口。
//
// ============================================================
// role_id 是必填
// ============================================================
//
// 数据权限**挂在角色上** —— 没有"全局数据权限"这种东西
// (那等于对所有角色生效,应该直接改查询逻辑而不是加一条规则)。
//
// 故 .api 里 role_id 不是 optional。
func (l *CreateScopeLogic) CreateScope(req *types.CreateScopeReq) (*types.CreateScopeResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.CreateScope(ctx, &v1_userv1.CreateScopeReq{
		Scope: &v1_userv1.Scope{
			RoleId:         req.RoleId,
			ResourceType:   req.ResourceType,
			FieldName:      req.FieldName,
			ConditionType:  req.ConditionType,
			ConditionValue: req.ConditionValue,
			Description:    req.Description,
			// ScopeId 与 CreatedAt 不传(服务端生成)
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 / 规则重复 → 400
		return nil, err
	}

	// 响应是**裸对象**(单体: utils.Success(c, created))
	//
	// 注意 Scope 的 role 名不由本接口返回(proto 注释:
	// "角色名由前端用角色列表本地映射,不由本接口返回")——
	// 故前端要显示角色名,得自己拿角色列表对着 role_id 查。
	return &types.CreateScopeResp{
		ScopeItem: converter.ScopeItem(resp.GetScope()),
	}, nil
}
