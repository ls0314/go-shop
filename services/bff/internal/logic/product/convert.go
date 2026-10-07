package product

import (
	"encoding/json"
	"errors"
	"fmt"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ============================================================
// 本域的四类差异(与类目/库存同源,写在 .api 顶部的那一段)
// ============================================================
//
//	① items → list    两个列表接口的 proto 字段叫 items,HTTP 叫 list
//	② JSON 文本 ↔ 对象 spec_values / spec_template 在 proto 是 string,
//	                  在 types 里是 interface{}
//	③ 管理端/用户端字段不同  AdminSku 有 cost_price / lock_stock,UserSku 没有,
//	                  而 proto 用**同一个 Sku message** 承载两者
//	④ 分页命名不一致 请求是 page_size(snake),响应是 pageSize(camel)
//
// ①③④ 是"照抄即对"的机械映射,② 是这个文件存在的主要理由。

// ============================================================
// ② JSON 文本 ↔ 对象
// ============================================================
//
// 本域有**两个** JSON 字段,方向与形状都不同:
//
//	spec_values    SKU 的规格键值对,形状是**对象** {"颜色":"红","尺码":"XL"}
//	spec_template  SPU 的规格模板,形状是**数组**
//	               [{"name":"颜色","values":["红","蓝"]}]
//
// 两者在 proto 里都是 string(JSON 文本,与 DB 的 JSONB 列一致 ——
// proto 注释:"JSON 文本,与 sys_product_spu.spec_template 一致"),
// 在 .api 的 types 里分别是 map[string]string 与 []SpecTemplateItem。
// 故需要三个函数:
//
//	响应方向  parseSpecValues   文本 → map[string]string
//	响应方向  parseSpecTemplate 文本 → []SpecTemplateItem
//	请求方向  jsonText          对象/数组 → 文本
//
// **三者的类型都必须与 .api 一致,不能图省事用 interface{}** ——
// spec_values / spec_template 所属的类型既是响应也是请求体(AdminSku /
// CreateSkuItem / UpdateProductFullReq),而 go-zero 的映射器不支持请求
// 方向出现 interface{}(见 parseSpecValues 与 SpecTemplateItem 的注释)。

// parseSpecValues 把 proto 的 JSON 文本解析成对象。
//
// 与 cart 域的同名函数**逐字同源**(那边是购物车行里的 sku 规格),
// 语义必须一致:同一份 spec_values 在两个接口里长得不一样,前端要为
// "购物车里的规格"和"商品详情里的规格"写两套取值逻辑。
//
// 解析失败 → 空对象 + 记日志、**不报错**。理由(照抄 cart 域的结论):
//
//	spec_values 只是规格展示,它坏掉说明 product-service 那条数据有问题,
//	但不该让**整个商品详情**失败 —— 用户该看到"这一个 SKU 没显示规格",
//	而不是"商品页打不开"。空对象({})而不是 nil:前端会做
//	spec_values.颜色 这类直接取值,nil 会抛错,{} 安全。
//
// **代价**:这类问题只会出现在日志里,不会以错误形式暴露 —— 故必须记日志。
//
// ============================================================
// 返回 map[string]string,不是 interface{}
// ============================================================
//
// 两个理由:
//
//  1. **契约本来就是这个形状**。前端 types/product.ts 里是
//     `spec_values: Record<string, string>` —— 规格名到规格值的映射,
//     值恒为字符串("红色"、"XL")。用 map[string]string 比 interface{}
//     更准确地表达了这一点。
//
//  2. **interface{} 在请求方向是坏的**,而本类型(`AdminSku`)是双用途的
//     —— 既做商品详情的响应,又做全量更新(UpdateProductFullReq)的请求
//     体。请求方向一旦是 interface{},go-zero 的映射器必然报
//     type mismatch(它的 Kind 是 reflect.Interface,与任何具体 Kind
//     都不相等)。详见 SpecTemplateItem 的注释。
//
// 故这里与请求侧统一成 map[string]string。购物车的同名函数仍返回
// interface{}(它只用于响应,不受这条约束)—— 那是两者唯一的差别,
// 语义(失败降级为空 map + 记日志)保持逐字一致。
func parseSpecValues(s string) map[string]string {
	// 空串是**正常情况**:product-service 的 jsonMapToText 把空 map
	// 写成空串(DB 该列 NOT NULL DEFAULT '{}',proto 侧用 "" 表示空)。
	// 走解析会得到一个 "unexpected end of JSON input" 并记一条无意义的日志。
	if s == "" {
		return map[string]string{}
	}

	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		// 不打印原始值:spec_values 是商品侧完全可控的输入,原样打进日志
		// 会让日志格式被外部数据影响(超长字符串、换行)。只打长度。
		//
		// 注意这里可能因**值不是字符串**而失败(如 {"数量": 2})——
		// 那属于数据形状问题,同样只记日志不报错:规格展示坏掉不该让
		// 商品详情打不开。
		logx.Errorf("product: spec_values 不是 string→string 的合法 JSON(len=%d): %v", len(s), err)
		return map[string]string{}
	}
	if m == nil {
		// `null` 是合法 JSON,Unmarshal 进 map 得到 nil。归一到 {} 保持一致。
		return map[string]string{}
	}
	return m
}

