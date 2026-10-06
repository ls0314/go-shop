package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	v1_userv1 "demo-shop/api/gen/user/v1"
)

// ============================================================
// 例外清单:管理端前缀但**明确不判权**的路由
// ============================================================
//
// 这些路由路径在 /api/v1/admin/ 下,但语义是"本人读写自己的数据",
// 故单体当年就没给它们挂权限中间件。逐条凭据:
//
//	POST /api/v1/admin/menu/tree
//	  单体 routes/menu_routes.go 里这一条没挂 permMW;
//	  services/user/migrations/000016 的注释也写明"例外公开(前端
//	  动态路由需要),未挂 PermissionMiddleware"。
//
//	/api/v1/admin/user/info/* (4 条)
//	  单体 routes/user_info_routes.go 只挂了 AuthMiddleware,
//	  没有 permMW。
//
// ============================================================
// 为什么"不判权"要**显式列出**而不是"查不到码就跳过"
// ============================================================
//
// 因为"查不到码"有两种完全不同的原因:
//
//	① 这条路由本来就不该判权(就是上面这些)   → 正确,继续
//	② 配权限点时漏了这一条                     → 事故,必须拦住
//
// 只看"查到没有"区分不了这两者。把 ① 写成清单,② 才会浮出来 ——
// 这就是 fail-closed 在生成期的落点。
//
// ============================================================
// 清单是"方法 + 完整模板"的集合,不是前缀
// ============================================================
//
// 因为同一组里可能混着两种:菜单组的 POST /tree 不判权,而
// GET /:id/tree 判权(system:menu:view)。所以只能逐条列。
//
// 这是**权宜之计**:这几条路由的正确做法是移出 admin 前缀
// (如 /api/v1/user/profile/:id)。改名会牵动前端与单体,已决定
// 搁置,故此处写死。改名之后本清单应当清空。
var noPermissionRoutes = map[string]bool{
	"POST /api/v1/admin/menu/tree":        true,
	"GET /api/v1/admin/user/info/:id":     true,
	"POST /api/v1/admin/user/info/create": true,
	"PUT /api/v1/admin/user/info/:id":     true,
	"DELETE /api/v1/admin/user/info/:id":  true,
}

// findCodes 归一化 RPC 返回的权限码。
//
// 响应是 `repeated string perm_codes`(没有 api_path 字段,无法也不能
// 再过滤)—— 服务端按 (api_path, request_method) **精确匹配**,
// 匹配不到就回空数组。所以这里只做去重与排序。
func findCodes(resp *v1_userv1.ListPermCodesByApiResp) []string {
	var out []string
	for _, raw := range resp.GetPermCodes() {
		if code := strings.TrimSpace(raw); code != "" {
			out = append(out, code)
		}
	}
	return dedupeSorted(out)
}

// dedupeSorted 去重并排序。
//
// 排序是刻意的:它让生成结果**稳定**。否则 RPC 每次返回的顺序变化
// 都会导致 routes.go 的 diff 抖动,而这个文件是要进版本控制的 ——
// 抖动会让 review 与 blame 都失去意义。
func dedupeSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// goStringSlice 把 []string 渲染成 Go 源码里的切片字面量。
func goStringSlice(codes []string) string {
	if len(codes) == 0 {
		return "nil"
	}
	parts := make([]string, 0, len(codes))
	for _, c := range codes {
		parts = append(parts, fmt.Sprintf("%q", c))
	}
	return "[]string{" + strings.Join(parts, ", ") + "}"
}

// target 一条待处理的管理端路由。
type target struct {
	idx  int    // blocks 里的下标
	full string // 含前缀的完整模板,如 /api/v1/admin/orders/:id
	key  string // "METHOD 完整模板",用于查例外清单与权限码
}

// configFingerprint 给"判权配置"算个稳定指纹:每条管理端路由的方法、
// 完整模板,以及它被固化的权限码(例外路由为空)。
//
// 为什么不直接对结果文本取哈希:结果文本**包含包装表达式**,而第二次
// 运行时输入已含包装 —— 同一个配置在第一次与第二次跑会得到不同哈希,
// 指纹就失去了"衡量配置是否变化"的意义。
//
// 用途:
//   - 两次生成指纹相同 ⇒ 权限配置没变,routes.go 的 diff 应当为空
//   - 指纹变了但你没改 .api 也没改权限点 ⇒ 有人在库里动了权限配置
func configFingerprint(targets []target, codeOf map[string][]string) string {
	lines := make([]string, 0, len(targets))
	for _, t := range targets {
		// 例外路由在 codeOf 里没有条目,这里是 nil → join 出空串,
		// 正好表示"不判权"
		lines = append(lines, t.key+" -> "+strings.Join(codeOf[t.key], ","))
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:8])
}
