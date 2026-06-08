package response

// CreateCategoryResp 创建类目接口响应结构体
type CreateCategoryResp struct {
	CategoryId   int64  `json:"category_id"`
	CategoryPath string `json:"category_path"`
}

// GetTreeCategoryResp 获取类目树响应结构体
type GetTreeCategoryResp struct {
	CategoryId    int64                  `json:"category_id"`
	CategoryName  string                 `json:"category_name"`
	CategoryLevel int64                  `json:"category_level"`
	Children      []*GetTreeCategoryResp `json:"children"`
	ParentId      int64                  `json:"-"`
}

// GetListCategoryResp 获取子类目列表响应结构体
type GetListCategoryResp struct {
	CategoryId    int64  `json:"category_id"`
	ParentId      int64  `json:"parent_id"`
	CategoryName  string `json:"category_name"`
	CategoryLevel int64  `json:"category_level"`
	SortOrder     int64  `json:"sort_order"`
	IsLeaf        bool   `json:"is_leaf"`
	IsVisible     bool   `json:"is_visible"`
	Status        string `json:"status"`
}
