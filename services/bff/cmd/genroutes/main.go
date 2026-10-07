// Command genroutes 把权限码固化进 BFF 的路由注册代码。
//
// 它在 `goctl api go` **之后**运行,是构建流水线的一环,不是产品代码:
//
//	goctl api go -api bff.api -dir . --style gozero --home ../../.goctl
//	go run ./cmd/genroutes
//
// ============================================================
// 它解决什么问题
// ============================================================
//
// 判权需要"这个接口要求哪些权限码",而那个信息在 user_db 的
// sys_permission 里。goctl 只认 .api,它生成的 routes.go 里没有权限码,
// 而权限码也不该进 .api(route 级 @server 注解 goctl 1.9.2 不支持,
// 已实测)。所以"把权限码写进代码"这一步必然落在 goctl 之外 —— 就是本命令。
//
// 产物是**编译期常量**:运行时不需要任何"接口 → 权限码"的反查,
// 也就不需要为那条链准备缓存与失效机制。
//
// ============================================================
// 幂等
// ============================================================
//
// 每次运行都先剥掉旧的包装(guard.RouteTemplate / guard.Permission),
// 再从裸处理器重新包。所以重复跑不会套两层,权限码变更也能正确刷新。
//
// ============================================================
// fail-closed:查不到权限码就**拒绝生成**
// ============================================================
//
// 管理端路由在 sys_permission 里查不到码,只有两种可能:
// 配权限点时漏了,或者这条路由本来就不该判权(例外清单)。前者必须拦住,
// 否则那条路由会没有任何保护地上线。故除例外清单外一律报错退出。
//
// ============================================================
// 运行前提
// ============================================================
//
// 需要 etcd 与 user-service 在运行 —— 权限码是经 RPC 取的(见
// docs 里"接口→权限码在构建期固化"的决策)。CI 里要先拉起这两个。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/config"
	"demo-shop/services/bff/internal/infra/rpc"

	"github.com/zeromicro/go-zero/core/conf"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	routesFileRel = "internal/handler/routes.go"
	configRel     = "etc/bff.yaml"
)

var (
	rePrefixLine = regexp.MustCompile(`WithPrefix\("([^"]*)"\)`)
	reMethodLine = regexp.MustCompile(`Method:\s+http\.Method(\w+),`)
	rePathLine   = regexp.MustCompile(`Path:\s+"([^"]*)",`)
	reHandler    = regexp.MustCompile(`Handler:\s+(.+),$`)
)

// routeBlock 生成物里的一个 rest.Route 字面量。
type routeBlock struct {
	indent   string // Handler 行前的缩进,改写时原样保留
	method   string // 来自 http.MethodXxx,已转成大写
	local    string // Path 字段的原始值,如 "/:id"
	handler  string // Handler 字段的表达式,可能带 guard 包装
	lineFrom int    // 该块起始行(1-based,含)
	lineTo   int    // 该块结束行(1-based,含)
}

// prefixBlock AddRoutes 语句里的前缀声明。
type prefixBlock struct {
	line int
	path string
}

func main() {
	routesPath := flag.String("routes", routesFileRel, "要改写的 routes.go(相对服务根)")
	configPath := flag.String("config", configRel, "BFF 配置(相对服务根,提供 etcd 与下游 key)")
	dryRun := flag.Bool("dry-run", false, "只打印将要做的改动,不落盘")
	flag.Parse()

	if err := run(*routesPath, *configPath, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		os.Exit(1)
	}
}

