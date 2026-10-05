// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateUserInfo 更新指定用户的档案(**仅限本人**)。
//
// 归属校验与 GetUserInfo 同一口径:令牌身份必须等于路径 :id,
// 否则回 400 + "用户越权修改地址"(单体文案,见 convert.go 的说明)。
//
// ============================================================
// 为什么入参是"指针字段"而不是 map
// ============================================================
//
// 单体这一层收的是 map[string]interface{}:
//
//	var updateUserInfo map[string]interface{}
//	c.ShouldBind(&updateUserInfo)
//	ui.userRPC.UpdateUserProfile(id, updateUserInfo)
//
// 即**任意字段都能传**,由服务端自己挑认识的。BFF 改成了
// .api 里声明的具体字段的指针(nickname / real_name / gender /
// avatar_url / birthdate 各一个 *string)。
//
// 这是一次**契约收紧**,影响:
//
//	好处:客户端传错字段名会被 httpx.Parse 丢弃,而不是静默无效;
//	     且 BFF 能在这里做类型/格式校验(如 birthdate 解析)
//	代价:客户端**无法把某字段更新为空串** —— nil 指针表示"不更新"
//	     (这是它相对 map 的语义改进:map 里 "" 与"没传"同样分不清)
//
// 对档案域来说这个代价可以接受:昵称/姓名/头像清空为空串没有业务意义,
// 真要清空应当是服务端的显式语义(如传 null 走另一条路径)。
//
// ============================================================
// birthdate 的字段名映射
// ============================================================
//
// .api 里是 birthdate(RFC3339 字符串),proto 的 FieldUpdate
// field 名也用 "birthdate" —— 与 DB 列名、json tag 一致
// (model.UserProfile.Birthdate 的 tag 是 json:"birthdate")。
func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoReq) (*types.UpdateUserInfoResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	if userId != req.Id {
		return nil, errUserInfoForbidden
	}

	// birthdate 需要先把 RFC3339 字符串转成 time.Time 再放进
	// FieldValue —— 但 FieldValue 的 oneof 只有 string / int64 / bool
	// 三种,**没有时间类型**。
	//
	// 故这里传**规范化后的 RFC3339 字符串**,由 user-service 解析。
	// 归一化(统一成 UTC、补齐格式)在 BFF 做的价值是:
	// 格式错误在这里就能报 400,而不用等下游。
	var birthdate *string
	if req.Birthdate != nil {
		ts, err := parseBirthdate(*req.Birthdate)
		if err != nil {
			return nil, err
		}
		if ts != nil {
			s := formatTimestamp(ts)
			birthdate = &s
		}
	}

	// 只有非 nil 的字段进 updates(见 convert.go 的 compactFields)
	updates := compactFields(
		strField("nickname", req.Nickname),
		strField("real_name", req.RealName),
		strField("gender", req.Gender),
		strField("avatar_url", req.AvatarUrl),
		strField("birthdate", birthdate),
	)

	// 一个字段都没传 → 不调下游,直接报参数错误。
	//
	// 为什么要特判:proto 的 UpdateUserProfileReq 若 updates 为空,
	// user-service 要么返回"无变更"要么报错 —— 两种都不理想。
	// 空更新请求在语义上是客户端的错误,在 BFF 拦掉更清晰,
	// 也省一次 RPC。
	//
	// **注意这与"只传了一个不认识的字段"是同一种情况** ——
	// 那个字段会被 httpx.Parse 丢弃,于是也走到这里。
	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.UpdateUserProfile(ctx, &v1_userv1.UpdateUserProfileReq{
		// **注意入参名是 user_profile_id 而不是 user_id**
		// (proto 如此),而路径参数 :id 在单体语义里是 user_id
		// (因为调用方传的是 userId)。
		//
		// 这是既有的一个不一致:单体 UpdateUserProfile(id, updates)
		// 里的 id 来自路径,直接当 user_profile_id 用了。
		// BFF 保持原样传,不做转换 —— 若服务端期望的是档案主键,
		// 那既有的调用方也是这么错的,一起改才是对的时机。
		UserProfileId: req.Id,
		Updates:       updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return toUpdateUserInfoResp(resp.GetProfile()), nil
}
