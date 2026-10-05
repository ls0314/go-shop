package user

import (
	"errors"
	"fmt"
	"time"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/types"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================================
// 本文件只放 **user 域专用** 的东西
// ============================================================
//
// 跨域共用的实体映射(Role / Permission / Menu / Dept / Scope / User)
// 已经抽到 internal/converter —— 因为 role / permission / menu /
// scope / dept 五个包需要同一批映射,放在这里它们用不到
// (Go 的包内标识符不跨包)。
//
// 留在这里的是:
//
//	时间与分页的工具(被本包的 logic 直接用)
//	user 域特有的错误哨兵
//	局部更新用的 FieldUpdate 构造(其它域有自己的字段集)
//	user 域特有的响应映射(那几个 types.*Resp 只有本域用)

// ============================================================
// 时间
// ============================================================

// formatTimestamp 转调 converter.Timestamp。
//
// 保留这个薄封装而不是让各处直接写 converter.Timestamp:
// user 域的 20 个 logic 都调它,名字短一点、且将来要换格式化口径时
// 改一处。converter 那份是给别的域用的同一实现。
func formatTimestamp(ts *timestamppb.Timestamp) string {
	return converter.Timestamp(ts)
}

// parseBirthdate string → timestamppb.Timestamp。
//
// ============================================================
// 为什么是 RFC3339 优先,再兜 date-only
// ============================================================
//
// 单体的 model.UserProfile.Birthdate 是 *time.Time,json tag 为
// "birthdate" —— 也就是由 encoding/json 解析。而 encoding/json 对
// time.Time 只认 **RFC3339**(time.Time.UnmarshalJSON 内部就是
// time.Parse(`"`+time.RFC3339+`"`, ...))。
//
// 所以既有前端传的一定是 RFC3339,如 "1990-05-20T00:00:00Z"。
//
// 但生日字段在前端 date picker 里常常只有日期部分,故额外兜一个
// "2006-01-02" —— 这不是"顺手放宽",而是承认一个现实:
// 单体那边如果前端传了纯日期,gin 绑定也会报错(同一个 encoding/json),
// 所以兜它属于**行为增强**,不影响既有调用方。
//
// 时区:生日对"时刻"不敏感,统一按 UTC 解释并在返程格式化回
// UTC —— 与 formatTimestamp 的 UTC 口径一致,避免出现
// "存进去是 5-20、读出来是 5-19"这种跨时区偏移。
//
// 空串返回 nil(proto 的未设置语义),不报错 ——
// "不填生日"是合法的,不该因此拒绝整个请求。
func parseBirthdate(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, nil
	}

	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t.UTC()), nil
		}
	}

	// 两种都不匹配才报错。**带上原值**:前端拿到
	// "birthdate 格式错误: 1990/05/20" 才能立刻定位是自己传错了,
	// 只报"参数错误"会让人翻半天。
	return nil, fmt.Errorf("birthdate 格式错误: %s", s)
}

// ============================================================
// 分页参数兜底
// ============================================================

// defaultPage / defaultPageSize 与单体的 DefaultQuery 口径一致。
//
// 单体(handler/user_handler.go:182-183):
//
//	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
//	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
//
// **必须兜底,不能把 0 传下去**:.api 里分页字段是 optional int,
// 不传时为 0。而 proto 的 page/page_size 是 int32 —— 传 0 时
// user-service 要么按"无限制"返回全部(危险),要么返回空页
// (前端看到空列表)。两种都不是单体行为。
const (
	defaultPage     = 1
	defaultPageSize = 10
)

// maxPageSize 单页上限。
//
// 单体的 DefaultQuery 没有上限 —— 客户端传 pageSize=100000 就会
// 让服务端一次性捞全表。BFF 在这里加一道闸:
//
//	**这是一次行为收紧**,理由是它防的是"客户端写错参数打垮 DB",
//	而不是某个正常业务场景。100 远高于任何页面的展示需求
//	(用户列表页通常 10~50)。
//
// 若前端确实有"一次拉全部"的页面(如导出),需要单独走一个
// 不分页的接口,而不是把 pageSize 开到极大。
const maxPageSize = 100

// normalizePage 兜底并夹紧分页参数。
//
//	<= 0            → 用默认值(1 / 10)
//	pageSize 超上限 → 夹到 maxPageSize
//	page 超上限     → 不夹(查后面的页是合法请求)
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// ============================================================
// user 域的错误哨兵
// ============================================================

// errNoIdentity 取不到调用方身份。
//
// 只在一种情况下出现:**该路由没挂 Auth 中间件**(配置错误),
// 而不是用户的问题。故它是一个普通 error,
// 经 response.Failure 落到 500 —— 让运维看见配置漏了,
// 而不是静默返回空数据让前端困惑。
//
// 注意 Auth 中间件自己拒绝无令牌/坏令牌时回的是 401,
// 走不到这里;能走到这里说明中间件压根没执行。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errUserInfoForbidden 越权访问他人档案。
//
// 文案 **"用户越权修改地址"** 是单体的 model.UserInfoError ——
// 逐字照抄,不改。
//
// 这句文案本身有歧义(说的是"地址",却用在档案接口上),看起来像
// 复制粘贴留下的。但**它是既有契约**:用户看到的就是这句话。
// 迁移期擅自"修正"文案 = 改变前端可见行为,而这一阶段的目标是
// 前端零感知。要改就单独一次改动 + 通知前端,不要夹带在迁移里。
//
// 触发条件:JWT 里的 user_id 与路径 :id 不一致
// (单体 user_info_handler.go:70 的 userId != id)。
var errUserInfoForbidden = errors.New("用户越权修改地址")

