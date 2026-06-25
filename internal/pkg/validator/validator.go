package validator

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators 注册自定义验证规则到 gin 的 validator
func RegisterCustomValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("password", validatePassword); err != nil {
		return err
	}
	return nil
}

// validatePassword 密码强度校验：至少8位，必须同时包含字母和数字，允许特殊字符
func validatePassword(fl validator.FieldLevel) bool {
	pwd := fl.Field().String()
	if len(pwd) < 8 {
		return false
	}

	hasLetter := false
	hasDigit := false
	for _, r := range pwd {
		if unicode.IsLetter(r) {
			hasLetter = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
		// 允许特殊字符，不做限制
	}
	return hasLetter && hasDigit
}
