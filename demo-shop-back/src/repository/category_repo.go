package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"errors"

	"gorm.io/gorm"
)

// CategoryRepo 类目表数据层实例
type CategoryRepo struct {
	DB *gorm.DB
}

// NewCategoryRepo 创建类目表数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*CategoryRepo - 类目表数据层实例指针
func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{
		DB: db.DB,
	}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*CategoryRepo - 绑定事务的类目表数据层指针
func (c *CategoryRepo) WithTx(tx *gorm.DB) *CategoryRepo {
	return &CategoryRepo{DB: tx}
}

// CreateCategory 创建类目
// 接收值：cate - 类目对象指针
// 返回值：error - 错误信息
func (c *CategoryRepo) CreateCategory(cate *model.SysCategory) error {
	return c.DB.Create(&cate).Error
}

// GetCategoryById 查询类目信息(按类目ID查)
// 接收值：id - 所查询类目唯一标识
// 返回值：
//
//	*model.SysCategory - 所查询类目信息
//	error - 错误信息
func (c *CategoryRepo) GetCategoryById(id int64) (*model.SysCategory, error) {
	var cate model.SysCategory
	err := c.DB.First(&cate, id).Error
	if err != nil {
		// 防止使用First出现的查询为空的数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cate, nil
}

// GetCategoryUk 联合查询类目信息(按父节点ID和类目名查）
// 接收值：
//
//	parentId - 所查询类目父节点ID
//	categoryName - 所查询类目名
//
// 返回值：
//
//	*model.SysCategory - 所查询类目对象指针
//	error - 错误信息
func (c *CategoryRepo) GetCategoryUk(parentId int64, categoryName string) (*model.SysCategory, error) {
	var cate model.SysCategory
	err := c.DB.Where("parent_id = ? AND category_name = ?", parentId, categoryName).First(&cate).Error
	if err != nil {
		// 防止使用First出现的查询为空的数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cate, nil
}

// GetCategoryByName 查询类目信息(按类目ID查)
// 接收值：categoryName - 所查询类目名
// 返回值：
//
//	*model.SysCategory - 所查询类目信息
//	error - 错误信息
//func (c *CategoryRepo) GetCategoryByName(categoryName string) (*model.SysCategory, error) {
//	var cate model.SysCategory
//	err := c.DB.Where("category_name = ?", categoryName).First(&cate).Error
//	if err != nil {
//		return nil, err
//	}
//	return &cate, nil
//}

// GetCategoryByParentId 查询该类目下所有子类目信息
// 接收值：
//
//	parentId - 所查询类目父节点ID
//
// 返回值：
//
//	[]model.SysCategory - 所查询类目下所有子类目信息
//	error - 错误信息
func (c *CategoryRepo) GetCategoryByParentId(parentId int64) ([]model.SysCategory, error) {
	var categoryList []model.SysCategory
	err := c.DB.Where("parent_id = ?", parentId).Find(&categoryList).Error
	if err != nil {
		// 防止使用First出现的查询为空的数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return categoryList, nil
}

// GetCategoryList 分页查询类目信息（可根据类型查询）
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//
// 返回值:
//
//	[]model.SysCategory - 分页类目信息
//	int64 - 类目总数
//	error - 错误信息
func (c *CategoryRepo) GetCategoryList(page, pageSize int) ([]*model.SysCategory, int64, error) {
	var cateList []*model.SysCategory
	var total int64

	cateDb := c.DB.Model(&model.SysCategory{}).Order("sort_order ASC, category_id ASC")

	if err := cateDb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := cateDb.Offset(offset).Limit(pageSize).Find(&cateList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	return cateList, total, nil
}

// GetAllCategory 查询所有符合条件的类目信息（可无条件查询）
// 接收值：
//
//	level - 层级
//	includeDisabled - 是否包含禁用类目
//
// 返回值:
//
//	[]*model.SysMenu - 所有符合条件的类目信息
//	error - 错误信息
func (c *CategoryRepo) GetAllCategory(level int64, includeDisabled bool) ([]*model.SysCategory, error) {
	var cateList []*model.SysCategory

	cateDb := c.DB.Model(&model.SysCategory{}).Order("sort_order ASC, category_id ASC")
	// 按includeDisabled条件查询
	if !includeDisabled {
		cateDb = cateDb.Where("status = ?", "active")
	}
	// 若传入level参数，则按条件过滤数据
	if level > 0 && level <= 4 {
		cateDb = cateDb.Where("category_level = ?", level)
	}

	if err := cateDb.Find(&cateList).Error; err != nil {
		// 防止使用First出现的查询为空的数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return cateList, nil
}

// UpdateCategory 更新类目信息
// 接收值：
//
//	cate - 待更新类目信息
//
// 返回值:
//
//	error - 错误信息
func (c *CategoryRepo) UpdateCategory(cate *model.SysCategory) error {
	return c.DB.Save(cate).Error
}

// UpdateChildrenPath 批量更新类目信息中的类目路径
// 接收值：
//
//	oldPath - 待更新路径
//	newPath - 更新后的路径
//
// 返回值:
//
//	error - 错误信息
func (c *CategoryRepo) UpdateChildrenPath(oldPath, newPath string) error {
	return c.DB.Model(&model.SysCategory{}).
		Where("category_path LIKE ?", oldPath+",%").
		Update(
			"category_path",
			gorm.Expr(
				"REPLACE(category_path, ?::varchar, ?::varchar)",
				oldPath+",",
				newPath+",",
			),
		).Error
}

// DeleteCategoryById 删除类目
// 接收值：
//
//	id - 待删除的类目唯一标识符
//
// 返回值:
//
//	error - 错误信息
func (c *CategoryRepo) DeleteCategoryById(id int64) error {
	return c.DB.Delete(&model.SysCategory{}, id).Error
}
