package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// DeptService 部门表服务层实例
type DeptService struct {
	DepartmentRepo *repository.DepartmentRepo
	UserDeptRepo   *repository.UserDeptRepo
	db             *gorm.DB
}

// NewDeptService 创建部门表服务层实例
// 接收值： deptRepo - 部门表数据层实例
// 返回值： *DeptService - 部门表服务层指针
func NewDeptService() *DeptService {
	return &DeptService{
		DepartmentRepo: repository.NewDeptRepo(db.DB),
		UserDeptRepo:   repository.NewUserDeptRepo(db.DB),
		db:             db.DB,
	}
}

// CreateDept 创建部门
// 保证同一父节点下没有相同的部门名
// 接收值： dept - 部门结构体
// 返回值： error - 错误信息
func (d *DeptService) CreateDept(dept *model.SysDept) error {
	// 联合判重：同一父级下部门名不能重复
	existing, _ := d.DepartmentRepo.GetDeptByUK(dept.DeptName, dept.ParentId)
	if existing != nil {
		return model.DeptExist
	}

	// 如果未传入Sort_id自动加到该父节点最后面
	if dept.SortOrder == 0 {
		sortId, err := d.DepartmentRepo.GetMaxSortId(dept.ParentId)
		if err != nil {
			return err
		}
		dept.SortOrder = sortId
	}
	// 开启事务
	tx := d.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	deptTxRepo := d.DepartmentRepo.WithTx(tx)
	// 调用数据层创建部门
	if err := deptTxRepo.CreateDept(dept); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
}

// GetDeptTreeByRoleId 根据用户ID获取部门树
// 接收值：userId - 用户ID
// 返回值：[]*model.SysDept - 部门树列表，error - 错误信息
func (d *DeptService) GetDeptTreeByRoleId(userId int64) ([]*model.SysDept, error) {
	// 获取角色关联的菜单ID列表
	deptIds, err := d.UserDeptRepo.GetUserDeptListByUserId(userId)
	if err != nil {
		return nil, err
	}
	if len(deptIds) == 0 {
		return nil, model.DeptNotExist
	}
	// 批量查询菜单，且已按 sort_order 排序
	deptList, err := d.DepartmentRepo.ListDeptByIds(deptIds)
	if err != nil {
		return nil, err
	}

	// 初始化 menuMap
	deptMap := make(map[int64]*model.SysDept)
	for _, dept := range deptList {
		deptMap[dept.DeptId] = dept
	}

	// 构建树形结构
	var treeList []*model.SysDept
	for _, dept := range deptList {
		parent, hasParent := deptMap[dept.ParentId]
		if !hasParent {
			treeList = append(treeList, dept)
			continue
		}
		parent.Children = append(parent.Children, dept)
	}
	return treeList, nil

}

// GetDept 查询部门信息（根据部门ID）
// 接收值： id - 部门唯一标识
// 返回值： *model.SysDept - 部门对象指针, error - 错误信息
func (d *DeptService) GetDept(id int64) (*model.SysDept, error) {
	// 调用数据层查询部门
	dept, err := d.DepartmentRepo.GetDeptById(id)
	if err != nil {
		return nil, model.DeptNotExist
	}
	return dept, nil
}

// GetDeptList 分页查询部门列表
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//	deptType - 部门类型（可选筛选）
//
// 返回值：
//
//	[]model.SysDept - 部门列表
//	int64 - 总数
//	error - 错误信息
func (d *DeptService) GetDeptList(page, pageSize int, deptType string) ([]model.SysDept, int64, error) {
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
	// 调用数据层返回分页数据
	return d.DepartmentRepo.GetDeptList(page, pageSize, deptType)
}

// UpdateDept 更新部门信息
// 接收值：
//
//	deptId - 部门ID
//	updateDept - 待更新字段map
//
// 返回值：
//
//	error - 错误信息
func (d *DeptService) UpdateDept(deptId int64, updateDept map[string]interface{}) error {
	// 查询原部门是否存在
	oldDept, err := d.DepartmentRepo.GetDeptById(deptId)
	if err != nil {
		return model.DeptNotExist
	}

	// 复制原部门数据
	newDept := *oldDept

	// mapstructure 解析更新字段（仅覆盖传入字段）
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newDept,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(updateDept); err != nil {
		return err
	}

	// 联合唯一索引校验：修改了部门名才需要校验
	if newDept.DeptName != oldDept.DeptName {
		existing, _ := d.DepartmentRepo.GetDeptByUK(newDept.DeptName, newDept.ParentId)
		if existing != nil {
			return model.DeptExist
		}
	}

	// 变更父节点自动插入队尾
	if newDept.ParentId != oldDept.ParentId {
		sortId, err := d.DepartmentRepo.GetMaxSortId(newDept.ParentId)
		if err != nil {
			return err
		}
		newDept.SortOrder = sortId
	}

	// 开启事务
	tx := d.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	deptTxRepo := d.DepartmentRepo.WithTx(tx)
	// 调用数据层更新部门
	if err := deptTxRepo.UpdateDept(&newDept); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

}

// DeleteDept 删除部门
// 校验：部门存在 + 无用户/角色关联
// 接收值： id - 部门ID
// 返回值： error - 错误信息
func (d *DeptService) DeleteDept(id int64) error {
	// 校验部门是否存在
	existing, err := d.DepartmentRepo.GetDeptById(id)
	if err != nil || existing == nil {
		return model.DeptNotExist
	}
	// 检查部门是否存在用户关联
	hasRel, err := d.DepartmentRepo.CheckDeptRelUser(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.DeptHasRel
	}
	// 开启事务
	tx := d.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	deptTxRepo := d.DepartmentRepo.WithTx(tx)
	// 调用数据层删除部门
	if err := deptTxRepo.DeleteDept(id); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
}
