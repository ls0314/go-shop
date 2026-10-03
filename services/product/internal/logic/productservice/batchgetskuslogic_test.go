package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/svc"
	"strings"
	"testing"
)

// BatchGetSkus 的上限守卫:超限必须在触库之前返回 error_msg。
// 这是纯入参校验,不依赖 DB/Redis,故可用零值 ServiceContext 断言 ——
// 若守卫被误删,测试会因为 nil repo panic 而失败,不会静默放过。
func TestBatchGetSkus_RejectsOverLimit(t *testing.T) {
	const limit = batchGetSkusLimit

	ids := make([]int64, limit+1)
	for i := range ids {
		ids[i] = int64(i + 1)
	}

	l := NewBatchGetSkusLogic(context.Background(), &svc.ServiceContext{})
	resp, err := l.BatchGetSkus(&v1_productv1.BatchGetSkusReq{SkuIds: ids})
	if err != nil {
		t.Fatalf("超限应走业务失败(error_msg),不应返回 gRPC error: %v", err)
	}
	if resp.ErrorMsg == "" {
		t.Fatal("超过上限时必须给出 error_msg")
	}
	if !strings.Contains(resp.ErrorMsg, "200") {
		t.Errorf("error_msg 应说明上限值, got %q", resp.ErrorMsg)
	}
	if len(resp.Items) != 0 {
		t.Errorf("超限时不应返回任何条目, got %d", len(resp.Items))
	}
}

// 空入参直接返回空结果,不触库(购物车为空是常态,不该为此打一次 RPC 内查询)
func TestBatchGetSkus_EmptyInput(t *testing.T) {
	l := NewBatchGetSkusLogic(context.Background(), &svc.ServiceContext{})
	resp, err := l.BatchGetSkus(&v1_productv1.BatchGetSkusReq{})
	if err != nil {
		t.Fatalf("空入参不应报错: %v", err)
	}
	if resp.ErrorMsg != "" || len(resp.Items) != 0 {
		t.Fatalf("空入参应返回空响应, got error_msg=%q items=%d", resp.ErrorMsg, len(resp.Items))
	}
}

// 并发度守卫:批量上限必须是正数,否则 BatchGetSkus 会把所有请求都拒掉
func TestBatchGetSkusLimit_IsPositive(t *testing.T) {
	if batchGetSkusLimit <= 0 {
		t.Fatalf("batchGetSkusLimit 必须为正数, got %d", batchGetSkusLimit)
	}
}
