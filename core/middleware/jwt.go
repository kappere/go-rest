// jwt使用说明
//
// 登录签发jwt
// engine.POST("/login", func(c *rest.Context) {
// 	middleware.CreateJwtToken(c, &middleware.UserClaims{
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			Subject:  "123456",
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(20 * time.Second)),
// 		},
// 	})
// 	c.JSON(http.StatusOK, rest.Success(nil))
// })
//
// jwt拦截器校验(不通过返回code=-999(rest.STATUS_NO_AUTHENTICATION))
// jwtGroup := engine.Group("/", middleware.JwtAuth(nil))
// jwtGroup.GET("/demo", DemoHandler())
//
package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kappere/go-rest/core/rest"
)

var (
	// 私钥签发jwt
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	// 公钥验证jwt
	publicKey *rsa.PublicKey
)

type UserClaims struct {
	jwt.RegisteredClaims
	Extra map[string]string `json:"extra,omitempty"`
}

func JwtAuth(pubKey *rsa.PublicKey) rest.HandlerFunc {
	// 未传入公钥，默认本地作为jwt签发方
	if pubKey == nil {
		publicKey = &privateKey.PublicKey
	}
	return BasicAuth(func(c *rest.Context) bool {
		tokenString := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil {
			c.JSON(http.StatusOK, rest.ErrorWithCode(err.Error(), rest.STATUS_NO_AUTHENTICATION))
			return false
		}
		claims, ok := token.Claims.(*UserClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusOK, rest.ErrorWithCode("invalid jwt token", rest.STATUS_NO_AUTHENTICATION))
			return false
		}
		refreshJwtToken(c, claims)
		c.Set("jwt/claims", claims)
		return true
	})
}

func CreateJwtToken(c *rest.Context, claims *UserClaims) string {
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	c.Writer.Header().Add("jwt", tokenString)
	return tokenString
}

func refreshJwtToken(c *rest.Context, claims *UserClaims) {
	duration := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	if time.Now().After(claims.ExpiresAt.Add(-duration / 2)) {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(duration))
		CreateJwtToken(c, claims)
	}
}
