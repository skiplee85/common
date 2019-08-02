package route

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

const (
	keyUserClaims = "keyUserClaims"
	keyRole       = "keyRole"
)

var (
	errMsg = map[int]string{
		CodeOk:                    "success",
		CodeErrorRequest:          "error request",
		CodeErrorInternal:         "server error.",
		CodeErrorInvalidArguments: "invalid arguments.",
	}
	jwtSecret = ""
)

// Context gin.Context 二次封装
type Context struct {
	*gin.Context
}

// BaseRoute 路由表
type BaseRoute struct {
	Method      string
	Path        string
	Handler     func(*Context)
	Middlewares []gin.HandlerFunc
	Role        int // 访问所需的最小角色，高级角色拥有低级角色的权限。0 表示没要求，未授权的用户也可以访问。
	Child       []*BaseRoute
}

// InitErrorMsg 初始化自定义错误码
func InitErrorMsg(msgMap map[int]string) {
	for no, msg := range msgMap {
		errMsg[no] = msg
	}
}

// GetRouteHandler 获取路由
func GetRouteHandler(routeConf []*BaseRoute, jwtToken string, isDebug bool) http.Handler {
	jwtSecret = jwtToken
	if !isDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	defaultRoute := gin.Default()
	for _, r := range routeConf {
		createRouteHandler(r, &defaultRoute.RouterGroup, 0, nil)
	}

	// 跨域请求
	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "*"},
		AllowedHeaders:   []string{"Authorization", "*"},
		AllowCredentials: true,
		Debug:            false,
	})

	return c.Handler(defaultRoute)
}

func createRouteHandler(rConf *BaseRoute, g *gin.RouterGroup, role int, hf []gin.HandlerFunc) {
	r := *rConf
	if r.Role > 0 {
		role = r.Role
	}
	if len(r.Middlewares) > 0 {
		if hf == nil {
			hf = []gin.HandlerFunc{}
		}
		hf = append(hf, r.Middlewares...)
	}
	// group
	if len(r.Child) > 0 {
		gg := g.Group(r.Path)
		for _, rr := range r.Child {
			createRouteHandler(rr, gg, role, hf)
		}
	} else {
		hs := []gin.HandlerFunc{}
		h := func(c *gin.Context) {
			r.Handler(&Context{Context: c})
		}
		if role > 0 && jwtSecret != "" {
			hs = append(hs, getRoleMiddleware(role), authMiddleware)
		}
		if len(hf) > 0 {
			hs = append(hs, hf...)
		}
		hs = append(hs, h)
		// 自定义中间件
		if len(r.Middlewares) > 0 {
			hs = append(hs, r.Middlewares...)
		}
		g.Handle(r.Method, r.Path, hs...)
	}
}

// ValidaArgs 检查参数
func (c *Context) ValidaArgs(args interface{}) error {
	if err := c.ShouldBind(args); err != nil {
		log.Printf("invalid args. %v. url=%s", err.Error(), c.Request.URL.String())
		c.AbortWithStatusJSON(http.StatusOK, &BaseResponse{Code: CodeErrorInvalidArguments, Msg: err.Error()})
		return err
	}
	return nil
}

// SendError 下发错误
func (c *Context) SendError(code int, msgs ...string) {
	httpStatus := http.StatusInternalServerError
	if code >= 1000 {
		httpStatus = http.StatusBadRequest
	}
	m := errMsg[code]
	if len(msgs) > 0 {
		m = msgs[0]
	}
	c.AbortWithStatusJSON(httpStatus, &BaseResponse{
		Code: code,
		Msg:  m,
	})
}

// Send 下发消息
func (c *Context) Send(data interface{}) {
	c.JSON(http.StatusOK, &BaseResponse{
		Data: data,
	})
}

// SendWithPagination 带翻页信息
func (c *Context) SendWithPagination(data interface{}, p *Pagination) {
	c.JSON(http.StatusOK, &BaseResponse{
		Data:       data,
		Pagination: p,
	})
}

// GetClaims 获取JWT信息包
func (c *Context) GetClaims() *UserClaims {
	v, ok := c.Get(keyUserClaims)
	if !ok {
		return nil
	}
	return v.(*UserClaims)
}

// GetIP 获取IP
func GetIP(c *gin.Context) string {
	ip := ""
	findHeader := []string{
		"X-Forwarded-For",
		"Proxy-Client-IP",
		"WL-Proxy-Client-IP",
		"HTTP_CLIENT_IP",
		"HTTP_X_FORWARDED_FOR",
	}
	for _, h := range findHeader {
		ip = c.GetHeader(h)
		if ip != "" {
			break
		}
	}
	if ip == "" {
		ips := strings.Split(c.Request.RemoteAddr, ":")
		ip = ips[0]
	}
	return ip
}

// Finish 完成请求
func (c *Context) Finish(data interface{}, eno int) {
	if eno != CodeOk {
		c.SendError(eno)
		return
	}
	c.Send(data)
}
