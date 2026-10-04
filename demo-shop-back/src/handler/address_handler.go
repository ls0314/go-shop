package handler

import (
	"demo-shop-back/src/infra/addressclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddressHandler 用户地址管理handler层实例。
//
// 地址表的所有权已迁 user-service 的 user_db(DS-A-25 §4.5.2 第 1 条),
// 本层只做 "HTTP 入参绑定 → RPC → HTTP 出参"。
//
// 请求体用 addressclient.Address 而不是某个实体的形状:地址是"一次调用"
// 的输入,不是"一行数据"—— 后者会带上 is_deleted 之类由服务端决定的列,
// 让"客户端能传哪些字段"从代码上看不出来。
type AddressHandler struct {
	addressRPC *addressclient.AddressClient
}

// NewAddressHandler 创建地址管理handler层实例
func NewAddressHandler(deps service.ServiceDeps) *AddressHandler {
	return &AddressHandler{
		addressRPC: deps.AddressRPC,
	}
}

// CreateAddress 新增收货地址接口
// 路由映射：POST /api/v1/user/addresses
// 功能：接收前端传入的地址信息，校验参数后调用服务层创建地址。首个地址自动设为默认，超过20条返回错误
// 参数：c *gin.Context Gin上下文，用于获取当前用户、接收请求体、返回响应
// 请求参数：
//
//	receiver_name  - 收货人姓名，string类型，1-50字符，必填
//	receiver_phone - 收货人手机号，string类型，11位，必填
//	province       - 省份，string类型，必填
//	city           - 城市，string类型，必填
//	district       - 区/县，string类型，必填
//	detail_address - 详细地址，string类型，1-200字符，必填
//	postal_code    - 邮政编码，string类型，可选
//	is_default     - 是否默认地址，bool类型，可选（默认false），首次添加自动设为true
//	address_tag    - 地址标签，string类型，可选（家/公司/学校）
//
// 响应：
//
//	400：请求参数绑定失败
//	500：服务层创建地址失败（6001姓名/手机非空 6002手机号格式 6003上限）
//	200：创建成功，返回新地址ID
func (ah *AddressHandler) CreateAddress(c *gin.Context) {
	// 实例化后绑定请求参数
	var address addressclient.Address
	if err := c.ShouldBind(&address); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID并注入地址对象
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	address.UserId = userId
	// 调用服务层创建地址
	addressId, err := ah.addressRPC.CreateAddress(&address)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 创建成功，返回新地址ID
	utils.Success(c, addressId)
}

// GetAddressList 获取用户地址列表接口
// 路由映射：GET /api/v1/user/addresses
// 功能：获取当前登录用户的所有有效地址（排除已删除），默认地址排在最前
// 参数：c *gin.Context Gin上下文，用于获取当前用户、返回响应
// 响应：
//
//	400：用户未登录
//	500：服务层查询地址列表失败
//	200：查询成功，返回地址数组（默认地址排最前）
func (ah *AddressHandler) GetAddressList(c *gin.Context) {
	// 从JWT上下文获取当前用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层获取地址列表
	addressList, err := ah.addressRPC.GetAddressList(userId)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 查询成功，返回地址列表
	utils.Success(c, addressList)
}

// GetAddress 获取单个地址详情接口
// 路由映射：GET /api/v1/user/addresses/:id
// 功能：从URL路径获取地址ID，校验地址归属当前用户后返回地址详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 地址ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层查询地址失败（6004地址不存在 6005无权访问）
//	200：查询成功，返回地址完整信息
func (ah *AddressHandler) GetAddress(c *gin.Context) {
	// 通过传入URL地址获取INT格式的地址ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层查询地址详情（含归属校验）
	address, err := ah.addressRPC.GetAddress(userId, id)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 查询成功，返回地址信息
	utils.Success(c, address)
}

// UpdateAddress 更新收货地址接口
// 路由映射：PUT /api/v1/user/addresses/:id
// 功能：从URL获取地址ID，接收前端传入的更新字段，校验归属后执行部分更新。若设为默认则自动取消旧默认
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、接收请求体、返回响应
// 路径参数：
//
//	id - 地址ID，int64类型
//
// 请求参数（Body，全部可选，仅更新传入字段）：
//
//	receiver_name/receiver_phone/province/city/district/detail_address/postal_code/is_default/address_tag
//
// 响应：
//
//	400：ID格式错误/参数绑定失败/用户未登录
//	500：服务层更新地址失败（6004地址不存在 6005无权访问）
//	200：更新成功，返回更新后的完整地址信息
func (ah *AddressHandler) UpdateAddress(c *gin.Context) {
	// 通过传入URL地址获取INT格式的地址ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 实例化后绑定更新参数（使用map支持部分更新）
	var updateAddress map[string]interface{}
	if err := c.ShouldBind(&updateAddress); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层更新地址（含归属校验）
	address, err := ah.addressRPC.UpdateAddress(userId, id, updateAddress)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 更新成功，返回更新后完整地址
	utils.Success(c, address)
}

// DeleteAddress 删除收货地址接口
// 路由映射：DELETE /api/v1/user/addresses/:id
// 功能：从URL获取地址ID，校验归属后执行软删除。若删除的是默认地址则自动将最近更新的有效地址设为默认
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 地址ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层删除地址失败（6004地址不存在 6005无权访问 6006订单引用保护）
//	200：删除成功
func (ah *AddressHandler) DeleteAddress(c *gin.Context) {
	// 通过传入URL地址获取INT格式的地址ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层执行软删除（含归属校验和默认转移）
	err = ah.addressRPC.DeleteAddress(userId, id)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 删除成功
	utils.Success(c, nil)
}

// SetDefaultAddress 设为默认地址接口
// 路由映射：PUT /api/v1/user/addresses/:id/default
// 功能：从URL获取地址ID，校验归属后在事务内清除旧默认并设置新默认
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 地址ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层设置默认失败（6004地址不存在 6005无权访问）
//	200：设置成功
func (ah *AddressHandler) SetDefaultAddress(c *gin.Context) {
	// 通过传入URL地址获取INT格式的地址ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层在事务内设置默认地址
	err = ah.addressRPC.SetDefaultAddress(userId, id)
	if err != nil {
		failAddressRPC(c, err)
		return
	}
	// 设置成功
	utils.Success(c, nil)
}
