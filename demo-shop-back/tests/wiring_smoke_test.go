package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"demo-shop-back/src/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/routes"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// 安全网:HTTP 接线冒烟测试
//
// 设计要点:**不硬编码路由路径**,而是用 engine.Routes() 枚举真实注册的路由。
// 这样测试不会因为"路径写错"而假红 —— 它只对真实的接线问题报警。
//
// 判据:
//   - 任一被探测端点返回 **5xx** → 接线断裂 / nil panic / 注入缺失(B0 改坏了东西)
//   - **4xx 允许**:测试库无种子数据、无权限、缺参数都属正常业务响应
//
// B0 重构期间此文件必须始终保持绿色。
// ============================================================

var initOnce struct {
	sync.Once
	err error
}

// testPrivateKey 测试内生成的 RS256 私钥,用于直接签发测试 token。
var testPrivateKey *rsa.PrivateKey

// initTestGlobals 显式初始化全局依赖(幂等)。
// 为什么不调 db.InitDB():它读 config.GlobalConfig,而测试从不 LoadConfig,
// 会拿到空 DSN。这里用与 TestMain 相同的环境变量显式构造,保证自洽。
func initTestGlobals() error {
	initOnce.Do(func() {
		cfg := config.DatabaseConfig{
			Host:     getenv("TEST_PG_HOST", "localhost"),
			Port:     getenv("TEST_PG_PORT", "5432"),
			User:     getenv("TEST_PG_USER", "postgres"),
			Password: getenv("TEST_PG_PASSWORD", "postgres"),
			Dbname:   "demo_shop_test",
			Sslmode:  "disable",
		}
		if err := db.InitDBWith(cfg); err != nil {
			initOnce.err = err
			return
		}
		// JWT 验签器需要公钥;测试内现生成一对,免去对 jwt_keys/ 磁盘文件的依赖。
		// 私钥留作后续直接签发测试 token 用。
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			initOnce.err = err
			return
		}
		testPrivateKey = priv
		middleware.InitJWTWithKey(&priv.PublicKey)
	})
	return initOnce.err
}

