package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

const TokenExpireDuration = time.Hour * 24 * 365

//var mySecret = []byte("我从山中来")

// MyClaims自定义声明结构体并内嵌jwt.StandardClaims
// jwt包自带的jwt.StandardClaims只包含了官方字段
// 我们这里额外记录一个username字段所以要自定义结构体

var (
	ErrorInvalidAccessToken  = errors.New("无效的access_token")
	ErrorInvalidRefreshToken = errors.New("无效的refresh_token")
)

var (
	mySecret                   = []byte("夏天夏天悄悄过去")
	RefreshTokenExpireDuration = 24 * time.Hour
)

type MyClaims struct {
	UserID             int64  `json:"user_id"`
	Username           string `json:"username"`
	jwt.StandardClaims        //官方字段(ExpiresAt)
}

// GenToken 生成JWT
func GenToken(userID int64, username string) (string, error) {
	//创建一个自己的声明
	c := MyClaims{
		userID,
		username, //自定义字段
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
	//var mc = new(MyClaims)
	//token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
	//	return mySecret, nil
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if token.Valid {
	//	//校验token
	//	return mc, nil
	//}
	//return nil, errors.New("invalid token")

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	mc, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, ErrorInvalidAccessToken
	}
	return mc, nil

}

// ParseExpireAccessToken 解析"已过期"的 access token（刷新流程用）
func ParseExpireAccessToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return mySecret, nil
	})
	v, _ := err.(*jwt.ValidationError) //转换成jwt提供的异常类型
	if v.Errors != jwt.ValidationErrorExpired {
		return nil, ErrorInvalidAccessToken
	}
	fmt.Println("is expired access")
	mc, ok := token.Claims.(*MyClaims)
	if !ok {
		return nil, ErrorInvalidAccessToken
	}
	return mc, nil
}

// GenRefreshToken 生成 refresh token
func GenRefreshToken(userID int64, username string) (token string, err error) {
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Duration(viper.GetInt("auth.refresh_expire")) * time.Hour).Unix(), //过期时间
		Issuer:    "bluebell",
	}).SignedString(mySecret)
	return
}

// CheckRefreshToken 校验 refresh token 是否有效
func CheckRefreshToken(refreshToken string) (err error) {
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return mySecret, nil
	})
	if err != nil {
		return
	}
	_, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		err = ErrorInvalidRefreshToken
		return
	}
	return
}

// RefreshAccessToken 用（可能已过期的）access token 换新 access token
func RefreshAccessToken(AccessToken string) (newAccessToken string, err error) {
	var mc MyClaims
	_, err = jwt.ParseWithClaims(AccessToken, &mc, func(token *jwt.Token) (i interface{}, err error) {
		return mySecret, nil
	})
	//依旧有效的access_token 直接返回
	if err == nil {
		newAccessToken = AccessToken
		return
	}
	v, _ := err.(*jwt.ValidationError)
	if v.Errors == jwt.ValidationErrorExpired {
		newAccessToken, err = GenToken(mc.UserID, mc.Username)
		return
	}
	return
}
