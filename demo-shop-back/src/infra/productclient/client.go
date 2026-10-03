package productclient

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/datatypes"
)

// ErrUnavailable 客户端未建连(etcd 连不上 / 服务未注册)时的统一错误。
//
// 为什么做成哨兵错误而不是各处 errors.New:
//   - 调用方要能判"这是依赖缺失(→503)还是业务/基础设施错误(→500)",
//     靠字符串比较太脆,靠 errors.Is 才稳;
//   - 文案本身是契约:HTTP 层据此回 503,接线冒烟测试据此跳过该路由。
var ErrUnavailable = errors.New("product-service 不可用")

// RestoreError 把服务端返回的 error_msg 还原成哨兵错误(能还原时)或普通错误。
//
// 为什么需要:业务失败以 error_msg 走响应体(gRPC error 为 nil),
// 若直接 errors.New(msg),调用方与 HTTP 层就丢了 errors.Is 的判据 ——
// "product-service 不可用"这类依赖缺失会被当成普通 500,无法回 503。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// productCallTimeout 单次商品/类目 RPC 的超时。与 inventoryclient、userclient 同一取舍:
// 用 context.Background() 而非透传请求 ctx,客户端断开后 RPC 仍会跑完
// (最多这个时长),避免写操作被中途取消留下状态不明。
const productCallTimeout = 3 * time.Second

// ProductClient 商品域(product-service)的 RPC 客户端。
//
// 覆盖该服务对外暴露的三个 gRPC 服务:商品、类目、库存。
// 表同库(sys_product_*/sys_category),商品列表要联类目取名称,
// 库存看板要联 SKU 取价格,拆成多个客户端只会让调用方各自建连。
//
// 与 inventoryclient 的分工:
//   - inventoryclient:库存**四操作**(下单锁定/支付扣减/取消释放/退款回补),
//     调用方是 order / payment 域,自带幂等语义与错误码还原;
//   - 本客户端:商品与类目的**读写**、SKU 级**读取**、库存**查询**。
type ProductClient struct {
	product   v1_productv1.ProductServiceClient
	category  v1_productv1.CategoryServiceClient
	inventory v1_productv1.InventoryServiceClient
	conn      *grpc.ClientConn
}

// NewProductClient 建连 product-service(etcd 服务发现)。
func NewProductClient(etcdHosts []string, etcdKey string) (*ProductClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	return &ProductClient{
		product:   v1_productv1.NewProductServiceClient(client.Conn()),
		category:  v1_productv1.NewCategoryServiceClient(client.Conn()),
		inventory: v1_productv1.NewInventoryServiceClient(client.Conn()),
		conn:      client.Conn(),
	}, nil
}

func (c *ProductClient) Close() error { return c.conn.Close() }

// ============================================================
// 商品:管理端
// ============================================================

