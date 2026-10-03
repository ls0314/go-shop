package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"demo-shop/services/product/internal/infra/es"
	"demo-shop/services/product/internal/infra/lock"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// esWatermarkKey 增量对账水位。键名与单体保持一致 ——
// **不换名是有意的**:Redis 是共享实例,换名会让搬迁后从零水位重跑一次全量;
// 而水位本身表达的就是"ES 已同步到哪个时刻",与跑在哪个进程无关。
const esWatermarkKey = "es:reconcile:watermark"

// 锁名与单体一致,理由同上(搬迁前后不能出现两把不同的锁同时跑)
const esReconcileLockName = "es:reconcile"

// ReconcileService ES 数据对账任务。
//
// 从单体 src/task/reconcile.go 平移(DS-A-26 §3:ES 索引同步与对账归 product-service)。
// 职责不变:
//   - 增量对账(短周期):把水位后变化过的 published SPU 同步到 ES
//   - 全量对账(长周期):清理 ES 中的孤儿文档
//
// 防重入:跨实例分布式锁;Redis 不可用时退化为进程内互斥(单机语义)。
type ReconcileService struct {
	repo  *repository.ProductRepo
	es    *es.ESClient
	store *redis.Redis
	locks *lock.TaskLockManager
}

// NewReconcileService 构造对账任务。esClient 为 nil 表示 ES 未启用,Start 会直接返回。
func NewReconcileService(repo *repository.ProductRepo, esClient *es.ESClient, store *redis.Redis) *ReconcileService {
	return &ReconcileService{
		repo:  repo,
		es:    esClient,
		store: store,
		locks: lock.NewTaskLockManager(store),
	}
}

// AsService 包装成可被 go-zero service group 托管的服务(支持优雅退出)。
// 周期 <= 0 的关闭判断在 main 的装配处,不在这里 —— 装配策略与任务实现分开。
func (r *ReconcileService) AsService(incInterval, fullInterval time.Duration) service.Service {
	return newTaskService("es-reconcile", func(ctx context.Context) {
		_ = r.Start(ctx, incInterval, fullInterval)
	})
}

// Start 启动对账定时任务(阻塞),由 service group 托管以便优雅退出。
// incInterval - 增量对账周期;fullInterval - 全量对账周期。
func (r *ReconcileService) Start(ctx context.Context, incInterval, fullInterval time.Duration) error {
	if r.es == nil {
		logx.Info("ES 未启用,跳过 ES 对账任务")
		// 不返回 error:没配 ES 是合法的部署形态(搜索降级为直查 DB),
		// 不该让整个服务起不来
		return nil
	}
	logx.Infof("启动 ES 对账: 增量 %v / 全量 %v", incInterval, fullInterval)

	incTicker := time.NewTicker(incInterval)
	defer incTicker.Stop()
	fullTicker := time.NewTicker(fullInterval)
	defer fullTicker.Stop()

	// 启动时先跑一轮增量(冷启动补齐),再按周期跑
	r.runIncremental(ctx)

	for {
		select {
		case <-ctx.Done():
			logx.Info("ES 对账任务退出")
			return nil
		case <-incTicker.C:
			r.runIncremental(ctx) // 同步执行:上一轮未完时本轮顺延,避免同实例内自我堆叠
		case <-fullTicker.C:
			r.runFull(ctx)
		}
	}
}

// runIncremental 增量对账:水位后变化过的 published SPU → Bulk upsert
func (r *ReconcileService) runIncremental(ctx context.Context) {
	if r.es == nil {
		return
	}
	// 防重入:拿不到锁说明另一个对账(本实例或其他实例)在跑,跳过本轮
	release, ok := r.locks.TryLock(esReconcileLockName, 10*time.Minute)
	if !ok {
		logx.Info("对账任务仍在执行,跳过本轮增量对账")
		return
	}
	defer release()

	// 读水位(第一次跑=零值 → 冷启动全量)
	watermark, err := r.getWatermark(ctx)
	if err != nil {
		logx.Errorf("读水位失败: %v", err)
		return
	}
	logx.Infof("增量对账开始,水位=%v", watermark)

	// 查增量变化的 SPU
	docs, err := r.repo.GetChangedSpuEsDocs(watermark)
	if err != nil {
		logx.Errorf("查询增量 SPU 失败: %v", err)
		return
	}

	// 组装并 Bulk 写入
	esDocs := make([]es.ESProduct, 0, len(docs))
	for i := range docs {
		esDocs = append(esDocs, *toESProduct(&docs[i]))
	}
	success, err := r.es.BulkIndex(ctx, esDocs)
	if err != nil {
		logx.Errorf("增量对账 Bulk 写入失败: %v", err)
	}

	// 更新水位 = now(必须在写入之后,否则漏数据)
	if err := r.setWatermark(ctx, time.Now()); err != nil {
		logx.Errorf("更新水位失败: %v", err)
	}

	logx.Infof("增量对账完成: 变化 %d 条, 成功 %d 条", len(docs), success)
}

// runFull 全量对账:清理 ES 孤儿文档(spu_id 不在 DB published 集合中的)
func (r *ReconcileService) runFull(ctx context.Context) {
	if r.es == nil {
		return
	}
	release, ok := r.locks.TryLock(esReconcileLockName, 10*time.Minute)
	if !ok {
		logx.Info("对账任务仍在执行,跳过本轮全量对账")
		return
	}
	defer release()

	// DB 应存在集合:全部 published SPU 的 id
	docs, err := r.repo.GetAllPublishedSpuEsDocs()
	if err != nil {
		logx.Errorf("查询全量 SPU 失败: %v", err)
		return
	}
	dbIDs := make(map[int64]struct{}, len(docs))
	for _, d := range docs {
		dbIDs[d.SpuId] = struct{}{}
	}

	// ES 实际存在的全部 id
	esIDs, err := r.es.ListAllSpuIds(ctx)
	if err != nil {
		logx.Errorf("遍历 ES 文档失败: %v", err)
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
			logx.Errorf("删除孤儿文档失败 spu_id=%d: %v", id, err)
		}
	}

	// 全量对齐后,水位推进到 now(清掉增量积压)
	if err := r.setWatermark(ctx, time.Now()); err != nil {
		logx.Errorf("更新水位失败: %v", err)
	}

	logx.Infof("全量对账完成: ES 文档 %d 个, 孤儿 %d 个", len(esIDs), len(orphans))
}

// getWatermark 读水位;key 不存在时返回零值时间(导致首次全量)
func (r *ReconcileService) getWatermark(ctx context.Context) (time.Time, error) {
	if r.store == nil {
		return time.Time{}, nil // 无 Redis → 水位恒为零,每次全量(降级可用)
	}
	val, err := r.store.GetCtx(ctx, esWatermarkKey)
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
	if r.store == nil {
		return nil
	}
	// seconds=0 → 不设过期,与单体 cache.Set(..., 0) 同语义
	return r.store.SetexCtx(ctx, esWatermarkKey, t.Format(time.RFC3339), 0)
}

// toESProduct model.SpuESDoc → es.ESProduct。
// 与 logic/productservice 的同名转换刻意重复:那个是"写入即同步"路径的私有辅助,
// 这里是"批量对账"路径,两者唯一的共同点是字段映射 —— 为省 15 行而跨包暴露
// 一个转换函数,反而把两条独立路径耦在一起。
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