func run(routesPath, configPath string, dryRun bool) error {
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("读不到 BFF 配置 %s(工作目录应为 services/bff/): %w", configPath, err)
	}
	var c config.Config
	conf.MustLoad(configPath, &c)

	raw, err := os.ReadFile(routesPath)
	if err != nil {
		return fmt.Errorf("读不到 %s(先跑 goctl api go 生成): %w", routesPath, err)
	}
	lines := strings.Split(string(raw), "\n")

	blocks, prefixes := parseRoutes(lines)
	if len(blocks) == 0 {
		return fmt.Errorf("在 %s 里没解析出任何路由 —— 生成物格式变了?", routesPath)
	}
	fmt.Printf("解析出 %d 条路由,%d 个前缀块\n", len(blocks), len(prefixes))

	// 把每条路由补全成完整模板,并找出需要处理的
	var targets []target
	for i := range blocks {
		full := blocks[i].local
		if p, ok := prefixFor(len(prefixes), prefixes, blocks, i); ok {
			if blocks[i].local == "/" {
				full = p
			} else {
				full = p + blocks[i].local
			}
		}
		if !strings.HasPrefix(full, "/api/v1/admin") {
			continue
		}
		targets = append(targets, target{i, full, blocks[i].method + " " + full})
	}
	fmt.Printf("其中管理端路由 %d 条,需要判权处理\n\n", len(targets))

	// ---- 查权限码 ----
	//
	// 连之前先探一下 etcd:zrpc.MustNewClient 在 etcd 不可达时会**直接 panic**
	// 并打出几十行堆栈,而真正的原因只有一句"etcd 没起来"。写生成脚本的人
	// 需要的是那句话,不是堆栈。
	if err := pingEtcd(c.Etcd.Hosts); err != nil {
		return fmt.Errorf("连不上 etcd(%v):%w\n"+
			"权限码是经 user-service 的 RPC 取的,故生成前需要 etcd 与 "+
			"user-service(etcd key %q)都在运行",
			c.Etcd.Hosts, err, c.User.EtcdKey)
	}

	conn := rpc.Connect(c.Etcd.Hosts, c.User.EtcdKey)
	defer conn.Close()
	rbac := v1_userv1.NewRBACServiceClient(conn)

	codeOf := make(map[string][]string, len(targets))
	var missing []string
	for _, t := range targets {
		if noPermissionRoutes[t.key] {
			continue
		}
		if _, done := codeOf[t.key]; done {
			continue
		}
		resp, err := listCodes(rbac, t.full, blocks[t.idx].method)
		if err != nil {
			return fmt.Errorf("查 %s 的权限码失败(etcd/user-service 起了吗?): %w", t.key, err)
		}
		codes := findCodes(resp)
		if len(codes) == 0 {
			missing = append(missing, t.key)
			continue
		}
		codeOf[t.key] = codes
	}

	// ---- fail-closed ----
	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "以下管理端路由在 sys_permission 里查不到权限码:")
		for _, m := range missing {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		return fmt.Errorf("有 %d 条管理端路由未配置权限点;"+
			"请先在权限管理里配上,或(若确实不该判权)加入 cmd/genroutes/spec.go 的例外清单", len(missing))
	}

	// ---- 改写 ----
	// 收集行级替换:map[行号]新内容(1-based)
	replacements := make(map[int]string)
	wrapped := 0
	for _, t := range targets {
		b := &blocks[t.idx]
		base := stripGuards(b.handler)
		expr := fmt.Sprintf("guard.RouteTemplate(%q, %s)", t.full, base)
		if codes, ok := codeOf[t.key]; ok {
			expr = fmt.Sprintf("guard.Permission(%s, serverCtx.RBACRPC, %s)", goStringSlice(codes), expr)
			wrapped++
		}
		replacements[b.lineTo] = b.indent + "Handler: " + expr + ","
	}

	// ---- 保证 import 里有 guard 包 ----
	// 必须在行替换**之前**做,但不能像早先那样"先插 import 再按旧行号
	// 替换" —— 插入会让后面所有行号位移 1,替换就落到错行上
	// (症状:Handler 行被重复、Path 行被覆盖)。故这里只返回"插入位置",
	// 由下面一次遍历同时完成插入与替换,行号始终以**原始 lines** 为准。
	insertAt, importLine, err := ensureImport(lines, `"demo-shop/services/bff/internal/guard"`)
	if err != nil {
		return err
	}

	// 一次遍历:插入 import + 替换 Handler 行。行号以**原始 lines** 为准。
	final := make([]string, 0, len(lines)+1)
	for i, line := range lines {
		if insertAt == i {
			final = append(final, importLine)
		}
		if repl, ok := replacements[i+1]; ok {
			final = append(final, repl)
			continue
		}
		final = append(final, line)
	}
	// import 要插在最后一行之后的情形(理论上不会,留个兜底)
	if insertAt == len(lines) {
		final = append(final, importLine)
	}
	result := strings.Join(final, "\n")

	fmt.Printf("改写完成: %d 条路由注入模板,其中 %d 条判权\n", len(targets), wrapped)
	fmt.Printf("例外清单(明确不判权): %d 条\n", len(noPermissionRoutes))
	fmt.Printf("判权配置指纹: %s\n", configFingerprint(targets, codeOf))

	if dryRun {
		fmt.Println("\n--dry-run:未落盘")
		return nil
	}
	if err := os.WriteFile(routesPath, []byte(result), 0o644); err != nil {
		return fmt.Errorf("写回 %s 失败: %w", routesPath, err)
	}
	fmt.Printf("\n已写回 %s\n", routesPath)
	fmt.Println("下一步: go build ./... && go vet ./...  确认生成物能编译")
	return nil
}

