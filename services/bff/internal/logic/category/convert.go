package category

import (
	"encoding/json"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/types"
)

// ============================================================
// 分页兜底
// ============================================================
//
// 与单体 handler 的 DefaultQuery 口径一致(category_handler.go 用
// c.DefaultQuery("page","1") / ("pageSize","10"))。
//
// **必须兜底**:.api 的 page/pageSize 是 optional int,不传时为 0,
// 而 proto 传 0 会让服务端算出的 offset 变成负数((0-1)*10 = -10)。
// 那不是"第一页"该有的样子。
//
// ============================================================
// 查询参数名是 pageSize(camelCase)—— **与其它域相反**
// ============================================================
//
// 单体是 c.DefaultQuery("pageSize", "10"),前端 GetCategoryListParams
// 也是 pageSize;而 orders / coupons / rbac 那些域用 page_size。故
// .api 的 CategoryListReq 写 form:"pageSize"。
//
// **请求与响应两侧都是 camelCase**:响应的 json tag 是 pageSize
// (见 types.CategoryListResp)。不要"顺手统一"成 snake_case ——
// 前端同时依赖这两侧。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
//
// maxPageSize 是**行为收紧**:单体没有上限,传 pageSize=100000 就会把
// 整张类目表拉出来。类目表的量级是百,100 远高于正常使用所需。
//
// 回写响应时用的是**兜底后**的值,而单体回显的是 DefaultQuery 的原文
// (传 page=0 时单体把 0 也一起回给前端,分页控件会拿到 0)。这里回 1
// 更符合前端的期望:拿到的永远是有效页码。
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
// proto 列表项 → types 列表项
// ============================================================
//
// 这两个结构体**逐字段一致**(category_id / parent_id / category_name /
// category_level / sort_order / is_leaf / is_visible / status),与单体
// GetListCategoryResp 也一致 —— 没有要改名、改类型或隐藏的字段。
// cart 域那种"改名 + 改类型"的坑在类目域不存在,但**空集合的口径仍然
// 要守**(见下)。

func toCategoryItem(p *v1_productv1.CategoryListItem) types.CategoryListItem {
	if p == nil {
		return types.CategoryListItem{}
	}
	return types.CategoryListItem{
		CategoryId:    p.GetCategoryId(),
		ParentId:      p.GetParentId(),
		CategoryName:  p.GetCategoryName(),
		CategoryLevel: p.GetCategoryLevel(),
		SortOrder:     p.GetSortOrder(),
		IsLeaf:        p.GetIsLeaf(),
		IsVisible:     p.GetIsVisible(),
		Status:        p.GetStatus(),
	}
}

// toCategoryItems 列表与子类目两个接口共用。
//
// make(..., 0, len) 而不是 var x []T —— 后者无数据时是 nil,序列化成
// null 而非 [],而前端有 list.length / .map() 这类写法。
func toCategoryItems(ps []*v1_productv1.CategoryListItem) []types.CategoryListItem {
	out := make([]types.CategoryListItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, toCategoryItem(p))
	}
	return out
}

// ============================================================
// proto 类目树 → types 类目树(递归)
// ============================================================
//
// **没有 parent_id**:单体在 GetTreeCategoryResp 上标了 json:"-"
// (树里父节点由层级隐含),proto 的 CategoryTreeNode 同样没有这个字段
// —— 两者正好一致。BFF 不要"顺手"从别处补一个,前端拿到多余的键会
// 误以为树是可以自解释的。
//
// children 是递归 message,这里递归转换。空的 children 必须是 []
// 而不是 null:前端对每个节点都会做 children.length / .map()。

func toCategoryTreeItem(p *v1_productv1.CategoryTreeNode) types.CategoryTreeItem {
	// 生成物的 getter 对 nil 接收者是安全的,故先取再判不用分两步
	children := p.GetChildren()

	out := types.CategoryTreeItem{
		CategoryId:    p.GetCategoryId(),
		CategoryName:  p.GetCategoryName(),
		CategoryLevel: p.GetCategoryLevel(),
		IsVisible:     p.GetIsVisible(),
		Status:        p.GetStatus(),
		// 叶子节点也必须是 [] 而不是 null
		Children: make([]types.CategoryTreeItem, 0, len(children)),
	}
	for _, c := range children {
		out.Children = append(out.Children, toCategoryTreeItem(c))
	}
	return out
}

func toCategoryTreeItems(ps []*v1_productv1.CategoryTreeNode) []types.CategoryTreeItem {
	out := make([]types.CategoryTreeItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, toCategoryTreeItem(p))
	}
	return out
}

// ============================================================
// proto 类目详情 → types 类目详情
// ============================================================
//
// 同样逐字段一致(十个字段与单体 GetCategoryResp 全对齐),
// 故这里只是搬运。

