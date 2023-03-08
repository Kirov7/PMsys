package model

import (
	"github.com/Kirov7/project-common/errs"
)

var (
	CacheError             = errs.NewError(501001, "Cache错误")
	DBError                = errs.NewError(501002, "DB错误")
	NoLegalMobile          = errs.NewError(101001, "手机号不合法")
	CaptchaError           = errs.NewError(101002, "验证码不合法")
	CaptchaNotExist        = errs.NewError(101003, "验证码不存在")
	EmailExist             = errs.NewError(101004, "邮箱已被注册")
	AccountExist           = errs.NewError(101005, "账号已被注册")
	MobileExist            = errs.NewError(101006, "手机号已被注册")
	AccountOrPasswordError = errs.NewError(101007, "用户或密码错误")
	NoLogin                = errs.NewError(101008, "未登录")
)
