package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"errors"
	"io"
)

// FileService 文件上传服务层接口
// 定义文件上传相关的业务操作契约
type FileService interface {
	// UploadChunk 处理单个分片上传，所有分片到齐后自动合并
	UploadChunk(req *model.ChunkUploadReq, file io.Reader) (*model.UploadResp, error)
}

// fileServiceImpl 文件上传服务层实现
type fileServiceImpl struct{}

// NewFileService 创建文件上传服务层实例
// 接收值：无接收值
// 返回值：FileService - 文件上传服务层接口
func NewFileService() FileService {
	return &fileServiceImpl{}
}

// UploadChunk 处理切片上传
// 流程：校验文件大小 → 秒传检测 → 断点续传检测 → 保存分片 → 检查是否可合并
// 接收值：
//
//	req  - 切片上传请求参数
//	file - 分片文件数据流
//
// 返回值：
//
//	*model.UploadResp - 上传响应（包含是否完成、文件URL）
//	error             - 错误信息
func (s *fileServiceImpl) UploadChunk(req *model.ChunkUploadReq, file io.Reader) (*model.UploadResp, error) {
	// 1. 校验文件总大小（不超过 5MB）
	if err := utils.ValidateFileSize(req.FileSize); err != nil {
		return nil, err
	}

	// 2. 检查文件是否已存在（秒传：相同MD5的文件跳过上传）
	finalFileName := req.FileMD5 + utils.GetExt(req.FileName)
	finalPath := utils.FinalDir + "/" + finalFileName
	if exists, err := utils.FileExists(finalPath); err == nil && exists {
		return &model.UploadResp{
			IsCompleted: true,
			FileURL:     "/uploads/" + finalFileName,
		}, nil
	}

	// 3. 检查当前分片是否已上传（断点续传：已上传的分片跳过）
	if utils.IsChunkExists(req.FileMD5, req.ChunkNumber) {
		// 如果所有分片已到齐，执行合并
		if utils.IsAllChunksUploaded(req.FileMD5, req.TotalChunks) {
			mergedPath, err := utils.MergeChunks(req.FileMD5, req.FileName, req.TotalChunks)
			if err != nil {
				return nil, errors.New("合并文件失败" + err.Error())
			}
			return &model.UploadResp{
				IsCompleted: true,
				FileURL:     "/uploads/" + utils.GetFileName(mergedPath),
			}, nil
		}

		return &model.UploadResp{
			IsCompleted: false,
		}, nil
	}

	// 4. 保存当前分片到临时目录
	if err := utils.SaveChunk(req.FileMD5, req.ChunkNumber, file); err != nil {
		return nil, errors.New("保存切片失败" + err.Error())
	}

	// 5. 检查是否所有分片均已上传完毕，若是则自动合并
	if utils.IsAllChunksUploaded(req.FileMD5, req.TotalChunks) {
		mergedPath, err := utils.MergeChunks(req.FileMD5, req.FileName, req.TotalChunks)
		if err != nil {
			return nil, errors.New("合并文件失败" + err.Error())
		}
		return &model.UploadResp{
			IsCompleted: true,
			FileURL:     "/uploads/" + utils.GetFileName(mergedPath),
		}, nil
	}

	// 6. 分片上传成功，等待后续分片
	return &model.UploadResp{
		IsCompleted: false,
	}, nil
}