func toCategory(p *v1_productv1.Category) *types.Category {
	if p == nil {
		// 成功路径上 proto 一定带 category("查不到"是 error_msg → 400,
		// 已经在 logic 里返回了)。走到这里只可能是服务端实现回归:
		// 回了空 category 却也没带 error_msg。
		//
		// 返回全零对象而不是 nil —— nil 会序列化成 "data": null,
		// 前端读 category_id 时抛错;与 cart 域 toCartItem 的兜底一致。
		// 代价:这种回归在 HTTP 层是静默的(一个全零类目),排查要看日志。
		return &types.Category{}
	}
	return &types.Category{
		CategoryId:    p.GetCategoryId(),
		ParentId:      p.GetParentId(),
		CategoryName:  p.GetCategoryName(),
		CategoryLevel: p.GetCategoryLevel(),
		CategoryPath:  p.GetCategoryPath(),
		SortOrder:     p.GetSortOrder(),
		IconUrl:       p.GetIconUrl(),
		IsLeaf:        p.GetIsLeaf(),
		IsVisible:     p.GetIsVisible(),
		Status:        p.GetStatus(),
	}
}

// ============================================================
// 局部更新:types 的字段 → updates_json
// ============================================================
//
// proto 的 UpdateCategoryReq 只有 {category_id, updates_json},而 HTTP
// 请求有六个可选字段。合并语义在服务端:product-service 把 updates_json
// 解成 map,用 mapstructure(TagName: "json")合并进实体 —— **传了才改**。
//
// 故这里拼 JSON 用的键名必须是**服务端实体的 json tag**
// (category_name / sort_order / icon_url / is_visible / status /
// parent_id),不是 proto 字段名,也不是 gorm 列名。
//
// ============================================================
// presence 由**指针**承担 —— nil 进不去,非 nil 才进
// ============================================================
//
// .api 里这六个字段全是 `*T`(见 bff.api 的 UpdateCategoryReq 注释),
// 所以判断"要不要放进 updates_json"就是判断指针是否为 nil:
//
//	nil    → 前端没传 → 不进 JSON → 服务端不动这个字段
//	非 nil → 前端传了 → 进 JSON   → 服务端按值改(哪怕值是零值)
//
// 这正是单体靠 map[string]interface{} 天然得到的行为
// (category_handler.go:232 把 body 绑成 map 后整体转发,map 保留
// "键是否出现")。goctl 生成的结构体做不到 —— 非指针字段会把"没传"
// 与"零值"抹成同一个值,故信息只能靠指针在绑定层保住。
//
// **指针让两处本来无解的操作变得可达**:
//
//	{"is_visible": false} → 能隐藏类目
//	{"parent_id": 0}      → 能把类目移回根层级
//
// 非指针写法下 false 与 0 都是"没传"的哨兵,这两个操作表达不出来。
// (menu 的 is_visible/is_cache、dept 的 leader_id/sort_order、role 的
// is_default 用同样的手法,是既有先例。)
//
// 注意**空串仍然是一个有效值**:`{"category_name": ""}` 会被放进 JSON,
// 于是"用空串清空字段"也可达 —— 若服务端对空串有校验,那由它拒绝并
// 回业务错误,而不是 BFF 悄悄丢掉这个操作。

// field 一个待写进 updates_json 的键值对。
//
// 用有序切片而不是 map:让"哪些字段会被发出去"在调用处一眼可见
// (compactFields 的入参就是字段清单),与 rbac / menu 域
// strField + compactFields 的写法保持一致。
type field struct {
	name  string
	value interface{}
}

// strPtrField 字符串字段:nil 不进 JSON,非 nil 就带上(含空串)。
func strPtrField(name string, v *string) *field {
	if v == nil {
		return nil
	}
	return &field{name: name, value: *v}
}

// intPtrField 整数字段:nil 不进 JSON,非 nil 就带上(含 0)。
//
// **0 在这里是合法取值**:parent_id=0 表示"移到顶层",sort_order=0
// 表示"排到最前"。非指针写法会把它们当成"没传"丢掉。
func intPtrField(name string, v *int64) *field {
	if v == nil {
		return nil
	}
	return &field{name: name, value: *v}
}

// boolPtrField 布尔字段:nil 不进 JSON,非 nil 就带上(含 false)。
//
// false 是合法取值("隐藏类目"),非指针写法无法把它与"没传"区分。
func boolPtrField(name string, v *bool) *field {
	if v == nil {
		return nil
	}
	return &field{name: name, value: *v}
}

// compactFields 丢掉 nil 项(未传的字段不进 updates_json)。
func compactFields(fields ...*field) []field {
	out := make([]field, 0, len(fields))
	for _, f := range fields {
		if f == nil {
			continue
		}
		out = append(out, *f)
	}
	return out
}

// encodeUpdates 把字段清单编码成 JSON 对象文本。
//
// **空清单是合法的**:六个字段一个都没传(或只传了路径参数)时得到
// `{}`,服务端不做任何修改。这与 order / role 域的 errNothingToUpdate
// 不同 —— 那两个域拒绝空更新,本域不拒绝(单体也不拒绝:空 body 转发
// 过去就是空操作)。迁移期保持与单体一致,避免把"前端漏带字段"变成
// 一次 400 而掩盖其它问题。
//
// 返回 error 而不是忽略:map 里只有 string/int64/bool,现在不可能失败;
// 保留它,是为了将来往 field 里加类型时不会**静默丢掉**一次更新。
// 这种失败是编码错误,落到 response.Failure 的 500 分支正合适。
func encodeUpdates(fields []field) (string, error) {
	m := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		m[f.name] = f.value
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
