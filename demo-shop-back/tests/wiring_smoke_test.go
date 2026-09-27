package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/routes"

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
		// JWT 只在 main.go 初始化过;未初始化时 GetJWTService() 会 panic
		middleware.InitJWT("test-only-secret-for-smoke")
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

func smokeLogin(t *testing.T, engine http.Handler, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("冒烟用户登录失败: %d body=%s", w.Code, trunc(w.Body.String()))
	}
	var lr struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &lr); err != nil || lr.Data.AccessToken == "" {
		t.Fatalf("登录响应无 token: %s", trunc(w.Body.String()))
	}
	return lr.Data.AccessToken
}

// TestWiring_HTTP_AllRegisteredRoutes 枚举全部已注册路由并逐个探测。
//
// 安全性:只探测 **GET 且无路径参数** 的路由 —— 这类是只读查询,
// 不会创建订单/扣库存/发消息,可以放心打。写操作与带参路由不在此测试范围。
func TestWiring_HTTP_AllRegisteredRoutes(t *testing.T) {
	if err := initTestGlobals(); err != nil {
		t.Skipf("全局初始化失败,跳过: %v", err)
	}

	const uname = "smoke-wiring-user"
	insertSmokeUser(t, uname, "Smoke-Passw0rd-1")

	engine := routes.InitRoutes(testDeps())
	token := smokeLogin(t, engine, uname, "Smoke-Passw0rd-1")

	checked := 0
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
			t.Errorf("[%s] %s → %d  **疑似接线断裂 / nil panic / 注入缺失**  body=%s",
				r.Handler, r.Path, w.Code, trunc(w.Body.String()))
			continue
		}
		t.Logf("[%s] %s → %d", r.Handler, r.Path, w.Code)
	}

	if checked == 0 {
		t.Fatal("没有枚举到任何可探测的 GET 路由 —— 路由注册本身可能坏了")
	}
	t.Logf("共探测 %d 条只读路由", checked)
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
