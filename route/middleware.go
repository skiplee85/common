package route

import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func authMiddleware(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		log.Printf("Authorization empty. %s %s", c.Request.Method, c.Request.URL)
		c.AbortWithStatusJSON(http.StatusUnauthorized, &BaseResponse{Code: http.StatusUnauthorized})
		return
	}

	claims, code := parseToken(auth)
	if code == http.StatusOK {
		role := c.GetInt(keyRole)
		if role > claims.Role {
			c.AbortWithStatusJSON(http.StatusForbidden, &BaseResponse{Code: http.StatusForbidden})
		} else {
			c.Set(keyUserClaims, claims)
		}
	} else {
		c.AbortWithStatusJSON(http.StatusOK, &BaseResponse{Code: code})
	}
}

func getRoleMiddleware(role int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keyRole, role)
	}
}

func parseToken(auth string) (*UserClaims, int) {
	token, err := jwt.ParseWithClaims(auth, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		log.Printf("Parse Authorization Fail. %s %+v", auth, err)
		return nil, http.StatusUnauthorized
	}

	claims, ok := token.Claims.(*UserClaims)
	if ok == false || token.Valid == false {
		return nil, http.StatusUnauthorized
	}
	return claims, http.StatusOK
}

// GenJwtToken 生成jwtToken
func GenJwtToken(userID, role, expire int) (*UserClaims, string, error) {
	expireToken := time.Now().Add(time.Duration(expire) * time.Second).Unix()
	claims := UserClaims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireToken,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	st, err := token.SignedString([]byte(jwtSecret))
	return &claims, st, err
}
