package service

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// CategoryService 类目表服务层实例
type CategoryService struct {
	CategoryRepo *repository.CategoryRepo // 类目表数据层实例
	ProductRepo  *repository.ProductRepo
	db           *gorm.DB
}

// NewCategoryService 创建类目表服务层实例
// 接收值：使用全局repository初始化，故无接收值
// 返回值：*CategoryService - 类目表服务层实例指针
func NewCategoryService() *CategoryService {
	return &CategoryService{
		CategoryRepo: repository.NewCategoryRepo(),
		ProductRepo:  repository.NewProductRepo(),
		db:           db.DB,
	}
}

// CreateCategory 创建类目
// 在保证同一父节点下没有相同的类目名的情况下创建新类目并自动计算level和path
// 接收值：menu - 类目结构体
// 返回值：error - 错误信息
func (c *CategoryService) CreateCategory(category *model.SysCategory) (*response.CreateCategoryResp, error) {
	// 判断同一父节点下是否存在相同名字的类目
	uk, err := c.CategoryRepo.GetCategoryUk(category.ParentId, category.CategoryName)
	if err != nil {
		return nil, err
	}

	if uk != nil {
		return nil, model.CategoryUkExist
	}

	// 定义父类目
	parentCategory := &model.SysCategory{}

	// 当父类目Id为0时初始化父类目，否则调用数据层获取父类目信息方便后续计算
	if category.ParentId == 0 {
		parentCategory.CategoryLevel = 0
		parentCategory.CategoryPath = "0"
		parentCategory.Status = "active"
		parentCategory.IsLeaf = false
	} else {
		var err error
		parentCategory, err = c.CategoryRepo.GetCategoryById(category.ParentId)
		if err != nil {
			return nil, model.CategoryParentNotExist
		}
	}

	// 被禁用的父类目下不能添加子类目
	if parentCategory.Status == "disabled" {
		return nil, model.CategoryDisable
	}

	// 最多支持4级类目
	if parentCategory.CategoryLevel >= 4 {
		return nil, model.CategoryLevelDeep
	}

	// 使用父类目level自动计算待创建类目的level
	category.CategoryLevel = parentCategory.CategoryLevel + 1

	// 新创建的类目默认为叶子类目
	category.IsLeaf = true

	// 如果未传入Status,默认初始化为active
	if category.Status == "" {
		category.Status = "active"
	}

	// path要在插入后再计算,先简单初始化
	category.CategoryPath = "0"

	// 开启事务
	tx := c.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建事务实例
	categoryTxRepo := c.CategoryRepo.WithTx(tx)

	// 调用数据层创建类目
	if err := categoryTxRepo.CreateCategory(category); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建类目后按规则计算path
	category.CategoryPath = parentCategory.CategoryPath + "," + strconv.FormatInt(category.CategoryId, 10)

	// 将计算得到的path更新进类目信息
	if err := categoryTxRepo.UpdateCategory(category); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 当父类目非根类目且是叶子类目时,再创建类目后将父节点更新为非叶子类目
	if category.ParentId != 0 && parentCategory.IsLeaf {
		parentCategory.IsLeaf = false
		// 调用数据层更新父类目
		if err := categoryTxRepo.UpdateCategory(parentCategory); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 设置创建响应结构体信息
	categoryResp := &response.CreateCategoryResp{
		CategoryId:   category.CategoryId,
		CategoryPath: category.CategoryPath,
	}

	// 传回响应信息
	return categoryResp, tx.Commit().Error
}

// GetCategory 查询类目信息（根据类目ID）
// 接收值：
//
//	categoryId - 所查询类目唯一标识
//
// 返回值：
//
//	*model.SysCategory - 类目对象指针
//	error - 错误信息
func (c *CategoryService) GetCategory(categoryId int64) (*model.SysCategory, error) {
	// 调用数据层查询类目信息
	category, err := c.CategoryRepo.GetCategoryById(categoryId)
	if err != nil {
		return nil, model.CategoryNotExist
	}
	return category, nil
}

// GetCategoryList 分页查询类目信息
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//
// 返回值:
//
//	[]response.GetListCategoryResp - 分页类目响应信息列表
//	int64 - 类目总数
//	error - 错误信息
func (c *CategoryService) GetCategoryList(page, pageSize int) ([]response.GetListCategoryResp, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	// 调用数据层返回分页类目信息
	categoryList, total, err := c.CategoryRepo.GetCategoryList(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	// 根据返回的分页类目信息初始化类目列表信息响应结构体,以达到过滤字段的目的
	var categoryListResp []response.GetListCategoryResp
	for _, category := range categoryList {
		categoryResp := response.GetListCategoryResp{
			CategoryId:    category.CategoryId,
			ParentId:      category.ParentId,
			CategoryName:  category.CategoryName,
			CategoryLevel: category.CategoryLevel,
			SortOrder:     category.SortOrder,
			IsLeaf:        category.IsLeaf,
			IsVisible:     category.IsVisible,
			Status:        category.Status,
		}
		categoryListResp = append(categoryListResp, categoryResp)
	}
	// 返回类目信息响应结构体列表信息
	return categoryListResp, total, nil
}

// GetCategoryChildrenList 查询类目下子类目列表
// 接收值：
//
//	parentId - 类目唯一标识符
//
// 返回值:
//
//	[]response.GetListCategoryResp - 子类目响应信息列表
//	error - 错误信息
func (c *CategoryService) GetCategoryChildrenList(parentId int64) ([]response.GetListCategoryResp, error) {
	// 调用数据层获取类目信息,保证所查找类目存在
	categoryParent, err := c.CategoryRepo.GetCategoryById(parentId)
	if err != nil {
		return nil, err
	}
	if categoryParent == nil {
		return nil, model.CategoryNotExist
	}

	// 调用数据层获取子类目列表
	categoryChildrenList, err := c.CategoryRepo.GetCategoryByParentId(parentId)
	if err != nil {
		return nil, err
	}

	// 定义并初始化类目列表信息响应
	var categoryListResp []response.GetListCategoryResp
	for _, category := range categoryChildrenList {
		categoryResp := response.GetListCategoryResp{
			CategoryId:    category.CategoryId,
			ParentId:      category.ParentId,
			CategoryName:  category.CategoryName,
			CategoryLevel: category.CategoryLevel,
			SortOrder:     category.SortOrder,
			IsLeaf:        category.IsLeaf,
			IsVisible:     category.IsVisible,
			Status:        category.Status,
		}
		categoryListResp = append(categoryListResp, categoryResp)
	}
	// 返回子类目列表
	return categoryListResp, nil
}

// GetCategoryTree 获取角色类目树
// 接收值：
//
//	level - 所查询指定层级
//	includeDisabled - 是否包含禁用类目
//
// 返回值：
//
//	[]*response.GetTreeCategoryResp - 类目树响应信息
//	error - 错误信息
func (c *CategoryService) GetCategoryTree(level int64, includeDisabled bool) ([]*response.GetTreeCategoryResp, error) {
	// 调用数据层获取符合条件的全部类目
	categoryList, err := c.CategoryRepo.GetAllCategory(level, includeDisabled)
	if err != nil {
		return nil, err
	}

	// 初始化类目树响应和categoryMap
	var categoryListResp []*response.GetTreeCategoryResp
	categoryMap := make(map[int64]*response.GetTreeCategoryResp)
	for _, category := range categoryList {
		categoryResp := &response.GetTreeCategoryResp{
			CategoryId:    category.CategoryId,
			CategoryName:  category.CategoryName,
			CategoryLevel: category.CategoryLevel,
			IsVisible:     category.IsVisible,
			Status:        category.Status,
			ParentId:      category.ParentId,
			Children:      []*response.GetTreeCategoryResp{},
		}
		categoryListResp = append(categoryListResp, categoryResp)
		categoryMap[category.CategoryId] = categoryResp
	}

	// 构建树形结构
	var categoryTree []*response.GetTreeCategoryResp
	for _, category := range categoryListResp {
		parent, hasParent := categoryMap[category.ParentId]
		if !hasParent {
			categoryTree = append(categoryTree, category)
			continue
		}
		parent.Children = append(parent.Children, category)

	}
	return categoryTree, nil

}

// UpdateCategory 更新类目信息
// 接收值：
//
//	categoryId - 更新类目唯一标识
//	updateCategory - 更新类目部分信息
//
// 返回值：
//
//	model.SysCategory - 更新后类目信息
//	error - 错误信息
func (c *CategoryService) UpdateCategory(categoryId int64, updateCategory map[string]interface{}) (model.SysCategory, error) {
	// 返回的类目信息
	var updatedCategory model.SysCategory
	// 开启事务
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 创建事务实例
		categoryTXRepo := c.CategoryRepo.WithTx(tx)
		// 查询旧类目信息
		oldCategory, err := categoryTXRepo.GetCategoryById(categoryId)
		if err != nil {
			return err
		}
		if oldCategory == nil {
			return model.CategoryNotExist
		}
		// 用旧类目信息初始化更新类目信息
		newCategory := *oldCategory
		// 配置mapstructure，用updateMenu覆盖更新类目信息（只覆盖传入字段）
		config := &mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &newCategory,
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return err
		}
		if err := decoder.Decode(updateCategory); err != nil {
			return err
		}
		// 初始化新path
		newPath := oldCategory.CategoryPath
		// 父类目发生改变
		if newCategory.ParentId != oldCategory.ParentId {
			// 新的父类目不能是自己
			if newCategory.ParentId == categoryId {
				return model.CategoryParentInvalid
			}
			// 新的父类目指向根类目
			if newCategory.ParentId == 0 {
				// 更新为一级类目
				newPath = fmt.Sprintf("0,%d", categoryId)
			} else {
				// 查询新的父类目
				parentCategory, err := categoryTXRepo.GetCategoryById(newCategory.ParentId)
				if err != nil {
					return err
				}
				if parentCategory == nil {
					return model.CategoryNotExist
				}
				// 新的父类目不能是自己的子类目
				if strings.HasPrefix(parentCategory.CategoryPath, oldCategory.CategoryPath+",") {
					return model.CategoryParentInvalid
				}
				// 更新path
				newPath = fmt.Sprintf("%s,%d", parentCategory.CategoryPath, categoryId)
			}
			// 将新的path传入待更新的类目信息中
			newCategory.CategoryPath = newPath

		}

		// 更新当前类目普通信息
		if err := categoryTXRepo.UpdateCategory(&newCategory); err != nil {
			return err
		}
		// 如果 path 发生变化，批量更新所有子节点 path
		if oldCategory.CategoryPath != newPath {
			if err := categoryTXRepo.UpdateChildrenPath(oldCategory.CategoryPath, newPath); err != nil {
				return err
			}
		}

		// 查询更新后的类目信息并返回
		categoryResp, err := categoryTXRepo.GetCategoryById(categoryId)
		if err != nil {
			return err
		}
		if categoryResp == nil {
			return model.CategoryNotExist
		}
		updatedCategory = *categoryResp
		return nil
	})
	if err != nil {
		return model.SysCategory{}, err
	}
	// 类目变更后失效缓存
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("category:%d", categoryId))
	}
	// 返回更新后的类目信息
	return updatedCategory, nil
}

