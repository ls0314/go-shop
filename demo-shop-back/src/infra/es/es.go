package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type ESClient struct {
	client *elasticsearch.Client
	index  string
}

type ESProduct struct {
	SpuId        int64     `json:"spu_id"`
	SpuName      string    `json:"spu_name"`
	Brand        string    `json:"brand"`
	Description  string    `json:"description"`
	CategoryId   int64     `json:"category_id"`
	CategoryName string    `json:"category_name"`
	MainImage    string    `json:"main_image"`
	TotalStock   int64     `json:"total_stock"`
	TotalSold    int64     `json:"total_sold"`
	Priority     int64     `json:"priority"`
	SpuStatus    string    `json:"spu_status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MinPrice     float64   `json:"min_price"`
	MaxPrice     float64   `json:"max_price"`
}

func NewESClient(addresses []string) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 ES 客户端失败: %w", err)
	}

	// Ping 校验:客户端是懒连接,new 出来并不真正连 ES, 不 Ping 一下无法发现 ES 是否可用
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("ES Ping 失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("ES Ping 返回错误: %s", res.String())
	}

	return &ESClient{client: client, index: "demo_shop_products"}, nil
}

// IndexProduct 索引文档(创建或全量覆盖)
func (e *ESClient) IndexProduct(ctx context.Context, doc *ESProduct) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("序列化商品文档失败: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      e.index,
		DocumentID: strconv.FormatInt(doc.SpuId, 10), // spu_id 当 _id
		Body:       strings.NewReader(string(body)),  // Body 必须是 io.Reader
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("索引商品 %d 失败: %w", doc.SpuId, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引商品 %d 失败, ES 返回: %s", doc.SpuId, res.String())
	}
	return nil
}

// DeleteProduct 删除文档
func (e *ESClient) DeleteProduct(ctx context.Context, spuId int64) error {
	req := esapi.DeleteRequest{
		Index:      e.index,
		DocumentID: strconv.FormatInt(spuId, 10),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("删除商品 %d 文档失败: %w", spuId, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil // 文档不存在 = 删除成功(幂等)
	}
	if res.IsError() {
		return fmt.Errorf("删除商品 %d 失败, ES 返回: %s", spuId, res.String())
	}
	return nil
}

// SearchRequest 搜索请求(对应文档 4.2 请求参数)
type SearchRequest struct {
	Keyword    string
	CategoryId int64
	Brand      string
	Sort       string // sold / price_asc / price_desc / newest
	Page       int
	PageSize   int
}

// AggBucket 聚合桶(一个桶 = 一个 facet 选项)
type AggBucket struct {
	Key      any   `json:"key"`
	DocCount int64 `json:"doc_count"`
}

// AggResult 聚合结果:品牌 / 类目列表(前端侧边栏筛选)
type AggResult struct {
	Brands     []AggBucket `json:"brands"`
	Categories []AggBucket `json:"categories"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Total    int64
	Products []ESProduct
	Aggs     *AggResult
}

