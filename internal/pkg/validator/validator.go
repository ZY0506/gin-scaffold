package validator

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	// 密码：至少8位，必须包含字母和数字
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// RegisterCustomValidators 注册自定义验证规则到 gin 的 validator
func RegisterCustomValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("password", validatePassword); err != nil {
		return err
	}
	return nil
}

// validatePassword 密码强度校验：至少8位，只能包含字母和数字，必须同时包含字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	pwd := fl.Field().String()
	if len(pwd) < 8 {
		return false
	}
	if !passwordRegex.MatchString(pwd) {
		return false
	}

	hasLetter := false
	hasDigit := false
	for _, r := range pwd {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