// parseSpecTemplate 把 proto 的 spec_template 解析出来。
//
// ============================================================
// 它**不能**像 spec_values 那样用 map 接 —— 这是本域最容易踩的一处
// ============================================================
//
// spec_template 的真实形状是**数组**。两处独立证据:
//
//	product-service 的 validateSpecSkuRules 直接 json.Unmarshal 成
//	[]specItem 再校验(name/values 两个字段、SKU 数=各规格取值数的笛卡尔积);
//	前端 types/product.ts 的声明也是 SpecTemplateItem[]。
//
// 若照 cart 域那样 Unmarshal 进 map[string]interface{},一个**完全合法**的
// 模板会得到 "cannot unmarshal array into Go value of type
// map[string]interface {}" —— 症状不是报错,而是:
//
//	日志里一片 spec_template 解析失败
//	接口返回 {}
//	前端的规格编辑器因此变空(它判 Array.isArray),一保存就把模板整个丢了
//
// 故这里 Unmarshal 进 []types.SpecTemplateItem,与 wire 形状一致。
//
// ============================================================
// 为什么不能返回 interface{} —— 以及它曾经导致的一次 400
// ============================================================
//
// 这个函数原本返回 interface{},而 `.api` 里 AdminProductResp.SpecTemplate
// 也是 interface{}。那是**响应**方向,走 encoding/json,没问题。
//
// 但**请求**方向不行:go-zero 的 httpx.Parse 走的是 core/mapping 的严格
// 映射器,而它的 processFieldPrimitive 是这样判的:
//
//	typeKind  := Deref(fieldType).Kind()      // interface{} → reflect.Interface
//	valueKind := reflect.TypeOf(mapValue).Kind()   // 数组 → reflect.Slice
//	if typeKind == valueKind { ... }
//	return newTypeMismatchError(fullName)
//
// Interface 与任何具体 Kind 都不相等 ⇒ **任何非 json.Number 的值都必然
// 报 type mismatch**。实测症状:
//
//	请求参数解析失败: type mismatch for field "spec_template"  → 400
//
// 所以 `.api` 里请求方向的字段**一律不能用 interface{}**,必须给出具体
// 形状(数组就写 []T,对象就写 map[string]string 或结构体)。
func parseSpecTemplate(s string) []types.SpecTemplateItem {
	if s == "" {
		// 空 → **空切片**(不是 nil、不是 {})
		//
		// 前端对 spec_template 一律先判 Array.isArray / .length,故必须是
		// 数组形状;nil 会被 encoding/json 序列化成 null,而前端的
		// `Array.isArray(null)` 是 false,会走到"没有规格模板"的分支 ——
		// 结果虽对,但契约上"有个数组、里面 0 项"更准确。
		return []types.SpecTemplateItem{}
	}

	var items []types.SpecTemplateItem
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		// 静默降级 + 记日志,理由同 parseSpecValues:模板坏掉不该让详情页打不开。
		//
		// **但形状不符要能看出来**:若服务端某天把模板写成对象而不是数组,
		// 这里会持续报错而接口照常返回 200 —— 前端表现为"规格编辑器空了"。
		// 日志里的这条是唯一线索,故把原始长度与错误一起打出来。
		logx.Errorf("product: spec_template 不是合法 JSON 数组(len=%d): %v", len(s), err)
		return []types.SpecTemplateItem{}
	}
	if items == nil {
		// `null` 归一到空切片,与空串分支一致。
		return []types.SpecTemplateItem{}
	}
	return items
}