// DeleteCategory 删除类目
// 接收值：categoryId - 待删除类目唯一标识
// 返回值：error - 错误信息
func (c *CategoryService) DeleteCategory(categoryId int64) error {
	// 开始事务
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 创建事务实例
		categoryTxRepo := c.CategoryRepo.WithTx(tx)

		// 判断待删除类目是否存在
		category, err := categoryTxRepo.GetCategoryById(categoryId)
		if err != nil {
			return err
		}
		if category == nil {
			return model.CategoryNotExist
		}

		// 判断待删除类目下是否存在子类目
		childCategory, err := categoryTxRepo.GetCategoryByParentId(categoryId)
		if err != nil {
			return err
		}
		if len(childCategory) > 0 {
			return model.CategoryHasChildren
		}

		// 判断待删除类目是否存在关联商品
		if spu, err := c.ProductRepo.GetSpuByCategory(categoryId); err != nil || spu != nil {
			return model.CategoryHasRel
		}

		// 调用数据层删除类目
		if err := categoryTxRepo.DeleteCategoryById(categoryId); err != nil {
			return err
		}

		if category.ParentId != 0 {
			categoryList, err := categoryTxRepo.GetCategoryByParentId(category.ParentId)
			if err != nil {
				return err
			}
			if len(categoryList) == 0 {
				updateCategory, err := categoryTxRepo.GetCategoryById(category.ParentId)
				if err != nil {
					return err
				}
				updateCategory.IsLeaf = true
				return categoryTxRepo.UpdateCategory(updateCategory)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	// 类目变更后失效缓存
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("category:%d", categoryId))
	}
	return nil
}