// Search 组合查询 + 排序 + 分页 + 聚合
// Query DSL 用 map 组装再 json.Marshal,保证 JSON 合法(手拼字符串易漏引号)
func (e *ESClient) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	// ① 参数兜底
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建 filter 列表(客观筛选条件,不参与打分)
	filters := []any{
		// 业务规则:仅已上架商品可被搜索
		map[string]any{"term": map[string]any{"spu_status": "published"}},
	}
	if req.CategoryId > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"category_id": req.CategoryId}})
	}
	if req.Brand != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"brand": req.Brand}})
	}

	// 组装 body(对应文档 4.3 的 Query DSL 结构)
	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					// 关键词匹配:spu_name 的 search_analyzer 是 ik_smart,无需显式指定
					map[string]any{"match": map[string]any{"spu_name": req.Keyword}},
				},
				"filter": filters,
			},
		},
		"from": (req.Page - 1) * req.PageSize,
		"size": req.PageSize,
	}
	if sortBody := buildSort(req.Sort); sortBody != nil {
		body["sort"] = sortBody
	}
	// 聚合:一次请求同时返回结果列表 + 筛选选项(facet)
	body["aggs"] = map[string]any{
		"brands":     map[string]any{"terms": map[string]any{"field": "brand"}},
		"categories": map[string]any{"terms": map[string]any{"field": "category_name"}},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化搜索请求失败: %w", err)
	}

	// 执行搜索
	searchReq := esapi.SearchRequest{
		Index: []string{e.index},
		Body:  strings.NewReader(string(jsonBody)),
	}
	res, err := searchReq.Do(ctx, e.client)
	if err != nil {
		return nil, fmt.Errorf("ES 搜索失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("ES 搜索返回错误: %s", res.String())
	}

	// 解析响应:7.x 的 hits.total 是对象 {value, relation},不是裸数字
	var searchResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source ESProduct `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
		Aggregations struct {
			Brands struct {
				Buckets []AggBucket `json:"buckets"`
			} `json:"brands"`
			Categories struct {
				Buckets []AggBucket `json:"buckets"`
			} `json:"categories"`
		} `json:"aggregations"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析 ES 响应失败: %w", err)
	}

	// 映射为 SearchResult
	result := &SearchResult{
		Total:    searchResp.Hits.Total.Value,
		Products: make([]ESProduct, 0, len(searchResp.Hits.Hits)),
		Aggs: &AggResult{
			Brands:     searchResp.Aggregations.Brands.Buckets,
			Categories: searchResp.Aggregations.Categories.Buckets,
		},
	}
	for _, h := range searchResp.Hits.Hits {
		result.Products = append(result.Products, h.Source)
	}
	return result, nil
}

// buildSort 排序映射
func buildSort(sort string) []map[string]any {
	switch sort {
	case "sold":
		return []map[string]any{{"total_sold": map[string]any{"order": "desc"}}}
	case "price_asc":
		return []map[string]any{{"min_price": map[string]any{"order": "asc"}}}
	case "price_desc":
		return []map[string]any{{"min_price": map[string]any{"order": "desc"}}}
	case "newest":
		return []map[string]any{{"created_at": map[string]any{"order": "desc"}}}
	default:
		return nil // 不传 sort → 按相关度打分(_score)排序
	}
}

