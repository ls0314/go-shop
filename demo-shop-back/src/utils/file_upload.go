package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// 文件上传配置常量
const (
	ChunkSize   = 1 << 20         // 单个切片大小：1MB
	MaxFileSize = 5 << 20         // 最大文件总大小：5MB
	TempDir     = "./temp/chunks" // 临时切片存储目录
	FinalDir    = "./uploads"     // 最终文件存储目录
)

// mergeLocks 合并锁映射，防止同一文件并发合并
var mergeLocks sync.Map

// init 初始化上传所需的目录结构
func init() {
	_ = os.MkdirAll(TempDir, 0755)
	_ = os.MkdirAll(FinalDir, 0755)
}

// GetChunkDir 获取指定文件的临时分片目录路径
// 接收值：fileMD5 - 文件MD5标识
// 返回值：string - 分片目录路径
func GetChunkDir(fileMD5 string) string {
	return filepath.Join(TempDir, fileMD5)
}

// SaveChunk 保存单个分片到临时目录
// 接收值：
//
//	fileMD5     - 文件MD5标识
//	chunkNumber - 分片序号
//	file        - 分片数据流
//
// 返回值：error - 错误信息
func SaveChunk(fileMD5 string, chunkNumber int, file io.Reader) error {
	chunkDir := GetChunkDir(fileMD5)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return errors.New("创建切片目录失败")
	}

	chunkPath := filepath.Join(chunkDir, strconv.Itoa(chunkNumber))
	dst, err := os.Create(chunkPath)
	if err != nil {
		return errors.New("创建切片文件失败")
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

// IsChunkExists 检查指定分片是否已存在
// 接收值：
//
//	fileMD5     - 文件MD5标识
//	chunkNumber - 分片序号
//
// 返回值：bool - 分片是否存在
func IsChunkExists(fileMD5 string, chunkNumber int) bool {
	chunkPath := filepath.Join(GetChunkDir(fileMD5), strconv.Itoa(chunkNumber))
	_, err := os.Stat(chunkPath)
	return err == nil
}

// IsAllChunksUploaded 检查所有分片是否已全部上传完毕
// 接收值：
//
//	fileMD5     - 文件MD5标识
//	totalChunks - 总分片数
//
// 返回值：bool - 是否全部上传完成
func IsAllChunksUploaded(fileMD5 string, totalChunks int) bool {
	for i := 1; i <= totalChunks; i++ {
		if !IsChunkExists(fileMD5, i) {
			return false
		}
	}
	return true
}

// CleanChunks 清理指定文件的临时分片目录
// 接收值：fileMD5 - 文件MD5标识
func CleanChunks(fileMD5 string) {
	chunkDir := GetChunkDir(fileMD5)
	_ = os.RemoveAll(chunkDir)
}

// MergeChunks 合并所有分片为最终文件
// 使用互斥锁防止同一文件被并发合并，合并完成后异步清理临时分片
// 接收值：
//
//	fileMD5     - 文件MD5标识
//	fileName    - 原始文件名（用于提取扩展名）
//	totalChunk  - 总分片数
//
// 返回值：
//
//	string - 最终文件路径
//	error  - 错误信息
func MergeChunks(fileMD5 string, fileName string, totalChunk int) (string, error) {

	// 同一文件的合并操作加锁，防止并发冲突
	lock, _ := mergeLocks.LoadOrStore(fileMD5, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		mergeLocks.Delete(fileMD5)
	}()

	// 使用 FileMD5 作为最终文件名，避免同名冲突
	ext := filepath.Ext(fileName)
	finalFileName := fileMD5 + ext
	finalFilePath := filepath.Join(FinalDir, finalFileName)

	// 已合并过则直接返回
	if _, err := os.Stat(finalFilePath); err == nil {
		return finalFilePath, nil
	}

	// 先写入临时文件，合并成功后再重命名
	tempFilePath := finalFilePath + ".tmp"
	dst, err := os.Create(tempFilePath)
	if err != nil {
		return "", errors.New("创建最终文件失败")
	}

	// 按分片序号依次读取并写入最终文件
	chunkDir := GetChunkDir(fileMD5)
	for i := 1; i <= totalChunk; i++ {
		chunkPath := filepath.Join(chunkDir, strconv.Itoa(i))
		src, err := os.Open(chunkPath)
		if err != nil {
			dst.Close()
			_ = os.Remove(tempFilePath)
			return "", errors.New("打开切片失败")
		}

		_, err = io.Copy(dst, src)
		src.Close()
		if err != nil {
			dst.Close()
			_ = os.Remove(tempFilePath)
			return "", errors.New("合并切片失败")
		}
	}

	// Windows 要求先关闭文件句柄才能重命名
	dst.Close()

	if err := os.Rename(tempFilePath, finalFilePath); err != nil {
		_ = os.Remove(tempFilePath)
		return "", fmt.Errorf("重命名最终文件失败: %w", err)
	}

	// 异步清理临时分片，不阻塞响应
	go CleanChunks(fileMD5)

	return finalFilePath, nil
}

// CalculateFileMD5 计算文件的MD5哈希值
// 接收值：filePath - 文件路径
// 返回值：
//
//	string - MD5十六进制字符串
//	error  - 错误信息
func CalculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidateFileSize 校验文件总大小是否超过限制（5MB）
// 接收值：fileSize - 文件大小（字节）
// 返回值：error - 超限时返回错误
func ValidateFileSize(fileSize int) error {
	if fileSize > MaxFileSize {
		return errors.New(fmt.Sprintf("文件大小不能超过%dMB", MaxFileSize>>20))
	}
	return nil
}

// GetExt 获取文件扩展名（包含点）
// 接收值：fileName - 文件名
// 返回值：string - 扩展名，如 ".png"
func GetExt(fileName string) string {
	return filepath.Ext(fileName)
}

// GetFileName 获取路径中的文件名（不含目录）
// 接收值：filePath - 文件路径
// 返回值：string - 文件名
func GetFileName(filePath string) string {
	return filepath.Base(filePath)
}

// FileExists 检查文件是否存在
// 接收值：filePath - 文件路径
// 返回值：
//
//	bool  - 文件是否存在
//	error - 系统错误（非文件不存在错误）
func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