// insertSmokeUser 直插一个已知口令的用户(哈希方式与 UserService 注册路径一致)。
// 不走注册接口:该接口当前稳定返回 500(DS-A-24 §6.5),冒烟测试不该被它阻塞。
func insertSmokeUser(t *testing.T, username, plainPassword string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("生成口令哈希失败: %v", err)
	}
	if err := db.DB.Exec(`
		INSERT INTO sys_user (username, password_hash, status, failed_attempts)
		VALUES (?, ?, 'active', 0)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		username, string(hash)).Error; err != nil {
		t.Fatalf("插入冒烟用户失败: %v", err)
	}
}

// smokeUserID 查冒烟用户的 user_id,供签发测试 token 使用。
func smokeUserID(t *testing.T, username string) int64 {
	t.Helper()
	var id int64
	if err := db.DB.Raw(`SELECT user_id FROM sys_user WHERE username = ?`, username).
		Scan(&id).Error; err != nil {
		t.Fatalf("查询冒烟用户ID失败: %v", err)
	}
	if id == 0 {
		t.Fatalf("冒烟用户不存在: %s", username)
	}
	return id
}

// 测试内直接签发 access token,不走登录接口
func issueTestToken(t *testing.T, priv *rsa.PrivateKey, userID int64, username string) string {
	claims := utils.CustomClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("签发测试 token 失败: %v", err)
	}
	return signed
}

// TestWiring_HTTP_AllRegisteredRoutes 枚举全部已注册路由并逐个探测。
//
// 安全性:只探测 **GET 且无路径参数** 的路由 —— 这类是只读查询,
// 不会创建订单/扣库存/发消息,可以放心打。写操作与带参路由不在此测试范围。
//
// 2026-10 调整(阶段 C2):判据从"5xx 即接线断裂"放宽为"5xx 需能归因"。
// 起因:user-service / product-service 已拆出进程,本测试环境不启动它们,
// 于是 /api/v1/user/info、/api/v1/products 会返回
//
//	{"code":503,"message":"user-service 不可用"}
//
// 这是**正确的降级响应**,不是接线问题。原来的"任何 5xx 都报错"会产生假红,
// 而假红比不测更糟 —— 它会训练人忽略这个测试。
//
// 现在的判据:
//   - 4xx:允许(无种子数据、无权限、缺参数);
//   - 503 且 message 指向某个服务不可用:允许,计入"依赖缺失"而非"接线问题";
//   - 其余 5xx(含 500):判为接线断裂 / nil panic / 注入缺失。
func TestWiring_HTTP_AllRegisteredRoutes(t *testing.T) {
	if err := initTestGlobals(); err != nil {
		t.Skipf("全局初始化失败,跳过: %v", err)
	}

	const uname = "smoke-wiring-user"
	insertSmokeUser(t, uname, "Smoke-Passw0rd-1")

	engine := routes.InitRoutes(testDeps())

	// 直接签发 token:本测试只验证路由接线,不依赖登录接口与 user-service
	userID := smokeUserID(t, uname)
	token := issueTestToken(t, testPrivateKey, userID, uname)

	checked := 0
	skippedRPC := 0
	for _, r := range engine.Routes() {
		if r.Method != http.MethodGet || strings.Contains(r.Path, ":") {
			continue
		}
		// /uploads/* 是静态文件服务,不属于接线验证范围
		if strings.HasPrefix(r.Path, "/uploads") {
			continue
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, r.Path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		engine.ServeHTTP(w, req)
		checked++

		if w.Code >= 500 {
			if unavailable := downstreamUnavailable(w); unavailable != "" {
				skippedRPC++
				t.Logf("[%s] %s → %d 跳过(下游 %s 未启动)", r.Handler, r.Path, w.Code, unavailable)
				continue
			}
			t.Errorf("[%s] %s → %d  **疑似接线断裂 / nil panic / 注入缺失**  body=%s",
				r.Handler, r.Path, w.Code, trunc(w.Body.String()))
			continue
		}
		t.Logf("[%s] %s → %d", r.Handler, r.Path, w.Code)
	}

	if checked == 0 {
		t.Fatal("没有枚举到任何可探测的 GET 路由 —— 路由注册本身可能坏了")
	}
	t.Logf("共探测 %d 条只读路由,其中 %d 条因下游服务未启动而跳过", checked, skippedRPC)
}

// downstreamUnavailable 识别"某个下游服务没连上"这一类 5xx。
//
// 契约:调用方在客户端为 nil(RPC 未建连)时返回 `xxx-service 不可用`。
// 接受两种状态码:
//   - 503:product-service 侧已统一走 utils.Unavailable(标准语义);
//   - 500:user-service 侧的历史写法仍走 utils.Error,尚未统一。
//
// 返回非空字符串表示属于该情形,值为服务名;返回空串表示这不是依赖缺失,需要人来查。
//
// 为什么不只看状态码:接线错误(handler 注入成 nil 指针、路由注册漏了)也会返回 500,
// 必须靠"响应体是否是那句约定的降级文案"来区分,否则会把真问题一起放过。
func downstreamUnavailable(w *httptest.ResponseRecorder) string {
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		return ""
	}
	if body.Code != http.StatusServiceUnavailable && body.Code != http.StatusInternalServerError {
		return ""
	}
	const suffix = " 不可用"
	if !strings.HasSuffix(body.Message, suffix) {
		return ""
	}
	return strings.TrimSuffix(body.Message, suffix)
}

// TestWiring_HTTP_AuthedRouteNeedsToken 反向验证鉴权中间件接线:
// 不带 token 打已鉴权端点应被拒(4xx),而不是 200 或 5xx。
func TestWiring_HTTP_AuthedRouteNeedsToken(t *testing.T) {
	if err := initTestGlobals(); err != nil {
		t.Skipf("全局初始化失败: %v", err)
	}
	engine := routes.InitRoutes(testDeps())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/cart", nil)
	engine.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("无 token 访问已鉴权端点竟然 200 —— 鉴权中间件未生效")
	}
	if w.Code >= 500 {
		t.Fatalf("无 token 访问返回 %d, 期望 4xx —— 鉴权链路接线异常", w.Code)
	}
	t.Logf("无 token → %d(符合预期)", w.Code)
}

func trunc(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
