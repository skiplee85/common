package route

import (
	"github.com/dgrijalva/jwt-go"
)

const (
	// CodeOk 成功
	CodeOk = 0
	// CodeErrorRequest 非法请求
	CodeErrorRequest = 400
	// CodeErrorInternal 服务器内部错误
	CodeErrorInternal = 500
	// CodeInvalidArguments 非法参数
	CodeInvalidArguments = 1000
)

// UserClaims 用户jwt结构
type UserClaims struct {
	UserID int
	Role   int
	jwt.StandardClaims
}

// BaseResponse API 公共响应参数
type BaseResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg,omitempty"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination 分页
type Pagination struct {
	Page  int `form:"page" json:"page" binding:"required,min=0"` // 第几页
	Size  int `form:"size" json:"size" binding:"required,min=0"` // 一页容纳最多容纳多少数据
	Total int `form:"total" json:"total,omitempty"`              // 共有多少数据
}

// GetDefaultPagination 获取默认分页
func GetDefaultPagination() *Pagination {
	return &Pagination{
		Page: 1,
		Size: 20,
	}
}