// listCodes 调 user-service 取某接口要求的权限码。
func listCodes(rbac v1_userv1.RBACServiceClient, apiPath, method string) (*v1_userv1.ListPermCodesByApiResp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rbac.ListPermCodesByApi(ctx, &v1_userv1.ListPermCodesByApiReq{
		ApiPath:       apiPath,
		RequestMethod: method,
	})
}

// pingEtcd 探一下 etcd 是否可达。
//
// 为什么值得单独做一次:zrpc.MustNewClient 在 etcd 不可达时 panic,输出是
// 几十行堆栈 + 一个 30 秒超时,而真实原因只有"etcd 没起来"。生成脚本是
// 构建流程的一环,它的失败信息应当一句话说清要启动什么。
//
// 用很短的单次超时(3s):本地 etcd 可达与否是立即能判定的,等 30 秒
// 只是让人以为卡住了。
func pingEtcd(hosts []string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := cli.Get(ctx, "health-probe"); err != nil {
		return err
	}
	return nil
}

// stripGuards 剥掉本脚本自己之前加过的包装,拿到裸处理器表达式。
//
// 必须剥:否则重复运行会套成 Permission(Permission(RouteTemplate(...)))。
// 括号平衡扫描而不是贪婪正则 —— 处理器表达式里可能有任意层括号
// (如 xxx.NewHandler(serverCtx)),正则配不准。
func stripGuards(expr string) string {
	for {
		trimmed := strings.TrimSpace(expr)
		var inner string
		switch {
		case strings.HasPrefix(trimmed, "guard.RouteTemplate("):
			inner = extractArgs(trimmed, "guard.RouteTemplate(")
		case strings.HasPrefix(trimmed, "guard.Permission("):
			inner = extractArgs(trimmed, "guard.Permission(")
		default:
			return trimmed
		}
		if inner == "" {
			return trimmed
		}
		// 取最后一个顶层逗号之后的参数作为"下一个表达式"
		parts := splitTopLevel(inner)
		expr = parts[len(parts)-1]
	}
}

// extractArgs 取出 "prefix(a, b)" 里括号内的一整段(不依赖正则的括号匹配)。
func extractArgs(s, prefix string) string {
	body := s[len(prefix):]
	depth := 1
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
			if depth == 0 {
				return body[:i]
			}
		}
	}
	return ""
}

// splitTopLevel 按顶层逗号切分参数(忽略括号/方括号/字符串内的逗号)。
func splitTopLevel(s string) []string {
	var (
		out   []string
		depth int
		inStr bool
		cur   strings.Builder
	)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' && (i == 0 || s[i-1] != '\\'):
			inStr = !inStr
			cur.WriteByte(c)
		case inStr:
			cur.WriteByte(c)
		case c == '(' || c == '[' || c == '{':
			depth++
			cur.WriteByte(c)
		case c == ')' || c == ']' || c == '}':
			depth--
			cur.WriteByte(c)
		case c == ',' && depth == 0:
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		out = append(out, strings.TrimSpace(cur.String()))
	}
	return out
}

