package payment

import "errors"

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errInvalidParam 回调请求缺少必要参数。
var errInvalidParam = errors.New("请求参数错误")
