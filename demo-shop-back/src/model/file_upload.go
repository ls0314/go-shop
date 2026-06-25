package model

// ChunkUploadReq 切片上传请求结构体
// 对应前端分片上传接口的 multipart/form-data 请求参数
type ChunkUploadReq struct {
	FileMD5     string `json:"file_md5"      form:"file_md5"`     // 文件MD5值，作为文件唯一标识
	FileName    string `json:"file_name"     form:"file_name"`    // 原始文件名
	ChunkNumber int    `json:"chunk_number"  form:"chunk_number"` // 分片序号（从1开始）
	TotalChunks int    `json:"total_chunks"  form:"total_chunks"` // 总分片数
	FileSize    int    `json:"file_size"     form:"file_size"`    // 文件总大小（字节）
}

// UploadResp 切片上传响应结构体
type UploadResp struct {
	IsCompleted bool   `json:"is_completed"`        // 是否全部上传完成（所有分片已上传并合并）
	FileURL     string `json:"file_url,omitempty"`  // 最终文件访问URL，仅在 IsCompleted=true 时返回
}