// errNothingToUpdate 局部更新请求里没有任何可更新的字段。
//
// 两种情况都会落到这里:
//
//	① 客户端一个字段都没传(空 body 或全是 null)
//	② 客户端传了字段但**名字都不认识** —— go-zero 的 httpx.Parse
//	   会静默丢弃未声明的字段(它不报"未知字段"错)
//
// 归一成一句"请求参数错误":对客户端来说这两种都是"我这次调用
// 没有意义",而不是服务端的问题。文案与 handler 层
// (response.InvalidParamMessage)一致,故前端看到的提示统一。
//
// 为什么在 BFF 拦而不是转发:updates 为空时 user-service 要么
// 返回"无变更"要么报错,两种都不理想;而空更新在语义上确实是
// 客户端的错误。拦在这里更清晰,也省一次 RPC。
var errNothingToUpdate = errors.New("请求参数错误")

// ============================================================
// 局部更新:types 的指针字段 → proto 的 FieldUpdate
// ============================================================
//
// proto 的 UpdateUserReq / UpdateUserProfileReq 收的是
// repeated FieldUpdate,每个是 {field: "phone", value: {string_value: "..."}}
// (见 api/proto/user/v1/common.proto)。
//
// 用 oneof 而不是 optional 是那个 proto 的既定选择(注释写明:
// "proto3 的 optional 依赖字段存在性追踪,跨语言生成物差异较大")。
//
// **nil 指针 = 该字段不更新**:这是"局部更新"的关键 ——
// 客户端只传要改的字段,其余保持原值。若把 nil 当成"更新为空串",
// 一次改手机号就会把邮箱清掉。

// strField 构造 string 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func strField(name string, ptr *string) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_StringValue{StringValue: *ptr},
		},
	}
}

// int64Field 构造 int64 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func int64Field(name string, ptr *int64) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_Int64Value{Int64Value: *ptr},
		},
	}
}

// boolField 构造 bool 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func boolField(name string, ptr *bool) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_BoolValue{BoolValue: *ptr},
		},
	}
}

// compactFields 丢掉 nil 项。
//
// 为什么要它:上面三个构造函数对 nil 指针返回 nil,而
// repeated 字段里塞 nil 元素会让 user-service 收到一个
// {field: "", value: nil} 的空更新 —— 它要么报错要么写坏数据。
//
// 在 BFF 这一层过滤掉,比要求每个服务端都判空更可靠。
func compactFields(fields ...*v1_userv1.FieldUpdate) []*v1_userv1.FieldUpdate {
	out := make([]*v1_userv1.FieldUpdate, 0, len(fields))
	for _, f := range fields {
		if f != nil {
			out = append(out, f)
		}
	}
	return out
}

// ============================================================
// user 域特有的响应映射
// ============================================================
//
// 下面这些 types.*Resp 只被 user 域用,故不放进 converter 包。
//
// 三组"字段集相同但类型名不同"的映射(User / UserProfile 各 3~4 个)
// 看起来是重复,但**不能合成一个返回公共类型**:
// 那样 logic 的返回类型就对不上 .api 生成的签名了。
// 这是"每个响应用独立类型"的代价,换来的是
// "改一个接口的契约不会波及其他接口"。

func toUserItem(u *v1_userv1.User) types.UserItem {
	return converter.UserItem(u)
}

func toGetUserResp(u *v1_userv1.User) *types.GetUserResp {
	return &types.GetUserResp{UserItem: converter.UserItem(u)}
}

func toCreateUserResp(u *v1_userv1.User) *types.CreateUserResp {
	return &types.CreateUserResp{UserItem: converter.UserItem(u)}
}

func toUpdateUserResp(u *v1_userv1.User) *types.UpdateUserResp {
	return &types.UpdateUserResp{UserItem: converter.UserItem(u)}
}

func toCreateUserInfoResp(p *v1_userv1.UserProfile) *types.CreateUserInfoResp {
	return &types.CreateUserInfoResp{
		UserInfoId: p.GetUserInfoId(),
		UserId:     p.GetUserId(),
		Nickname:   p.GetNickname(),
		RealName:   p.GetRealName(),
		Gender:     p.GetGender(),
		AvatarUrl:  p.GetAvatarUrl(),
		Birthdate:  formatTimestamp(p.GetBirthdate()),
		CreatedAt:  formatTimestamp(p.GetCreatedAt()),
		UpdatedAt:  formatTimestamp(p.GetUpdatedAt()),
	}
}

func toGetUserInfoResp(p *v1_userv1.UserProfile) *types.GetUserInfoResp {
	return &types.GetUserInfoResp{
		UserInfoId: p.GetUserInfoId(),
		UserId:     p.GetUserId(),
		Nickname:   p.GetNickname(),
		RealName:   p.GetRealName(),
		Gender:     p.GetGender(),
		AvatarUrl:  p.GetAvatarUrl(),
		Birthdate:  formatTimestamp(p.GetBirthdate()),
		CreatedAt:  formatTimestamp(p.GetCreatedAt()),
		UpdatedAt:  formatTimestamp(p.GetUpdatedAt()),
	}
}

func toUpdateUserInfoResp(p *v1_userv1.UserProfile) *types.UpdateUserInfoResp {
	return &types.UpdateUserInfoResp{
		UserInfoId: p.GetUserInfoId(),
		UserId:     p.GetUserId(),
		Nickname:   p.GetNickname(),
		RealName:   p.GetRealName(),
		Gender:     p.GetGender(),
		AvatarUrl:  p.GetAvatarUrl(),
		Birthdate:  formatTimestamp(p.GetBirthdate()),
		CreatedAt:  formatTimestamp(p.GetCreatedAt()),
		UpdatedAt:  formatTimestamp(p.GetUpdatedAt()),
	}
}
