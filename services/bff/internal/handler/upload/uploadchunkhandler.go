// 手写 handler:不使用 goctl 生成的模板。
//
// 原因:go-zero 的 httpx.Parse 走 core/mapping 解析器,而它**不支持
// multipart 文件**(unmarshaler.go 里没有任何 FileHeader/multipart 处理)。
// 文本字段其实能解析(GetFormValues 会调 r.ParseMultipartForm),但文件
// 必须手工 r.FormFile 取,所以这里整体手工解析,不拆成两半。
package upload

import (
	"net/http"
	"strconv"

	// 别名:本包也叫 upload,两个同名包会冲突
	uploadlogic "demo-shop/services/bff/internal/logic/upload"
	"demo-shop/services/bff/internal/response"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadChunkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 32MB 内存阈值:超过的部分 multipart 落到临时文件。
		// 与 logic 的 10MB 上限配合,正常路径不落盘。
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			response.BadRequest(w, http.StatusBadRequest, "请求不是合法的 multipart 表单")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			response.BadRequest(w, http.StatusBadRequest, "缺少文件字段 file")
			return
		}
		defer func() { _ = file.Close() }()

		// 这 5 个字段只作记录,不参与校验:缺失或非法一律归 0。
		req := &types.UploadChunkReq{
			FileMd5:     r.FormValue("file_md5"),
			FileName:    r.FormValue("file_name"),
			ChunkNumber: atoi(r.FormValue("chunk_number")),
			TotalChunks: atoi(r.FormValue("total_chunks")),
			FileSize:    atoi64(r.FormValue("file_size")),
		}

		resp, err := uploadlogic.NewUploadChunkLogic(r.Context(), svcCtx).UploadChunk(req, file, header)
		if err != nil {
			response.Failure(w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
