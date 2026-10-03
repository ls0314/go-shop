package tests

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// 唯一标识生成
// ============================================================

var factorySeq atomic.Int64
var orderSeq atomic.Int64

func uniqueTag() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), factorySeq.Add(1))
}

func nextOrderID() int64 {
	return time.Now().UnixMilli()*1_000_000 + orderSeq.Add(1)
}

// ============================================================
// 用户工厂(券链路测试实际用不到;预留给 HTTP 级用例)
// ============================================================

var testPasswordHash = sync.OnceValue(func() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("Test-Passw0rd"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
})

func mustCreateUser(t *testing.T) int64 {
	t.Helper()
	u := model.SysUser{
		Username:     "TESTU-" + uniqueTag(),
		PasswordHash: testPasswordHash(),
		Status:       "active",
	}

	if err := db.DB.Select("username", "password_hash", "status").Create(&u).Error; err != nil {
		t.Fatal("造用户失败：v%", err)
	}
	return u.UserID
}

func mustCreateUsers(t *testing.T, n int) []int64 {
	t.Helper()
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, mustCreateUser(t))
	}
	return ids
}

// ============================================================
// 库存工厂 —— 真外键链:sys_category ← sys_product_spu ← sys_product_sku,按链造
// ============================================================

func mustCreateSkuWithStock(t *testing.T, stock int64) (skuId, spuId int64) {
	t.Helper()

	cat := model.SysCategory{
		CategoryName:  "TESTCAT-" + uniqueTag(),
		CategoryLevel: 1,
		CategoryPath:  "0",
		IsLeaf:        true,
		Status:        "active",
	}

	// 注:原为 "id_leaf"(拼写错误,模型里没有该字段)。此列有 DB 默认值 FALSE,故不影响建行;
	// IsLeaf 从未被赋 true,TestFactorySmoke 里对 IsLeaf 的断言本身是错的(与本轮改造无关)。
	if err := db.DB.Select("category_name", "category_level", "category_path", "is_leaf", "status").Create(&cat).Error; err != nil {
		t.Fatal("造类目失败: v%", err)
	}

	spu := model.SysProductSpu{
		SpuName:    "TESTSPU-" + uniqueTag(),
		CategoryId: cat.CategoryId,
		SpuStatus:  "published",
	}

	if err := db.DB.Select("spu_name", "category_id", "spu_status").Create(&spu).Error; err != nil {
		t.Fatal("造SPU失败：v%", err)
	}

	sku := model.SysProductSku{
		SkuName:   "TESTSKU-" + uniqueTag(),
		SpuId:     spu.SpuId,
		Price:     9.9,
		Stock:     stock,
		SkuCode:   "TESTSKUCODE-" + uniqueTag(),
		SkuStatus: "active",
	}

	if err := db.DB.Select("sku_name", "spu_id", "sku_status", "price", "sku_code", "stock").Create(&sku).Error; err != nil {
		t.Fatal("造SKU失败v%", err)
	}
	return sku.SkuId, spu.SpuId
}

// ============================================================
// 优惠券工厂 —— 已随券域迁走(marketing-service)
// ============================================================
//
// createTemplate / createUserCoupon 曾在此直插 coupon_template / user_coupon,
// 供领券并发与核销幂等用例使用。券表的所有权迁至 marketing-service 的
// marketing_db 后,单体测试进程连不到那些表,夹具与用例一并移交:
// 见 services/marketing/TESTDATA-券域用例待迁.md。

func TestFactorySmoke(t *testing.T) {
	skuId, _ := mustCreateSkuWithStock(t, 10)
	var sku model.SysProductSku
	if err := db.DB.Where("sku_id = ?", skuId).First(&sku).Error; err != nil {
		t.Fatalf("查回SKU失败: %v", err)
	}
	if sku.Stock != 10 || sku.LockStock != 0 || sku.SkuStatus != "active" || sku.IsDeleted {
		t.Fatalf("SKU默认值异常: %+v", sku)
	}
}
