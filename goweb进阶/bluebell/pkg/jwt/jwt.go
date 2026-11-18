package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

const TokenExpireDuration = time.Hour * 24 * 365

var mySecret = []byte("我从山中来")

// MyClaims自定义声明结构体并内嵌jwt.StandardClaims
// jwt包自带的jwt.StandardClaims只包含了官方字段
// 我们这里额外记录一个username字段所以要自定义结构体
type MyClaims struct {
	UserID   int64  `json:"user_id`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenToken 生成JWT
func GenToken(userID int64, username string) (string, error) {
	//创建一个自己的声明
	c := MyClaims{
		userID,
		"username", //自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(viper.GetInt("auth.jwt_expire")) * time.Hour).Unix(), //过期时间
			Issuer:    "bluebell",                                                                        //签发人
		},
	}
	//使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	//使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// ParaseToken 解析JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	//解析token
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if token.Valid {
		//校验token
		return mc, nil
	}
	return nil, errors.New("invalid token")

}