// parseRoutes 逐行解析生成物,抽出路由块与前缀块。
//
// 用一个**行状态机**而不是正则整块匹配,是因为 Method 与 Path 分处两行,
// 而"相邻配对"会错位(见 prefixFor 的注释:有中间件包装时缩进多一层)。
// 状态机只在"遇到 Method 行"时记录方法,到"遇到 Handler 行"时收块 ——
// 每个 rest.Route 字面量里恰好三个字段各一次。
func parseRoutes(lines []string) ([]routeBlock, []prefixBlock) {
	var (
		blocks   []routeBlock
		prefixes []prefixBlock
		cur      routeBlock
		hasMeth  bool
	)
	for i, line := range lines {
		lineNo := i + 1

		if m := rePrefixLine.FindStringSubmatch(line); m != nil {
			prefixes = append(prefixes, prefixBlock{line: lineNo, path: m[1]})
			continue
		}
		if m := reMethodLine.FindStringSubmatch(line); m != nil {
			cur = routeBlock{method: strings.ToUpper(m[1]), lineFrom: lineNo}
			hasMeth = true
			continue
		}
		if !hasMeth {
			continue
		}
		if m := rePathLine.FindStringSubmatch(line); m != nil {
			cur.local = m[1]
			continue
		}
		if m := reHandler.FindStringSubmatch(line); m != nil {
			cur.handler = strings.TrimSpace(m[1])
			cur.indent = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			cur.lineTo = lineNo
			blocks = append(blocks, cur)
			hasMeth = false
		}
	}
	return blocks, prefixes
}

// prefixFor 找出第 idx 条路由属于哪个前缀块。
//
// 生成物里每个 server.AddRoutes(...) 语句**末尾**才是 WithPrefix。
// 所以某条路由的前缀是"它之后最近的那个 WithPrefix"。
//
// 唯一没有后续 WithPrefix 的是**最后一个** AddRoutes 语句 —— 它沿用
// 上一个块的前缀(goctl 对重复前缀不会再写一次)。故此处用"取最后一个
// 已见前缀"兜底。
func prefixFor(nPrefix int, prefixes []prefixBlock, blocks []routeBlock, idx int) (string, bool) {
	for _, p := range prefixes {
		if p.line > blocks[idx].lineTo {
			return p.path, true
		}
	}
	if nPrefix > 0 {
		return prefixes[nPrefix-1].path, true
	}
	return "", false
}

// ensureImport 确认 import 块里有指定的一行。
//
// 返回"插到第几行之前"(0-based,即原 lines 的下标)与那一行的内容。
// 已有则返回 ("", 0) 表示无需插入 —— 调用方靠 importLine == "" 判断。
//
// **不直接返回改好的切片**:那会让调用方拿到一个行数与 replacements
// 的键不同步的切片,替换就会落到错行上(这个坑已经踩过一次)。
//
// 插入位置选"项目内 import 组里第一行之前":goctl 生成物的 import 块
// 是标准库一组(如 net/http)、空行、项目内一组(handler 子包与 svc)。
// 把 guard 插进第二组,go fmt 之后位置是稳定的;插进第一组则 gofmt 会
// 把它挪走,导致"跑一次脚本 + gofmt"与"只跑 gofmt"产出不同 —— 那样
// 每次生成都会多出无意义的 diff。
func ensureImport(lines []string, want string) (int, string, error) {
	if strings.Contains(strings.Join(lines, "\n"), want) {
		return 0, "", nil
	}
	inImport := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import (") {
			inImport = true
			continue
		}
		if !inImport {
			continue
		}
		if trimmed == ")" {
			break
		}
		// 项目内 import:带引号的路径里含 "demo-shop/"
		if strings.Contains(trimmed, `"demo-shop/`) {
			return i, "\t" + want, nil
		}
	}
	return 0, "", fmt.Errorf("找不到项目内 import 组,无法插入 %s", want)
}
