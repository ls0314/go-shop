package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"demo-shop/services/bff/internal/response"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// maxUploadSize 单文件上限。图片场景够用;BFF 要先把文件转存出去,
// 上限越大越容易被单个请求拖住。
const maxUploadSize = 10 << 20

// allowedExt 扩展名白名单。
//
// 单体不校验,这里收紧:上传落点改成**匿名可读**的对象存储后,
// 任意文件(html/svg/js)都能被浏览器直接打开,是存储型 XSS 的入口。
var allowedExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".webp": true, ".bmp": true,
}

type UploadChunkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadChunkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadChunkLogic {
	return &UploadChunkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadChunk 把 multipart 文件转存对象存储,返回可直接访问的 URL。
//
// 只支持"一次传完"(total_chunks=1):前端 upload.ts 固定这样发。
// 不实现分片 —— 那是一条从未被走过的路径,写了也无从验证。
func (l *UploadChunkLogic) UploadChunk(req *types.UploadChunkReq, file multipart.File, header *multipart.FileHeader) (*types.UploadResp, error) {
	if req.TotalChunks > 1 {
		return nil, fmt.Errorf("%w(暂不支持分片上传)", response.ErrInvalidParam)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt[ext] {
		return nil, fmt.Errorf("%w(不支持的文件类型 %s)", response.ErrInvalidParam, ext)
	}
	if header.Size > maxUploadSize {
		return nil, fmt.Errorf("%w(文件超过 %dMB)", response.ErrInvalidParam, maxUploadSize>>20)
	}

	key := objectKey(req.FileMd5, header.Filename)
	ctx, cancel := context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	url, err := l.svcCtx.Storage.Put(ctx, key, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		// 对象存储不可达 → 503(接线问题,不是用户错了)
		return nil, fmt.Errorf("%w: %v", response.ErrDownstreamUnavailable, err)
	}

	l.Infof("上传完成 key=%s size=%d", key, header.Size)
	return &types.UploadResp{IsCompleted: true, FileUrl: url}, nil
}

// objectKey 优先用文件 MD5:同一张图重复上传天然覆盖自己,不会堆垃圾。
func objectKey(md5, filename string) string {
	if md5 != "" {
		return md5 + strings.ToLower(filepath.Ext(filename))
	}
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(filename))
}