// jsonText 请求方向:types 的 interface{}(对象/数组)→ proto 要的 JSON 文本。
//
// ============================================================
// 空值一律返回空串,而不是 "{}"
// ============================================================
//
// 与 product-service 的 jsonToText / jsonMapToText 同一口径
// (它的注释:"空值统一为空串 —— proto 侧用 "" 表示 SQL NULL")。
// 若把空对象序列化成 "{}" 发下去,服务端 textToJSON("{}") 会把 DB 的
// NULL 写成 '{}',读回来时前端拿到的从"没有规格模板"变成"空对象" ——
// 形状变了,而且这种差异只会在端到端对比时才发现。
//
// 失败返回 error 而不是像响应方向那样静默降级:这是**请求方向**,
// 静默丢一个字段等于"用户以为改了、实际没改"。(实际上经 httpx.Parse
// 解析出来的值一定是 JSON 可表达的,这里只是不埋 panic 隐患。)
func jsonText(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("JSON 字段序列化失败: %w", err)
	}
	switch string(b) {
	case "null", "{}", "[]":
		// 三种都表示"客户端没给内容",统一成空串。见上。
		return "", nil
	}
	return string(b), nil
}

// ============================================================
// ③ 管理端 / 用户端的 SKU 映射**必须分开写**
// ============================================================
//
// proto 只有一个 Sku message(注释:"管理端与用户端共用,用户端不展示的
// 字段留空"),而 types 有两个类型:**AdminSku 有 cost_price / lock_stock,
// UserSku 没有**。
//
// 不能"复用一个再把多的字段清零":
//
//	复用 AdminSku 再删字段 → 改不动(Go 的 struct 不能删字段)
//	复用 UserSku 再补字段 → 补不回来(UserSku 里的 CostPrice 是零值,
//	                        而管理端要的是真实成本价)
//
// 所以这里是**故意**的两份相似代码。判据是"返回类型由 .api 决定":
// 一个 logic 的返回类型对不上签名就编译不过,合并成一个公共类型做不到。

func toAdminSku(p *v1_productv1.Sku) types.AdminSku {
	if p == nil {
		return types.AdminSku{}
	}
	return types.AdminSku{
		SkuId:      p.GetSkuId(),
		SpuId:      p.GetSpuId(),
		SkuName:    p.GetSkuName(),
		SpecValues: parseSpecValues(p.GetSpecValues()),
		Price:      p.GetPrice(),
		CostPrice:  p.GetCostPrice(),
		Stock:      p.GetStock(),
		LockStock:  p.GetLockStock(),
		SoldCount:  p.GetSoldCount(),
		SkuCode:    p.GetSkuCode(),
		SkuImage:   p.GetSkuImage(),
		SkuStatus:  p.GetSkuStatus(),
	}
}

func toAdminSkus(ps []*v1_productv1.Sku) []types.AdminSku {
	// make(..., 0, len) 而不是 var x []T:后者无数据时是 nil,
	// 序列化成 null 而非 [],而前端有 list.length / .map() 这类写法。
	out := make([]types.AdminSku, 0, len(ps))
	for _, p := range ps {
		out = append(out, toAdminSku(p))
	}
	return out
}

func toUserSku(p *v1_productv1.Sku) types.UserSku {
	if p == nil {
		return types.UserSku{}
	}
	return types.UserSku{
		SkuId:      p.GetSkuId(),
		SpuId:      p.GetSpuId(),
		SkuName:    p.GetSkuName(),
		SpecValues: parseSpecValues(p.GetSpecValues()),
		Price:      p.GetPrice(),
		Stock:      p.GetStock(),
		SoldCount:  p.GetSoldCount(),
		SkuCode:    p.GetSkuCode(),
		SkuImage:   p.GetSkuImage(),
		SkuStatus:  p.GetSkuStatus(),
		// **刻意不取 CostPrice / LockStock** —— 用户端不展示内部字段,
		// 这不是"省事",是契约:成本价泄露给前台等于把毛利公开。
	}
}

