package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeptHandler handler层部门对象实例
type DeptHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewDeptHandler 新建handler层部门对象实例
func NewDeptHandler(userRPC *userclient.PermCodesClient) *DeptHandler {
	return &DeptHandler{userRPC: userRPC}
}

// CreateDept 创建部门接口
// 路由映射：POST /api/v1/admin/dept
func (d *DeptHandler) CreateDept(c *gin.Context) {
	var dept model.SysDept
	if err := c.ShouldBind(&dept); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	created, errMsg, err := d.userRPC.CreateDept(&dept)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, created)
}

// GetDept 根据ID获取单个部门信息接口
// 路由映射：GET /api/v1/admin/dept/:id
func (d *DeptHandler) GetDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist)
		return
	}

	dept, errMsg, err := d.userRPC.GetDept(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, dept)
}

// GetDeptTreeByUserId 根据用户Id获取其对应的用户部门树
// 路由映射：GET /api/v1/admin/dept/tree/:userId
func (d *DeptHandler) GetDeptTreeByUserId(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist)
		return
	}

	deptTree, errMsg, err := d.userRPC.GetDeptTreeByUserId(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, deptTree)
}

// GetDeptList 分页获取部门列表接口
// 路由映射：GET /api/v1/admin/dept
func (d *DeptHandler) GetDeptList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	deptType := c.Query("deptType")

	deptList, total, errMsg, err := d.userRPC.ListDepts(page, pageSize, deptType)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":     deptList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateDept 根据ID更新部门信息接口
// 路由映射：PUT /api/v1/admin/dept/:id
func (d *DeptHandler) UpdateDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	// 用 map 接收,只传需要更新的字段
	var updateDept map[string]interface{}
	if err := c.ShouldBind(&updateDept); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	dept, errMsg, err := d.userRPC.UpdateDept(id, updateDept)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, dept)
}

// DeleteDept 根据ID删除部门接口
// 路由映射：DELETE /api/v1/admin/dept/:id
func (d *DeptHandler) DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := d.userRPC.DeleteDept(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, nil)
}
