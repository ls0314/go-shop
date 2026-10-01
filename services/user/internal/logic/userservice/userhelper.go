package userservicelogic

import (
	"errors"

	"gorm.io/gorm"
)

// isNotFound 判断是否"记录不存在"。
// 唯一性检查与存在性检查都靠它把"不存在"从数据库故障里分出来。
func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