func toUserSkus(ps []*v1_productv1.Sku) []types.UserSku {
	out := make([]types.UserSku, 0, len(ps))
	for _, p := range ps {
		out = append(out, toUserSku(p))
	}
	return out
}

func toProductImage(p *v1_productv1.SpuImage) types.ProductImage {
	if p == nil {
		return types.ProductImage{}
	}
	return types.ProductImage{
		ImageId:   p.GetImageId(),
		ImageUrl:  p.GetImageUrl(),
		SortOrder: p.GetSortOrder(),
		IsMain:    p.GetIsMain(),
	}
}

func toProductImages(ps []*v1_productv1.SpuImage) []types.ProductImage {
	out := make([]types.ProductImage, 0, len(ps))
	for _, p := range ps {
		out = append(out, toProductImage(p))
	}
	return out
}

// ============================================================
// ① items → list(列表项映射)
// ============================================================
//
// 时间字段用 converter.Timestamp 而不是在本包再抄一份 formatTimestamp:
// user / order / payment 三个域已各有一份,第四份只会让"时间格式口径"
// 散在四处(而它必须是 RFC3339,见 converter 的说明)。

func toAdminSpuListItem(p *v1_productv1.SpuListItem) types.AdminSpuListItem {
	if p == nil {
		return types.AdminSpuListItem{}
	}
	return types.AdminSpuListItem{
		SpuId:        p.GetSpuId(),
		SpuName:      p.GetSpuName(),
		CategoryId:   p.GetCategoryId(),
		CategoryName: p.GetCategoryName(),
		Brand:        p.GetBrand(),
		MainImage:    p.GetMainImage(),
		SpuStatus:    p.GetSpuStatus(),
		Priority:     p.GetPriority(),
		CreatedAt:    converter.Timestamp(p.GetCreatedAt()),
		UpdatedAt:    converter.Timestamp(p.GetUpdatedAt()),
		// min_price / max_price 是 int64:proto 注释写明"DB 侧按整数分位返回",
		// 前端也是 number,故**不做任何换算**(不乘 100、不转 float)。
		MinPrice:   p.GetMinPrice(),
		MaxPrice:   p.GetMaxPrice(),
		TotalStock: p.GetTotalStock(),
		TotalSold:  p.GetTotalSold(),
	}
}

func toAdminSpuListItems(ps []*v1_productv1.SpuListItem) []types.AdminSpuListItem {
	out := make([]types.AdminSpuListItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, toAdminSpuListItem(p))
	}
	return out
}

func toUserSpuListItem(p *v1_productv1.UserSpuListItem) types.UserSpuListItem {
	if p == nil {
		return types.UserSpuListItem{}
	}
	return types.UserSpuListItem{
		SpuId:        p.GetSpuId(),
		SpuName:      p.GetSpuName(),
		CategoryName: p.GetCategoryName(),
		Brand:        p.GetBrand(),
		MainImage:    p.GetMainImage(),
		MinPrice:     p.GetMinPrice(),
		MaxPrice:     p.GetMaxPrice(),
		TotalSold:    p.GetTotalSold(),
		Stock:        p.GetStock(),
	}
}

func toUserSpuListItems(ps []*v1_productv1.UserSpuListItem) []types.UserSpuListItem {
	out := make([]types.UserSpuListItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, toUserSpuListItem(p))
	}
	return out
}

// ============================================================
// 详情:同一个 ProductDetail → 两个 types 类型
// ============================================================

// toAdminProductResp 管理端详情(含成本价 / 锁库存)。
func toAdminProductResp(p *v1_productv1.ProductDetail) *types.AdminProductResp {
	if p == nil {
		// "成功但没给 product"是服务端契约违约(如它忘了填 product 字段)。
		// 这里给空壳而不是返回 nil:nil 会序列化成 "data": null,
		// 前端解 spu_id 直接抛错;空壳至少结构完整、能看出"没查到"。
		// 同 cart 域 toCartItem(nil) 的取舍。
		return &types.AdminProductResp{
			SpecTemplate: parseSpecTemplate(""),
			SkuList:      []types.AdminSku{},
			ImageList:    []types.ProductImage{},
		}
	}
	return &types.AdminProductResp{
		SpuId:        p.GetSpuId(),
		SpuName:      p.GetSpuName(),
		CategoryId:   p.GetCategoryId(),
		CategoryName: p.GetCategoryName(),
		Brand:        p.GetBrand(),
		Description:  p.GetDescription(),
		MainImage:    p.GetMainImage(),
		SpecTemplate: parseSpecTemplate(p.GetSpecTemplate()),
		SpuStatus:    p.GetSpuStatus(),
		Priority:     p.GetPriority(),
		SkuList:      toAdminSkus(p.GetSkuList()),
		ImageList:    toProductImages(p.GetImageList()),
		CreatedAt:    converter.Timestamp(p.GetCreatedAt()),
		UpdatedAt:    converter.Timestamp(p.GetUpdatedAt()),
	}
}

