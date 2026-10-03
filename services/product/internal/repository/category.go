package repository

import (
	"demo-shop/services/product/internal/model"
	"errors"

	"gorm.io/gorm"
)

// CategoryRepo 类目表数据层实例
type CategoryRepo struct {
	DB *gorm.DB
}

func NewCategoryRepo(conn *gorm.DB) *CategoryRepo {
	return &CategoryRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (c *CategoryRepo) WithTx(tx *gorm.DB) *CategoryRepo {
	return &CategoryRepo{DB: tx}
}

// CreateCategory 创建类目
func (c *CategoryRepo) CreateCategory(cate *model.SysCategory) error {
	return c.DB.Create(&cate).Error
}

// GetCategoryById 查询类目信息(按类目ID查)。
// 查不到返回 (nil, nil),与单体行为一致。
func (c *CategoryRepo) GetCategoryById(id int64) (*model.SysCategory, error) {
	var cate model.SysCategory
	err := c.DB.First(&cate, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cate, nil
}

// GetCategoryUk 联合查询类目信息(按父节点ID和类目名查)。
// 查不到返回 (nil, nil),与单体行为一致。
func (c *CategoryRepo) GetCategoryUk(parentId int64, categoryName string) (*model.SysCategory, error) {
	var cate model.SysCategory
	err := c.DB.Where("parent_id = ? AND category_name = ?", parentId, categoryName).First(&cate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cate, nil
}

// GetCategoryByParentId 查询该类目下所有子类目信息
func (c *CategoryRepo) GetCategoryByParentId(parentId int64) ([]model.SysCategory, error) {
	var categoryList []model.SysCategory
	err := c.DB.Where("parent_id = ?", parentId).Find(&categoryList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return categoryList, nil
}

// GetCategoryList 分页查询类目信息
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

// GetAllCategory 查询所有符合条件的类目信息
func (c *CategoryRepo) GetAllCategory(level int64, includeDisabled bool) ([]*model.SysCategory, error) {
	var cateList []*model.SysCategory

	cateDb := c.DB.Model(&model.SysCategory{}).Order("sort_order ASC, category_id ASC")
	if !includeDisabled {
		cateDb = cateDb.Where("status = ?", "active")
	}
	if level > 0 && level <= 4 {
		cateDb = cateDb.Where("category_level = ?", level)
	}

	if err := cateDb.Find(&cateList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return cateList, nil
}

// UpdateCategory 更新类目信息
func (c *CategoryRepo) UpdateCategory(cate *model.SysCategory) error {
	return c.DB.Save(cate).Error
}

// UpdateChildrenPath 批量更新子类目的类目路径
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
func (c *CategoryRepo) DeleteCategoryById(id int64) error {
	return c.DB.Delete(&model.SysCategory{}, id).Error
}
