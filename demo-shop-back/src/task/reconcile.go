// Package task 后台定时任务(ES 数据对账等)
package task

import (
	"context"
	"errors"
	"fmt"
	"log"

	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/es"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	watermarkKey = "es:reconcile:watermark" // 增量对账水位(存 Redis)
)

// ReconcileService ES 数据对账任务
// 职责：
//   - 增量对账(短周期)：把水位后变化过的 published SPU 同步到 ES
//   - 全量对账(长周期)：清理 ES 中的孤儿文档
//
// 防重入：分布式锁(Redsync)保证跨实例同一时刻只有一个对账在跑;
// Redis 不可用时退化为进程内互斥(单机语义)
type ReconcileService struct {
	repo  *repository.ProductRepo
	es    *es.ESClient
	cache *cache.RedisService
	locks *DistributedLockManager
}

// NewReconcileService 创建对账服务(依赖从全局 infra 获取)
func NewReconcileService() *ReconcileService {
	return &ReconcileService{
		repo:  repository.NewProductRepo(db.DB),
		es:    infra.GetES(),
		cache: infra.GetCache(),
		locks: NewDistributedLockManager(infra.GetCache()),
	}
}

// Start 启动对账定时任务(阻塞调用方,通常放 goroutine)
// incInterval - 增量对账周期(5 分钟)
// fullInterval - 全量对账周期(1 小时)
func (r *ReconcileService) Start(incInterval, fullInterval time.Duration) {
	if r.es == nil {
		log.Println("[WARN] ES 未初始化,跳过对账任务")
		return
	}
	log.Printf("[INFO] 启动 ES 对账: 增量 %v / 全量 %v", incInterval, fullInterval)

	// 启动时先跑一次增量(冷启动),然后按周期跑
	r.runIncremental()

	incTicker := time.NewTicker(incInterval)
	defer incTicker.Stop()
	fullTicker := time.NewTicker(fullInterval)
	defer fullTicker.Stop()

	for {
		select {
		case <-incTicker.C:
			go r.runIncremental() // 异步,不阻塞 ticker
		case <-fullTicker.C:
			go r.runFull()
		}
	}
}

// runIncremental 增量对账:水位后变化过的 published SPU → Bulk upsert
func (r *ReconcileService) runIncremental() {
	if r.es == nil {
		log.Println("[WARN] ES 未初始化,跳过对账")
		return
	}
	// 防重入:拿不到锁说明另一个对账(本实例或其他实例)在跑,跳过本轮
	release, ok := r.locks.TryLock("es:reconcile", 10*time.Minute)
	if !ok {
		log.Println("[WARN] 对账任务仍在执行,跳过本轮增量对账")
		return
	}
	defer release()

	ctx := context.Background()

	// 读水位(第一次跑=零值 → 冷启动全量)
	watermark, err := r.getWatermark(ctx)
	if err != nil {
		log.Printf("[WARN] 读水位失败: %v", err)
		return
	}
	log.Printf("[INFO] 增量对账开始,水位=%v", watermark)

	// 查增量变化的 SPU
	docs, err := r.repo.GetChangedSpuEsDocs(watermark)
	if err != nil {
		log.Printf("[WARN] 查询增量 SPU 失败: %v", err)
		return
	}

	// 组装并 Bulk 写入
	esDocs := make([]es.ESProduct, 0, len(docs))
	for i := range docs {
		esDocs = append(esDocs, *toESProduct(&docs[i]))
	}
	success, err := r.es.BulkIndex(ctx, esDocs)
	if err != nil {
		log.Printf("[WARN] 增量对账 Bulk 写入失败: %v", err)
	}

	// ④ 更新水位 = now(必须在写入之后,否则漏数据)
	if err := r.setWatermark(ctx, time.Now()); err != nil {
		log.Printf("[WARN] 更新水位失败: %v", err)
	}

	log.Printf("[INFO] 增量对账完成: 变化 %d 条, 成功 %d 条", len(docs), success)
}

// runFull 全量对账:清理 ES 孤儿文档(spu_id 不在 DB published 集合中的)
func (r *ReconcileService) runFull() {
	if r.es == nil {
		log.Println("[WARN] ES 未初始化,跳过对账")
		return
	}
	release, ok := r.locks.TryLock("es:reconcile", 10*time.Minute)
	if !ok {
		log.Println("[WARN] 对账任务仍在执行,跳过本轮全量对账")
		return
	}
	defer release()

	ctx := context.Background()

	// DB 应存在集合:全部 published SPU 的 id
	docs, err := r.repo.GetAllPublishedSpuEsDocs()
	if err != nil {
		log.Printf("[WARN] 查询全量 SPU 失败: %v", err)
		return
	}
	dbIDs := make(map[int64]struct{}, len(docs))
	for _, d := range docs {
		dbIDs[d.SpuId] = struct{}{}
	}

	// ES 实际存在的全部 id
	esIDs, err := r.es.ListAllSpuIds(ctx)
	if err != nil {
		log.Printf("[WARN] 遍历 ES 文档失败: %v", err)
		return
	}

	// 差集 = 孤儿 → 逐个删除
	orphans := make([]int64, 0)
	for _, id := range esIDs {
		if _, ok := dbIDs[id]; !ok {
			orphans = append(orphans, id)
		}
	}
	for _, id := range orphans {
		if err := r.es.DeleteProduct(ctx, id); err != nil {
			log.Printf("[WARN] 删除孤儿文档失败 spu_id=%d: %v", id, err)
		}
	}

	// ④ 全量对齐后,水位推进到 now(清掉增量积压)
	if err := r.setWatermark(ctx, time.Now()); err != nil {
		log.Printf("[WARN] 更新水位失败: %v", err)
	}

	log.Printf("[INFO] 全量对账完成: ES 文档 %d 个, 孤儿 %d 个", len(esIDs), len(orphans))
}

// getWatermark 读水位;key 不存在时返回零值时间(导致首次全量)
func (r *ReconcileService) getWatermark(ctx context.Context) (time.Time, error) {
	if r.cache == nil {
		return time.Time{}, nil // 无 Redis → 水位恒为零,每次全量(降级可用)
	}
	val, err := r.cache.Get(ctx, watermarkKey)
	if err != nil {
		// redis.Nil = key 不存在 = 冷启动
		if errors.Is(err, redis.Nil) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析水位失败: %w", err)
	}
	return t, nil
}

// setWatermark 写水位(RFC3339 字符串,永不过期)
func (r *ReconcileService) setWatermark(ctx context.Context, t time.Time) error {
	if r.cache == nil {
		return nil
	}
	return r.cache.Set(ctx, watermarkKey, t.Format(time.RFC3339), 0)
}

// toESProduct model.SpuESDoc → es.ESProduct
func toESProduct(doc *model.SpuESDoc) *es.ESProduct {
	return &es.ESProduct{
		SpuId:        doc.SpuId,
		SpuName:      doc.SpuName,
		Brand:        doc.Brand,
		Description:  doc.Description,
		CategoryId:   doc.CategoryId,
		CategoryName: doc.CategoryName,
		MainImage:    doc.MainImage,
		TotalStock:   doc.TotalStock,
		TotalSold:    doc.TotalSold,
		Priority:     doc.Priority,
		SpuStatus:    doc.SpuStatus,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		MinPrice:     doc.MinPrice,
		MaxPrice:     doc.MaxPrice,
	}
}