// toUserProductResp 用户端详情。
//
// 与管理端的差别只在 sku_list 的元素类型(AdminSku vs UserSku)——
// spec_template 用户端也要(选购时要按规格选),故字段集其余部分相同。
func toUserProductResp(p *v1_productv1.ProductDetail) *types.UserProductResp {
	if p == nil {
		return &types.UserProductResp{
			SpecTemplate: parseSpecTemplate(""),
			SkuList:      []types.UserSku{},
			ImageList:    []types.ProductImage{},
		}
	}
	return &types.UserProductResp{
		SpuId:        p.GetSpuId(),
		SpuName:      p.GetSpuName(),
		CategoryId:   p.GetCategoryId(),
		CategoryName: p.GetCategoryName(),
		Brand:        p.GetBrand(),
		Description:  p.GetDescription(),
		MainImage:    p.GetMainImage(),
		SpecTemplate: parseSpecTemplate(p.GetSpecTemplate()),
		SpuStatus:    p.GetSpuStatus(),
		Priority:     p.GetPriority(),
		SkuList:      toUserSkus(p.GetSkuList()),
		ImageList:    toProductImages(p.GetImageList()),
		CreatedAt:    converter.Timestamp(p.GetCreatedAt()),
		UpdatedAt:    converter.Timestamp(p.GetUpdatedAt()),
	}
}

// ============================================================
// 请求:types → proto
// ============================================================

// optionalInt64 把"HTTP 侧的零值"翻译成"proto 侧的未设置"。
//
// ============================================================
// 为什么要它:proto 用 wrappers 区分"没传"与"传了 0"
// ============================================================
//
// GetProductListReq.category_id 是 google.protobuf.Int64Value:
//
//	nil  → 不过滤
//	&0   → 明确要筛 category_id = 0
//
// 而 .api 里 category_id 是**普通 int64**(form tag),不传即 0 ——
// 这一层信息在 HTTP 侧已经丢了(与库存流水域的 sku_id 同一情况)。
// 故这里按"零值即不过滤"处理:0 → nil。
//
// **代价**:无法按 category_id = 0 筛选。但类目主键自增从 1 开始,
// 0 不是合法类目,故没有实际影响(库存域 .api 的注释也是这个结论)。
//
// 同样用于 UpdateProductFullReq 的 category_id / priority:服务端把
// nil 当"该字段不改",若不包一层而传 &0,全量更新会把类目/权重清零。
func optionalInt64(v int64) *wrapperspb.Int64Value {
	if v == 0 {
		return nil
	}
	return wrapperspb.Int64(v)
}

// ============================================================
// ④ 局部更新:非零值字段 → updates_json
// ============================================================
//
// proto 的 UpdateProductReq 只有 {spu_id, updates_json} —— 服务端把
// JSON 解成 map,再用 mapstructure 按 **json tag** 合并到 SysProductSpu
// 实体上(与类目域同一套;product-service 的 utils.DecodeUpdateJSON 两个
// 域共用)。
//
// 单体时代 HTTP 层是把请求体**原样**解成 map[string]interface{} 转发,
// 键名完全由前端决定。新的 .api 改成了类型化结构体,于是 BFF 必须自己
// 拼这个 map。用 struct + omitempty 而不是手写 map:
//
//	键名来自 json tag,不可能打错(map 的字符串键没有任何检查)
//	omitempty 精确表达"零值即不改",与 .api 里字段全 optional 的意图一致
//	多一个字段时编译器会提醒(结构体字面量要补全),map 不会
type productUpdates struct {
	SpuName     string `json:"spu_name,omitempty"`
	CategoryId  int64  `json:"category_id,omitempty"`
	Brand       string `json:"brand,omitempty"`
	Description string `json:"description,omitempty"`
	MainImage   string `json:"main_image,omitempty"`
	SpuStatus   string `json:"spu_status,omitempty"`
	Priority    int64  `json:"priority,omitempty"`
}