// BulkIndex 批量索引文档(初始全量脚本与对账任务共用)
// 接收值：docs - 待写入的商品文档列表
// 返回值：成功条数;error - 请求失败或部分失败时返回错误(含失败 spu_id 列表)
func (e *ESClient) BulkIndex(ctx context.Context, docs []ESProduct) (int, error) {
	if len(docs) == 0 {
		return 0, nil // 空输入直接返回,不发无意义请求
	}

	// 拼接 Bulk body:Bulk 协议要求每两条 line 一组
	// {index 元数据行} + {文档行},行尾 \n,最后一行也必须 \n
	var buf strings.Builder
	for _, doc := range docs {
		// 元数据行:动作(index)+ _id(spu_id)
		fmt.Fprintf(&buf, `{"index":{"_id":"%d"}}`+"\n", doc.SpuId)

		docJSON, err := json.Marshal(doc)
		if err != nil {
			return 0, fmt.Errorf("序列化商品 %d 失败: %w", doc.SpuId, err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	// 发送 Bulk 请求(Body 同样是 io.Reader,和 IndexProduct 一致)
	req := esapi.BulkRequest{
		Index: e.index,
		Body:  strings.NewReader(buf.String()),
	}
	res, err := req.Do(ctx, e.client)
	if err != nil {
		return 0, fmt.Errorf("Bulk 写入失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return 0, fmt.Errorf("Bulk 写入失败, ES 返回: %s", res.String())
	}

	// 解析响应 Bulk 是"部分成功"语义, HTTP 200 不代表每条都成功,逐条检查 items
	var bulkResp struct {
		Items []struct {
			Index struct {
				Status int `json:"status"`
				Error  *struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"error"`
			} `json:"index"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return 0, fmt.Errorf("解析 Bulk 响应失败: %w", err)
	}

	// 逐条统计成功/失败(items 顺序与 docs 输入顺序一致)
	success := 0
	var failedIDs []int64
	for i, item := range bulkResp.Items {
		if item.Index.Status >= 200 && item.Index.Status < 300 {
			success++
			continue
		}
		// 失败条目:记录 spu_id 和原因,供日志/重试
		if i < len(docs) {
			failedIDs = append(failedIDs, docs[i].SpuId)
		}
		if item.Index.Error != nil {
			log.Printf("[WARN] Bulk 条目失败 spu_id=%d: %s %s",
				docs[i].SpuId, item.Index.Error.Type, item.Index.Error.Reason)
		}
	}

	if success < len(docs) {
		return success, fmt.Errorf("Bulk 部分失败: 成功 %d/%d, 失败 spu_id=%v",
			success, len(docs), failedIDs)
	}
	return success, nil
}

// ListAllSpuIds 遍历索引全部文档的 spu_id(对账孤儿清理用)
func (e *ESClient) ListAllSpuIds(ctx context.Context) ([]int64, error) {
	const batchSize = 500    // 每批拉取条数
	const maxBatches = 10000 // 防死循环上限(500*10000=500万条,远超 demo 数据)

	ids := make([]int64, 0)

	// 发起首轮 Scroll 请求 Scroll: "1m" 表示游标在 ES 服务端存活 1 分钟, 翻页必须在这个时间内完成,否则游标过期报错
	query := `{"query":{"match_all":{}},"size":` + strconv.Itoa(batchSize) + `}`
	searchReq := esapi.SearchRequest{
		Index:  []string{e.index},
		Body:   strings.NewReader(query),
		Scroll: 1 * time.Minute,
	}
	res, err := searchReq.Do(ctx, e.client)
	if err != nil {
		return nil, fmt.Errorf("Scroll 初始化失败: %w", err)
	}

	var firstResp struct {
		ScrollId string `json:"_scroll_id"`
		Hits     struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&firstResp); err != nil {
		res.Body.Close()
		return nil, fmt.Errorf("解析 Scroll 首轮响应失败: %w", err)
	}
	res.Body.Close()

	// 首轮也带数据,先收集
	for _, h := range firstResp.Hits.Hits {
		id, err := strconv.ParseInt(h.ID, 10, 64)
		if err != nil {
			continue // 跳过非数字 id(理论上不会出现,防御性跳过)
		}
		ids = append(ids, id)
	}

	// 循环翻页,直到某批为空
	scrollID := firstResp.ScrollId
	batches := 1
	for len(firstResp.Hits.Hits) > 0 && batches < maxBatches {
		// 翻页请求:ScrollID 走 URL query 参数,body 只放 scroll 时长
		scrollBody := `{"scroll":"1m"}`
		scrollReq := esapi.ScrollRequest{
			ScrollID: scrollID,
			Body:     strings.NewReader(scrollBody),
		}
		res, err := scrollReq.Do(ctx, e.client)
		if err != nil {
			return nil, fmt.Errorf("Scroll 翻页失败: %w", err)
		}

		var pageResp struct {
			ScrollId string `json:"_scroll_id"`
			Hits     struct {
				Hits []struct {
					ID string `json:"_id"`
				} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.NewDecoder(res.Body).Decode(&pageResp); err != nil {
			res.Body.Close()
			return nil, fmt.Errorf("解析 Scroll 翻页响应失败: %w", err)
		}
		res.Body.Close()

		for _, h := range pageResp.Hits.Hits {
			id, err := strconv.ParseInt(h.ID, 10, 64)
			if err != nil {
				continue
			}
			ids = append(ids, id)
		}

		scrollID = pageResp.ScrollId // 更新游标
		batches++
		firstResp.Hits.Hits = pageResp.Hits.Hits // 更新"本批是否为空"的判断依据
	}

	// 清理服务端游标(释放 search context 内存;即使忘了也会 1 分钟自动过期)
	if scrollID != "" {
		clearReq := esapi.ClearScrollRequest{
			ScrollID: []string{scrollID},
		}
		if clearRes, err := clearReq.Do(ctx, e.client); err == nil {
			clearRes.Body.Close()
		}
	}

	return ids, nil
}
