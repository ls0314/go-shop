package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传handler层实例
type UploadHandler struct {
	fileService service.FileService // 文件上传服务层接口
}

// NewUploadHandler 新建文件上传的HTTP handler实例
// 接收值：无接收值，全局实例化
// 返回值：*UploadHandler - 文件上传handler指针
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		fileService: service.NewFileService(),
	}
}

// UploadChunk 切片上传接口
// 路由映射：POST /api/v1/upload/chunk
// 功能：接收前端传递的文件分片，校验参数后调用服务层处理分片上传，全部上传完毕后自动合并
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	file_md5      - 文件MD5值，作为文件唯一标识
//	file_name     - 原始文件名
//	chunk_number  - 分片序号（从1开始）
//	total_chunks  - 总分片数
//	file_size     - 文件总大小（字节）
//	file          - 分片文件数据
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层处理失败，返回服务器异常信息
//	200：上传成功，返回 is_completed（是否全部完成）和 file_url（完成后返回文件访问路径）
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	// 解析表单参数
	var req model.ChunkUploadReq
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	// 获取上传的文件分片
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	defer file.Close()

	// 调用业务层处理分片上传
	resp, err := h.fileService.UploadChunk(&req, file)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}