// empty 判断这次请求有没有任何要改的字段。
func (u productUpdates) empty() bool {
	return u == productUpdates{}
}

// productUpdatesFrom 把 HTTP 请求里**非零值**的字段挑出来。
//
// ============================================================
// 取舍:"零值即不改" 让一部分取值改不了
// ============================================================
//
// 类型化结构体**无法区分**"没传"与"传了零值"(optional 在 goctl 里
// 只影响绑定,不产生指针)。故:
//
//	brand = ""        改不成空(想清空品牌做不到)
//	priority = 0      改不成 0(想把它排到最后做不到)
//	spu_status = ""   同上(但它本来也不该被"清空" —— 见下)
//
// 要用指针(*string / *int64)才能两者兼顾,那需要 .api 对应字段改成
// 指针类型。当前 .api 全是值类型,故按零值处理。
//
// 本请求**没有布尔字段**(唯一的 bool 是 proto Spu.is_deleted,属软删
// 标记,不该从"改商品"入口暴露)。若将来加了 bool,必须改用 *bool:
// 非指针的 false 与"没传"无法区分,一次只想改名称的请求会顺手把那个
// 开关关掉(menu 域的 is_visible / is_cache 就是为此改成指针的)。
//
// ============================================================
// spu_status 的值域由服务端把关
// ============================================================
//
// 服务端只接受 withdrawn → draft 这一种转换,其余(含 draft → published)
// 一律回 error_msg。发布/下架要走专门的 publish / withdraw 接口 ——
// 那两条有自己的前置校验(类目、SKU、库存),从"改商品"这里绕过去
// 会让校验被跳过。BFF 不预检,让服务端做唯一的判定方。
func productUpdatesFrom(req *types.UpdateProductReq) productUpdates {
	return productUpdates{
		SpuName:     req.SpuName,
		CategoryId:  req.CategoryId,
		Brand:       req.Brand,
		Description: req.Description,
		MainImage:   req.MainImage,
		SpuStatus:   req.SpuStatus,
		Priority:    req.Priority,
	}
}

// ============================================================
// 创建路径的两个状态值
// ============================================================
//
// 前端的建品请求体里有这两个值(create/index.vue 的 payload:
// `spu_status: "published"`,每个 SKU `sku_status: 'active'`),
// .api 里也已声明它们(CreateProductReq.SpuStatus / CreateSkuItem.SkuStatus),
// 所以**不再是"字段被静默丢弃"的问题** —— 现在是两个刻意的策略选择。
//
// ------------------------------------------------------------
// sku_status:透传,缺省兜 active
// ------------------------------------------------------------
//
//	active   —— 与 DB 默认值、前端取值都一致
//	inactive —— 会让商品永远上不了架(上架校验要求至少一个 active SKU)
//
// 前端建品时固定传 active,故透传即可;只有在不传时才兜 active。
//
// ------------------------------------------------------------
// spu_status:**一律强制 draft**,不透传前端的 published
// ------------------------------------------------------------
//
// 前端固定传 "published"(创建即上架)。**BFF 刻意覆盖它**,理由:
//
//	product-service 的上架路径(PublishProduct)会做库存/价格校验,
//	而"创建时就写 published"**完全绕过那次校验** —— 单体正是如此,
//	结果是能建出没库存、没价格却已上架的商品。
//
// 强制 draft 之后,运营必须点一次上架,而那一次会走完整校验。
//
// **代价(前端可见行为变了)**:新建商品以草稿出现,而不是直接上架。
// 这是有意的行为变更,不是兼容性疏漏 —— 已与项目负责人确认。
//
// 注意空串在这里**不是**一个可接受的取值:DB 有
// `ck_spu_status CHECK (spu_status IN ('draft','published','withdrawn'))`,
// 空串不满足 → INSERT 直接失败 → 回 503。所以必须给出一个合法值。
// 又及:GORM 不会替我们兜 —— 模型没写 `default` 标签的字段,零值会被
// **显式写进** INSERT,DB 的 DEFAULT 'draft' 不会生效。
const (
	spuStatusDraft  = "draft"
	skuStatusActive = "active"
)

