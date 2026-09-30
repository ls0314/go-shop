package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserDeptHandler struct {
	userRPC *userclient.PermCodesClient
}

func NewUserDeptHandler(userRPC *userclient.PermCodesClient) *UserDeptHandler {
	return &UserDeptHandler{userRPC: userRPC}
}

// CreateUserDeptRel 为用户分配部门接口(全量替换)
// 路由映射：POST /api/v1/admin/user/assign-dept
// isPrimaryId 是 dept_ids 中的下标,不是部门ID。
func (ud *UserDeptHandler) CreateUserDeptRel(c *gin.Context) {
	var req struct {
		UserID      int64   `json:"user_id"`
		DeptIds     []int64 `json:"dept_ids"`
		IsPrimaryId int64   `json:"isPrimaryId"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := ud.userRPC.AssignUserDepts(req.UserID, req.DeptIds, req.IsPrimaryId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, req)
}

// GetUserDeptRelList 根据用户ID查询关联的部门列表接口
// 路由映射：GET /api/v1/admin/user/:id/dept
func (ud *UserDeptHandler) GetUserDeptRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	userDeptList, total, primaryId, errMsg, err := ud.userRPC.ListUserDepts(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":      userDeptList,
		"total":     total,
		"primaryId": primaryId,
	})
}

// DeleteUserAllDeptRel 根据用户ID清空关联的所有部门接口
// 路由映射：DELETE /api/v1/admin/user/:id/clear-dept
func (ud *UserDeptHandler) DeleteUserAllDeptRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := ud.userRPC.ClearUserDepts(id)
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