// CreateProduct 创建商品(SPU + SKU 列表 + 图片列表,服务端同一事务)。
// 返回 (spuId, errMsg, err):errMsg 非空表示业务失败,由调用方决定 HTTP 码。
func (c *ProductClient) CreateProduct(spu *model.SysProductSpu) (int64, string, error) {
	if c == nil {
		return 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.CreateProduct(ctx, &v1_productv1.CreateProductReq{
		Spu: toProtoSpu(spu),
	})
	if err != nil {
		return 0, "", err
	}
	return resp.SpuId, resp.ErrorMsg, nil
}

// GetProductList 管理端商品分页查询
func (c *ProductClient) GetProductList(req requset.SpuQueryReq) (*response.GetProductListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.GetProductList(ctx, &v1_productv1.GetProductListReq{
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		SpuName:    req.SpuName,
		CategoryId: toInt64Value(req.CategoryId),
		SpuStatus:  req.SpuStatus,
		Brand:      req.Brand,
		Sort:       req.Sort,
	})
	if err != nil {
		return nil, "", err
	}
	return toModelProductList(resp), resp.ErrorMsg, nil
}

// GetProduct 管理端商品详情
func (c *ProductClient) GetProduct(spuId int64) (*response.GetProductResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.GetProduct(ctx, &v1_productv1.GetProductReq{SpuId: spuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelProductDetail(resp.Product), "", nil
}

// UpdateProduct 局部更新:updates 为字段名 → 值,服务端返回合并后的完整对象
func (c *ProductClient) UpdateProduct(spuId int64, updates map[string]interface{}) (*response.GetProductResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	updatesJSON, err := json.Marshal(updates)
	if err != nil {
		return nil, "更新字段序列化失败: " + err.Error(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.UpdateProduct(ctx, &v1_productv1.UpdateProductReq{
		SpuId:       spuId,
		UpdatesJson: string(updatesJSON),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelProductDetail(resp.Product), "", nil
}

// UpdateProductFull 全量更新(SKU 与图片按增改删语义对比)
func (c *ProductClient) UpdateProductFull(spuId int64, req requset.FullUpdateProductReq) (*response.GetProductResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.UpdateProductFull(ctx, toProtoFullUpdateReq(spuId, req))
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelProductDetail(resp.Product), "", nil
}

// DeleteProduct 删除商品(软删 SPU 并级联软删 SKU)
func (c *ProductClient) DeleteProduct(spuId int64) (string, error) {
	if c == nil {
		return "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.DeleteProduct(ctx, &v1_productv1.DeleteProductReq{SpuId: spuId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// PublishProduct 上架
func (c *ProductClient) PublishProduct(spuId int64) (string, error) {
	if c == nil {
		return "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.PublishProduct(ctx, &v1_productv1.PublishProductReq{SpuId: spuId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// WithdrawProduct 下架
func (c *ProductClient) WithdrawProduct(spuId int64) (string, error) {
	if c == nil {
		return "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.WithdrawProduct(ctx, &v1_productv1.WithdrawProductReq{SpuId: spuId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ============================================================
// 商品:用户端
// ============================================================

// UserGetProductList 用户端商品列表(仅已上架)
func (c *ProductClient) UserGetProductList(req requset.SpuQueryReq) (*response.UserGetProductListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.UserGetProductList(ctx, &v1_productv1.UserGetProductListReq{
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		SpuName:    req.SpuName,
		CategoryId: toInt64Value(req.CategoryId),
		Brand:      req.Brand,
		Sort:       req.Sort,
	})
	if err != nil {
		return nil, "", err
	}
	return toModelUserProductList(resp), resp.ErrorMsg, nil
}

// UserGetProduct 用户端商品详情
func (c *ProductClient) UserGetProduct(spuId int64) (*response.UserGetProductResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.UserGetProduct(ctx, &v1_productv1.UserGetProductReq{SpuId: spuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelUserProductDetail(resp.Product), "", nil
}

// ============================================================
// SKU / SPU 级读接口(购物车校验、库存看板)
// ============================================================

// ErrSkuUnavailable 商品不可购买(下架/软删/SKU 禁用)。
// 与库存四操作不同,这里无法复用单体 model 的既有错误变量:
// 单体是在购物车里手写三条件判定,没有对应的错误常量。
var ErrSkuUnavailable = errors.New("商品已下架或不可购买")

// GetSku 取 SKU 快照。
// 返回的 SysProductSku 只保证 sku_id/spu_id/price/stock 等实体字段可用 ——
// 归属 SPU 的状态不在这个结构里,要用 IsSkuPurchasable 判定。
func (c *ProductClient) GetSku(skuId int64) (*model.SysProductSku, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.GetSku(ctx, &v1_productv1.GetSkuReq{SkuId: skuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelSku(resp.Sku), "", nil
}

// IsSkuPurchasable 判定 SKU 能否购买:SKU 未禁用未删除,且归属 SPU 已上架未删除。
//
// 判定放在服务端(product-service 的 GetSkuResp 已算好三个状态位),
// 调用方不再自己拼条件 —— 单体时代 cart_item_service 手写过一遍,
// 规则漂移的风险随调用方数量线性增长。
func (c *ProductClient) IsSkuPurchasable(skuId int64) (bool, string, error) {
	if c == nil {
		return false, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.GetSku(ctx, &v1_productv1.GetSkuReq{SkuId: skuId})
	if err != nil {
		return false, "", err
	}
	if resp.ErrorMsg != "" {
		return false, resp.ErrorMsg, nil
	}
	if resp.Sku == nil || resp.SkuDeleted || !resp.SkuActive {
		return false, model.ErrSkuDisabled.Error(), nil
	}
	if resp.SpuDeleted || !resp.SpuPublished {
		return false, model.ErrSpuDisabled.Error(), nil
	}
	return true, "", nil
}

// BatchGetSkus 批量取 SKU。查不到的 ID 静默缺席,由调用方按差集判定失效项。
//
// 返回的 CartSkuSnapshot 自带归属 SPU 的名称/主图/状态,
// 调用方(购物车列表)不必再逐个查 SPU。
func (c *ProductClient) BatchGetSkus(skuIds []int64) ([]CartSkuSnapshot, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	if len(skuIds) == 0 {
		return nil, "", nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.BatchGetSkus(ctx, &v1_productv1.BatchGetSkusReq{SkuIds: skuIds})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	out := make([]CartSkuSnapshot, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, CartSkuSnapshot{
			Sku:          toModelSku(it.Sku),
			SpuId:        it.Sku.GetSpuId(),
			SpuName:      it.SpuName,
			SpuMainImage: it.SpuMainImage,
			SpuStatus:    it.SpuStatus,
			SpuPublished: it.SpuPublished,
			SpuDeleted:   it.SpuDeleted,
			SkuActive:    it.SkuActive,
			SkuDeleted:   it.SkuDeleted,
		})
	}
	return out, "", nil
}

// CartSkuSnapshot 购物车列表要展示的商品字段集合:
// SKU 实体 + 归属 SPU 的三个展示/判定字段。一次批量调用即可拼齐,
// 避免"批量取 SKU 再逐个取 SPU"的 N+1。
type CartSkuSnapshot struct {
	Sku          *model.SysProductSku
	SpuId        int64
	SpuName      string
	SpuMainImage string
	SpuStatus    string
	SpuPublished bool
	SpuDeleted   bool
	SkuActive    bool
	SkuDeleted   bool
}

// GetSpu 取 SPU 快照(不含 SKU/图片列表)。
// 返回的 SysProductSpu 只填充了 SPU 自身字段,sku_list/image_list 恒为 nil ——
// 需要明细请用 GetProduct。
func (c *ProductClient) GetSpu(spuId int64) (*model.SysProductSpu, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.product.GetSpu(ctx, &v1_productv1.GetSpuReq{SpuId: spuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	p := resp.Spu
	if p == nil {
		return nil, "", nil
	}
	return &model.SysProductSpu{
		SpuId:        p.SpuId,
		SpuName:      p.SpuName,
		CategoryId:   p.CategoryId,
		Brand:        p.Brand,
		Description:  p.Description,
		MainImage:    p.MainImage,
		SpecTemplate: datatypes.JSON(p.SpecTemplate),
		SpuStatus:    p.SpuStatus,
		Priority:     p.Priority,
		IsDeleted:    p.IsDeleted,
	}, "", nil
}

// GetSkuStock 单个 SKU 的库存详情(管理端库存看板)
func (c *ProductClient) GetSkuStock(skuId int64) (*response.SkuInventoryResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.inventory.GetSkuStock(ctx, &v1_productv1.GetSkuStockReq{SkuId: skuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelSkuInventory(resp.Sku), "", nil
}

// GetSkuStockList 一个 SPU 下全部 SKU 的库存 + 聚合值
func (c *ProductClient) GetSkuStockList(spuId int64) (*response.SkuInventoryListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.inventory.GetSkuStockList(ctx, &v1_productv1.GetSkuStockListReq{SpuId: spuId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.SkuInventoryResp, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, *toModelSkuInventory(it))
	}
	return &response.SkuInventoryListResp{
		List:       list,
		TotalStock: resp.TotalStock,
		TotalLock:  resp.TotalLock,
		TotalSold:  resp.TotalSold,
	}, "", nil
}

// GetWarnStockList 低库存预警列表(按库存升序)
func (c *ProductClient) GetWarnStockList(threshold int, spuStatus string) ([]response.InventoryWarnResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.inventory.GetWarnStockList(ctx, &v1_productv1.GetWarnStockListReq{
		Threshold: int32(threshold),
		SpuStatus: spuStatus,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.InventoryWarnResp, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, response.InventoryWarnResp{
			SkuId:     it.SkuId,
			SpuName:   it.SpuName,
			SkuName:   it.SkuName,
			Stock:     it.Stock,
			LockStock: it.LockStock,
			SoldCount: it.SoldCount,
			SkuStatus: it.SkuStatus,
		})
	}
	return list, "", nil
}

// AdjustStock 手动调整库存(管理端)。
//
// 为什么必须走 RPC 而不是本地事务:实现是行锁 + 写流水 + 改库存三步同事务,
// 拆库后这两张表都在 product-service 的库里,本地事务够不着。
func (c *ProductClient) AdjustStock(skuId, changeQty int64, remark string, userId int64) (*response.InventoryAdjustResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.inventory.AdjustStock(ctx, &v1_productv1.AdjustStockReq{
		SkuId:     skuId,
		ChangeQty: changeQty,
		Remark:    remark,
		UserId:    userId,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return &response.InventoryAdjustResp{
		BeforeStock: resp.BeforeStock,
		AfterStock:  resp.AfterStock,
	}, "", nil
}

// ListStockLogs 分页查询库存流水(管理端库存流水页)。
//
// 为什么必须走 RPC:库存流水的**写入**早已全在 product-service(落在 product_db),
// 而本方法此前读 demo_shop —— 读写在两个库,页面永远看不到新流水。
func (c *ProductClient) ListStockLogs(req requset.InventoryLogReq) (*response.InventoryLogResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.inventory.ListStockLogs(ctx, &v1_productv1.ListStockLogsReq{
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		SkuId:      toInt64Value(req.SkuId),
		SpuId:      toInt64Value(req.SpuId),
		ChangeType: req.ChangeType,
		StartTime:  toTimestamp(req.StartTime),
		EndTime:    toTimestamp(req.EndTime),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.InventoryLogList, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, response.InventoryLogList{
			SysProductStockLog: model.SysProductStockLog{
				LogId:       it.LogId,
				SkuId:       it.SkuId,
				ChangeType:  it.ChangeType,
				ChangeQty:   it.ChangeQty,
				BeforeStock: it.BeforeStock,
				AfterStock:  it.AfterStock,
				BeforeLock:  it.BeforeLock,
				AfterLock:   it.AfterLock,
				OrderId:     it.OrderId,
				Remark:      it.Remark,
				CreateBy:    it.CreateBy,
				CreatedAt:   asTime(it.CreatedAt),
			},
			SkuName: it.SkuName,
			SpuName: it.SpuName,
		})
	}

	return &response.InventoryLogResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}, "", nil
}

// toTimestamp 把可空时间包成 proto Timestamp;nil 表示"该维度不过滤"
func toTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// ============================================================
// 类目
// ============================================================

// CreateCategory 创建类目(level/path 由服务端按父节点推导)
func (c *ProductClient) CreateCategory(category *model.SysCategory) (*response.CreateCategoryResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.CreateCategory(ctx, &v1_productv1.CreateCategoryReq{
		Category: toProtoCategory(category),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return &response.CreateCategoryResp{
		CategoryId:   resp.CategoryId,
		CategoryPath: resp.CategoryPath,
	}, "", nil
}

// GetCategory 类目详情
func (c *ProductClient) GetCategory(categoryId int64) (*model.SysCategory, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.GetCategory(ctx, &v1_productv1.GetCategoryReq{CategoryId: categoryId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelCategory(resp.Category), "", nil
}

// GetCategoryList 类目分页列表
func (c *ProductClient) GetCategoryList(page, pageSize int) ([]response.GetListCategoryResp, int64, string, error) {
	if c == nil {
		return nil, 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.GetCategoryList(ctx, &v1_productv1.GetCategoryListReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, 0, "", err
	}
	return toModelCategoryList(resp.Items), resp.Total, resp.ErrorMsg, nil
}

// GetCategoryChildren 直接子类目列表
func (c *ProductClient) GetCategoryChildren(parentId int64) ([]response.GetListCategoryResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.GetCategoryChildren(ctx, &v1_productv1.GetCategoryChildrenReq{ParentId: parentId})
	if err != nil {
		return nil, "", err
	}
	return toModelCategoryList(resp.Items), resp.ErrorMsg, nil
}

// GetCategoryTree 类目树
func (c *ProductClient) GetCategoryTree(level int64, includeDisabled bool) ([]*response.GetTreeCategoryResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.GetCategoryTree(ctx, &v1_productv1.GetCategoryTreeReq{
		Level:           level,
		IncludeDisabled: includeDisabled,
	})
	if err != nil {
		return nil, "", err
	}
	return toModelCategoryTree(resp.Items), resp.ErrorMsg, nil
}

// UpdateCategory 局部更新类目,服务端返回合并后的完整对象
func (c *ProductClient) UpdateCategory(categoryId int64, updates map[string]interface{}) (*model.SysCategory, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	updatesJSON, err := json.Marshal(updates)
	if err != nil {
		return nil, "更新字段序列化失败: " + err.Error(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.UpdateCategory(ctx, &v1_productv1.UpdateCategoryReq{
		CategoryId:  categoryId,
		UpdatesJson: string(updatesJSON),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelCategory(resp.Category), "", nil
}

// DeleteCategory 删除类目(有子类目或已关联商品时服务端拒绝)
func (c *ProductClient) DeleteCategory(categoryId int64) (string, error) {
	if c == nil {
		return "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), productCallTimeout)
	defer cancel()

	resp, err := c.category.DeleteCategory(ctx, &v1_productv1.DeleteCategoryReq{CategoryId: categoryId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ============================================================
// 类型转换:单体 model ↔ proto
// ============================================================

// toInt64Value 把可空 int64 包成 wrappers.Int64Value。
// 必须区分"没传"与"传了 0":category_id 为空表示不过滤,传 0 是明确的根类目筛选。
func toInt64Value(v *int64) *wrapperspb.Int64Value {
	if v == nil {
		return nil
	}
	return wrapperspb.Int64(*v)
}

// jsonMapToText 把 JSONB map 转成 proto 承载的 JSON 文本,空 map 视为空串
func jsonMapToText(m datatypes.JSONMap) string {
	if len(m) == 0 {
		return ""
	}
	b, err := m.MarshalJSON()
	if err != nil {
		return ""
	}
	return string(b)
}

// textToJSONMap 把 JSON 文本还原为 map。空串还原为空 map
// (DB 该列 NOT NULL DEFAULT '{}',不能写 NULL)。
func textToJSONMap(s string) datatypes.JSONMap {
	if s == "" {
		return datatypes.JSONMap{}
	}
	m := datatypes.JSONMap{}
	if err := m.UnmarshalJSON([]byte(s)); err != nil {
		return datatypes.JSONMap{}
	}
	return m
}

// toProtoSku 单体 SPU 的 SKU → proto
func toProtoSku(s *model.SysProductSku) *v1_productv1.Sku {
	if s == nil {
		return nil
	}
	return &v1_productv1.Sku{
		SkuId:      s.SkuId,
		SpuId:      s.SpuId,
		SkuName:    s.SkuName,
		SpecValues: jsonMapToText(s.SpecValues),
		Price:      s.Price,
		CostPrice:  s.CostPrice,
		Stock:      s.Stock,
		LockStock:  s.LockStock,
		SoldCount:  s.SoldCount,
		SkuCode:    s.SkuCode,
		SkuImage:   s.SkuImage,
		SkuStatus:  s.SkuStatus,
	}
}

func toProtoImage(i *model.SysProductSpuImage) *v1_productv1.SpuImage {
	if i == nil {
		return nil
	}
	return &v1_productv1.SpuImage{
		ImageId:   i.ImageId,
		SpuId:     i.SpuId,
		ImageUrl:  i.ImageUrl,
		SortOrder: i.SortOrder,
		IsMain:    i.IsMain,
	}
}

// toProtoSpu 单体 SPU(含 SKU/图片列表)→ proto
func toProtoSpu(spu *model.SysProductSpu) *v1_productv1.Spu {
	if spu == nil {
		return nil
	}
	out := &v1_productv1.Spu{
		SpuId:        spu.SpuId,
		SpuName:      spu.SpuName,
		CategoryId:   spu.CategoryId,
		Brand:        spu.Brand,
		Description:  spu.Description,
		MainImage:    spu.MainImage,
		SpecTemplate: string(spu.SpecTemplate),
		SpuStatus:    spu.SpuStatus,
		Priority:     spu.Priority,
		IsDeleted:    spu.IsDeleted,
	}
	if spu.SkuList != nil {
		for i := range *spu.SkuList {
			out.SkuList = append(out.SkuList, toProtoSku(&(*spu.SkuList)[i]))
		}
	}
	if spu.ImageList != nil {
		for i := range *spu.ImageList {
			out.ImageList = append(out.ImageList, toProtoImage(&(*spu.ImageList)[i]))
		}
	}
	return out
}

// toProtoFullUpdateReq 全量更新请求 → proto。
// 单体的"字段为零值即不改"语义靠 proto 的 optional 字段承载:
// category_id/priority 用 wrappers,其余空串/空指针表示不更新。
func toProtoFullUpdateReq(spuId int64, req requset.FullUpdateProductReq) *v1_productv1.UpdateProductFullReq {
	out := &v1_productv1.UpdateProductFullReq{
		SpuId:          spuId,
		SpuName:        req.SpuName,
		Brand:          req.Brand,
		Description:    req.Description,
		MainImage:      req.MainImage,
		DeleteImageIds: req.DeleteImageIds,
	}
	if req.CategoryId != nil {
		out.CategoryId = wrapperspb.Int64(*req.CategoryId)
	}
	if req.Priority != nil {
		out.Priority = wrapperspb.Int64(*req.Priority)
	}
	if req.SpecTemplate != nil {
		out.SpecTemplate = string(*req.SpecTemplate)
	}
	if req.SkuList != nil {
		for i := range *req.SkuList {
			out.SkuList = append(out.SkuList, toProtoSku(&(*req.SkuList)[i]))
		}
	}
	if req.ImageList != nil {
		for i := range *req.ImageList {
			out.ImageList = append(out.ImageList, toProtoImage(&(*req.ImageList)[i]))
		}
	}
	return out
}

func toProtoCategory(c *model.SysCategory) *v1_productv1.Category {
	if c == nil {
		return nil
	}
	return &v1_productv1.Category{
		CategoryId:    c.CategoryId,
		ParentId:      c.ParentId,
		CategoryName:  c.CategoryName,
		CategoryLevel: c.CategoryLevel,
		CategoryPath:  c.CategoryPath,
		SortOrder:     c.SortOrder,
		IconUrl:       c.IconUrl,
		IsLeaf:        c.IsLeaf,
		IsVisible:     c.IsVisible,
		Status:        c.Status,
	}
}

func toModelCategory(p *v1_productv1.Category) *model.SysCategory {
	if p == nil {
		return nil
	}
	return &model.SysCategory{
		CategoryId:    p.CategoryId,
		ParentId:      p.ParentId,
		CategoryName:  p.CategoryName,
		CategoryLevel: p.CategoryLevel,
		CategoryPath:  p.CategoryPath,
		SortOrder:     p.SortOrder,
		IconUrl:       p.IconUrl,
		IsLeaf:        p.IsLeaf,
		IsVisible:     p.IsVisible,
		Status:        p.Status,
	}
}

// toModelCategoryList 类目列表 proto → 单体响应
func toModelCategoryList(items []*v1_productv1.CategoryListItem) []response.GetListCategoryResp {
	out := make([]response.GetListCategoryResp, 0, len(items))
	for _, it := range items {
		out = append(out, response.GetListCategoryResp{
			CategoryId:    it.CategoryId,
			ParentId:      it.ParentId,
			CategoryName:  it.CategoryName,
			CategoryLevel: it.CategoryLevel,
			SortOrder:     it.SortOrder,
			IsLeaf:        it.IsLeaf,
			IsVisible:     it.IsVisible,
			Status:        it.Status,
		})
	}
	return out
}

func toModelCategoryTree(items []*v1_productv1.CategoryTreeNode) []*response.GetTreeCategoryResp {
	out := make([]*response.GetTreeCategoryResp, 0, len(items))
	for _, n := range items {
		out = append(out, &response.GetTreeCategoryResp{
			CategoryId:    n.CategoryId,
			CategoryName:  n.CategoryName,
			CategoryLevel: n.CategoryLevel,
			IsVisible:     n.IsVisible,
			Status:        n.Status,
			Children:      toModelCategoryTree(n.Children),
		})
	}
	return out
}

// toModelProductList 管理端列表 proto → 单体响应
func toModelProductList(p *v1_productv1.GetProductListResp) *response.GetProductListResp {
	if p == nil {
		return nil
	}
	list := make([]response.SpuList, 0, len(p.Items))
	for _, it := range p.Items {
		list = append(list, response.SpuList{
			SpuId:        it.SpuId,
			SpuName:      it.SpuName,
			CategoryId:   it.CategoryId,
			CategoryName: it.CategoryName,
			Brand:        it.Brand,
			MainImage:    it.MainImage,
			SpuStatus:    it.SpuStatus,
			Priority:     it.Priority,
			CreatedAt:    it.GetCreatedAt().AsTime(),
			UpdatedAt:    it.GetUpdatedAt().AsTime(),
			MinPrice:     it.MinPrice,
			MaxPrice:     it.MaxPrice,
			TotalStock:   it.TotalStock,
			TotalSold:    it.TotalSold,
		})
	}
	return &response.GetProductListResp{
		List:     &list,
		Total:    p.Total,
		Page:     int(p.Page),
		PageSize: int(p.PageSize),
	}
}

// toModelUserProductList 用户端列表 proto → 单体响应
func toModelUserProductList(p *v1_productv1.UserGetProductListResp) *response.UserGetProductListResp {
	if p == nil {
		return nil
	}
	list := make([]response.UserSpuList, 0, len(p.Items))
	for _, it := range p.Items {
		list = append(list, response.UserSpuList{
			SpuId:        it.SpuId,
			SpuName:      it.SpuName,
			CategoryName: it.CategoryName,
			Brand:        it.Brand,
			MainImage:    it.MainImage,
			MinPrice:     it.MinPrice,
			MaxPrice:     it.MaxPrice,
			TotalSold:    it.TotalSold,
			Stock:        it.Stock,
		})
	}
	return &response.UserGetProductListResp{
		List:     &list,
		Total:    p.Total,
		Page:     int(p.Page),
		PageSize: int(p.PageSize),
	}
}

// toModelProductDetail proto → 管理端详情。
// spec_template/spec_values 在 proto 里是 JSON 文本,这里还原成 JSONB 类型,
// 保证 handler 序列化出的 JSON 结构与拆分前逐字一致(前端零感知)。
func toModelProductDetail(p *v1_productv1.ProductDetail) *response.GetProductResp {
	if p == nil {
		return nil
	}
	skuList := make([]response.SkuList, 0, len(p.SkuList))
	for _, s := range p.SkuList {
		skuList = append(skuList, response.SkuList{
			SkuId:      s.SkuId,
			SpuId:      s.SpuId,
			SkuName:    s.SkuName,
			SpecValues: textToJSONMap(s.SpecValues),
			Price:      s.Price,
			CostPrice:  s.CostPrice,
			Stock:      s.Stock,
			LockStock:  s.LockStock,
			SoldCount:  s.SoldCount,
			SkuCode:    s.SkuCode,
			SkuImage:   s.SkuImage,
			SkuStatus:  s.SkuStatus,
		})
	}
	imageList := make([]response.ImageList, 0, len(p.ImageList))
	for _, img := range p.ImageList {
		imageList = append(imageList, response.ImageList{
			ImageId:   img.ImageId,
			ImageUrl:  img.ImageUrl,
			SortOrder: img.SortOrder,
			IsMain:    img.IsMain,
		})
	}
	return &response.GetProductResp{
		SpuId:        p.SpuId,
		SpuName:      p.SpuName,
		CategoryId:   p.CategoryId,
		CategoryName: p.CategoryName,
		Brand:        p.Brand,
		Description:  p.Description,
		MainImage:    p.MainImage,
		SpecTemplate: textToJSONText(p.SpecTemplate),
		SpuStatus:    p.SpuStatus,
		Priority:     p.Priority,
		SkuList:      &skuList,
		ImageList:    &imageList,
		CreatedAt:    asTime(p.CreatedAt),
		UpdatedAt:    asTime(p.UpdatedAt),
	}
}

// toModelUserProductDetail proto → 用户端详情(不含 cost_price / lock_stock)
func toModelUserProductDetail(p *v1_productv1.ProductDetail) *response.UserGetProductResp {
	if p == nil {
		return nil
	}
	skuList := make([]response.UserSkuList, 0, len(p.SkuList))
	for _, s := range p.SkuList {
		skuList = append(skuList, response.UserSkuList{
			SkuId:      s.SkuId,
			SpuId:      s.SpuId,
			SkuName:    s.SkuName,
			SpecValues: textToJSONMap(s.SpecValues),
			Price:      s.Price,
			Stock:      s.Stock,
			SoldCount:  s.SoldCount,
			SkuCode:    s.SkuCode,
			SkuImage:   s.SkuImage,
			SkuStatus:  s.SkuStatus,
		})
	}
	imageList := make([]response.ImageList, 0, len(p.ImageList))
	for _, img := range p.ImageList {
		imageList = append(imageList, response.ImageList{
			ImageId:   img.ImageId,
			ImageUrl:  img.ImageUrl,
			SortOrder: img.SortOrder,
			IsMain:    img.IsMain,
		})
	}
	return &response.UserGetProductResp{
		SpuId:        p.SpuId,
		SpuName:      p.SpuName,
		CategoryId:   p.CategoryId,
		CategoryName: p.CategoryName,
		Brand:        p.Brand,
		Description:  p.Description,
		MainImage:    p.MainImage,
		SpecTemplate: textToJSONText(p.SpecTemplate),
		SpuStatus:    p.SpuStatus,
		Priority:     p.Priority,
		SkuList:      &skuList,
		ImageList:    &imageList,
		CreatedAt:    asTime(p.CreatedAt),
		UpdatedAt:    asTime(p.UpdatedAt),
	}
}

// textToJSONText JSON 文本 → JSONB。空串还原为 NULL,与 DB 该列可空一致。
func textToJSONText(s string) datatypes.JSON {
	if s == "" {
		return nil
	}
	return datatypes.JSON([]byte(s))
}

// asTime nil 的时间戳按零值处理,避免调用方多一层判空
func asTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// toModelSku proto SKU → 单体实体。
// created_at/updated_at/create_by/update_by 不在 proto 契约内(SKU 级读接口
// 只服务下单校验与库存看板,不需要审计字段),故不填充。
func toModelSku(s *v1_productv1.Sku) *model.SysProductSku {
	if s == nil {
		return nil
	}
	return &model.SysProductSku{
		SkuId:      s.SkuId,
		SpuId:      s.SpuId,
		SkuName:    s.SkuName,
		SpecValues: textToJSONMap(s.SpecValues),
		Price:      s.Price,
		CostPrice:  s.CostPrice,
		Stock:      s.Stock,
		LockStock:  s.LockStock,
		SoldCount:  s.SoldCount,
		SkuCode:    s.SkuCode,
		SkuImage:   s.SkuImage,
		SkuStatus:  s.SkuStatus,
	}
}

// toModelSkuInventory proto 库存条目 → 单体库存响应。
// sku_status 为 "inactive" 才视为已删除:proto 的 SkuStock 不带 is_deleted,
// 而库存看板只关心"能不能卖"。
func toModelSkuInventory(s *v1_productv1.SkuStock) *response.SkuInventoryResp {
	if s == nil {
		return nil
	}
	return &response.SkuInventoryResp{
		SkuId:      s.SkuId,
		SkuName:    s.SkuName,
		SpuId:      s.SpuId,
		SpuName:    s.SpuName,
		SpecValues: textToJSONMap(s.SpecValues),
		Stock:      s.Stock,
		LockStock:  s.LockStock,
		// 与单体口径一致:可售总量 = 可用 + 锁定
		TotalStock: s.Stock + s.LockStock,
		SoldCount:  s.SoldCount,
		SkuStatus:  s.SkuStatus,
	}
}