// skuStatusOrDefault 透传前端的 sku_status,空串时兜 active。
//
// 单独抽出来是为了让"这里有一次兜底"在调用处可见 —— 直接写
// `SkuStatus: it.SkuStatus` 会漏掉 DB 的 CHECK 约束,表现为建品 503。
func skuStatusOrDefault(s string) string {
	if s == "" {
		return skuStatusActive
	}
	return s
}

// forceDraftSpuStatus 建品时**一律**用 draft,忽略请求里带来的值。
//
// 刻意不写成 `req.SpuStatus`(透传)也不写成"空串才兜":前端固定传
// published,透传就等于沿用单体"创建即上架、跳过上架校验"的老行为。
// 详见上方长注释 —— 这是有意的行为变更,已与项目负责人确认。
//
// 抽成函数是为了让这次"覆盖"在一个地方可见,而不是散在字段赋值里
// 看着像随手写死。
func forceDraftSpuStatus(requested string) string {
	_ = requested // 显式忽略:保留参数让调用处一眼看出"这里收了一个值但不用"
	return spuStatusDraft
}

// toProtoSkuFromCreate 建品时的 SKU(types.CreateSkuItem → proto Sku)。
//
// types.CreateSkuItem **没有** sku_id —— 新建的 SKU 由服务端生成主键,
// 故这里不设(proto 的零值 0 就是"新增",全量更新也靠它区分增改)。
func toProtoSkuFromCreate(it types.CreateSkuItem) (*v1_productv1.Sku, error) {
	specValues, err := jsonText(it.SpecValues)
	if err != nil {
		return nil, fmt.Errorf("sku_list[].spec_values %w", err)
	}
	return &v1_productv1.Sku{
		SkuName:    it.SkuName,
		SpecValues: specValues,
		Price:      it.Price,
		CostPrice:  it.CostPrice,
		Stock:      it.Stock,
		SkuCode:    it.SkuCode,
		SkuImage:   it.SkuImage,
		// sku_status 透传,只有前端没传时才兜 active。
		//
		// 不能直接发空串:DB 的 ck_sku_status 约束只接受 active/inactive,
		// 空串会让整次建品失败(而且表现为 503,像依赖挂了)。
		SkuStatus: skuStatusOrDefault(it.SkuStatus),

		// SpuId / SoldCount / LockStock 不设:归属由服务端按新建的 spu_id
		// 回填,销量与锁定库存属于运行时数据,建品时永远是 0。
	}, nil
}

// toProtoSkuFromAdminSku 全量更新时的 SKU(types.AdminSku → proto Sku)。
//
// 与创建路径的差别只有一处但是关键:这里**必须原样带上 sku_id**。
// 服务端按它区分"更新"与"新增":
//
//	sku_id > 0 → 更新该 SKU
//	sku_id = 0 → 插入新 SKU
//
// 不带 sku_id 的后果不是"少更新一个字段",而是**多插一行重复 SKU**
// (并且可能撞上规格组合唯一性校验)。
//
// sku_status 这里**不兜默认值**(与创建路径相反):AdminSku 有这个字段,
// 空值只可能是客户端显式传了空。兜成 active 等于把"客户端漏传"变成
// "悄悄把这个 SKU 启用",而它原本可能是被禁用的。
func toProtoSkuFromAdminSku(s types.AdminSku) (*v1_productv1.Sku, error) {
	specValues, err := jsonText(s.SpecValues)
	if err != nil {
		return nil, fmt.Errorf("sku_list[].spec_values %w", err)
	}
	return &v1_productv1.Sku{
		SkuId:      s.SkuId,
		SkuName:    s.SkuName,
		SpecValues: specValues,
		Price:      s.Price,
		CostPrice:  s.CostPrice,
		Stock:      s.Stock,
		SkuCode:    s.SkuCode,
		SkuImage:   s.SkuImage,
		SkuStatus:  s.SkuStatus,
		// SpuId 不取自请求:服务端会用自己的 spu_id 覆盖它(sku.SpuId = in.SpuId),
		// 传了也没用,反而让人以为客户端能决定归属。
	}, nil
}

