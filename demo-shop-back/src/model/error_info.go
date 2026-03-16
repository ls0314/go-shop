package model

import "errors"

var (
	RegPasswordInvalid   = errors.New("密码必须同时包含字母数字标点符号且>=8位")
	LoginPasswordInvalid = errors.New("用户名或密码错误")
	PhoneMalformed       = errors.New("手机号格式错误")
	PhoneExist           = errors.New("手机号已注册")
	UsernameExist        = errors.New("用户名已存在")
	EmailExist           = errors.New("邮箱已注册")
)
