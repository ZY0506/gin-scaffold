package errors

import (
	"fmt"

	pkgerrors "github.com/pkg/errors"
)

type Error struct {
	Code int    // 业务错误码
	Msg  string // 可读错误信息
	err  error  // 内部包装的 error（含堆栈）
}

func (e *Error) Error() string {
	return e.Msg
}

func (e *Error) Unwrap() error {
	return e.err
}

// New 创建一个带错误码的错误
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg, err: pkgerrors.New(msg)}
}

// Wrap 包装已有 error，附加错误码
func Wrap(err error, code int, msg string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Code: code, Msg: msg, err: pkgerrors.Wrap(err, msg)}
}

// Wrapf 格式化包装错误
func Wrapf(err error, code int, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return &Error{Code: code, Msg: msg, err: pkgerrors.Wrap(err, msg)}
}

// FromError 从普通 error 提取业务错误码，如果是业务错误则返回原错误码，否则返回 Internal
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var bizErr *Error
	if asErr := pkgerrors.As(err, &bizErr); asErr {
		return bizErr
	}
	return New(ErrInternal, err.Error())
}

// IsCode 判断错误码是否匹配
func IsCode(err error, code int) bool {
	var bizErr *Error
	if pkgerrors.As(err, &bizErr) {
		return bizErr.Code == code
	}
	return false
}

// Cause 获取原始 error
func Cause(err error) error {
	return pkgerrors.Cause(err)
}