func toProtoImages(imgs []types.ProductImage) []*v1_productv1.SpuImage {
	// 用 make(..., 0, len):nil 切片在 proto 里与"空"不可区分,
	// 而服务端是按 len > 0 判断"要不要处理图片列表"的,故显式给空切片。
	out := make([]*v1_productv1.SpuImage, 0, len(imgs))
	for _, img := range imgs {
		out = append(out, &v1_productv1.SpuImage{
			ImageId:   img.ImageId,
			ImageUrl:  img.ImageUrl,
			SortOrder: img.SortOrder,
			IsMain:    img.IsMain,
		})
	}
	return out
}

// toProtoCreateSpu 把**扁平**的 HTTP 请求组装成**嵌套**的 proto Spu。
//
// ============================================================
// 形状差一层:HTTP 扁平,proto 用 Spu 包住全部字段
// ============================================================
//
// proto:CreateProductReq { Spu spu = 1 } —— SPU、SKU 列表、图片列表
// 全在 spu 里面,由服务端同一事务创建(proto 注释:"创建商品,同时创建
// SKU 列表与图片列表(同一事务)")。
//
// HTTP:types.CreateProductReq 是扁平的(spu_name / sku_list / image_list
// 平铺),因为前端表单本来就是一个对象。故这一步是**必然的**组装,
// 不是可有可无的包装。
//
// SpuId / SpuStatus / IsDeleted 不取自请求:
//
//	SpuId     新建,由 DB 自增
//	SpuStatus 见 spuStatusDraft 的说明(当前只能由 BFF 兜默认值)
//	IsDeleted 软删标记,建品时必为 false
func toProtoCreateSpu(req *types.CreateProductReq) (*v1_productv1.Spu, error) {
	specTemplate, err := jsonText(req.SpecTemplate)
	if err != nil {
		return nil, fmt.Errorf("spec_template %w", err)
	}

	skuList := make([]*v1_productv1.Sku, 0, len(req.SkuList))
	for _, it := range req.SkuList {
		sku, err := toProtoSkuFromCreate(it)
		if err != nil {
			return nil, err
		}
		skuList = append(skuList, sku)
	}

	return &v1_productv1.Spu{
		SpuName:      req.SpuName,
		CategoryId:   req.CategoryId,
		Brand:        req.Brand,
		Description:  req.Description,
		MainImage:    req.MainImage,
		SpecTemplate: specTemplate,
		// 收下 req.SpuStatus 但**不采用** —— 见 forceDraftSpuStatus。
		// 这行注释也把结构体字面量的对齐切成两组:gofmt 对"被注释分隔的
		// 字段"要求各自成组对齐,写成一组会让 gofmt -l 报这个文件。
		SpuStatus: forceDraftSpuStatus(req.SpuStatus),

		Priority:  req.Priority,
		SkuList:   skuList,
		ImageList: toProtoImages(req.ImageList),
	}, nil
}

// ============================================================
// 错误哨兵:局部更新一个字段都没传
// ============================================================
//
// 回 400,与 role / user / menu 三个域一致 —— 但**与单体有一处刻意的偏离**。
//
// 单体把请求体原样当 map 转发,空 body 会变成 "{}",服务端解出空 map、
// 合并零个字段、返回 200 —— 也就是"什么都没发生的成功"。
//
// BFF 改成 400,理由是**类型化结构体把一种错误变成了静默**:
//
//	单体:put { "spu_nmae": "x" }(打错字)→ 服务端按 json tag 合并不中,
//	      但那是"服务端看得见一个它不认识的键";若真加了字段就能生效
//
// go-zero 的 httpx.Parse **静默丢弃未声明的字段** → 打错字 = 一个字段
// 都没传 = 空更新 → 回 200 且前端以为改成功了
//
// 静默成功会掩盖前端 bug(用户反馈"改了没生效",查不出原因)。
// 报 400「请求参数错误」至少让它当场可见。
//
// 代价:前端若真的发过空 body 的更新,行为从 200 变成 400。
// 端到端验证时留意一次(空 body 的局部更新在页面上几乎不可能出现 ——
// 表单总是带全字段)。
var errNothingToUpdate = errors.New("请求参数错误")
