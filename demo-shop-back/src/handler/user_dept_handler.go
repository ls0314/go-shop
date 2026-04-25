package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserDeptHandler struct {
	UserDeptService *service.UserDeptService
}

func NewUserDeptHandler() *UserDeptHandler {
	return &UserDeptHandler{
		UserDeptService: service.NewUserDeptService(),
	}
}

func (ud *UserDeptHandler) CreateUserDeptRel(c *gin.Context) {
	var userDeptIds struct {
		UserID      int64   `json:"user_id"`
		DeptIds     []int64 `json:"dept_ids"`
		IsPrimaryId int64   `json:"isPrimaryId"`
	}
	if err := c.ShouldBind(&userDeptIds); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := ud.UserDeptService.CreateUserDept(userDeptIds.UserID, userDeptIds.DeptIds, userDeptIds.IsPrimaryId); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, userDeptIds)
}

func (ud *UserDeptHandler) GetUserDeptRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	userDeptList, total, primaryId, err := ud.UserDeptService.GetUserDeptList(id)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"List":      userDeptList,
		"total":     total,
		"primaryId": primaryId,
	})
}

func (ud *UserDeptHandler) DeleteUserAllDeptRel(c *gin.Context) {
	// 从URL路径参数中获取待删除的权限范围ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	if err := ud.UserDeptService.DeleteUserDeptById(id); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
