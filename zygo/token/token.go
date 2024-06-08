package token

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
	"web/zygo"
)

type JwtHandler struct {
	//算法
	Alg string
	//登录认证方法
	Authenticator func(ctx *zygo.Context) (map[string]any, error)
	//过期时间 秒
	TimeOut        time.Duration
	RefreshTimeOut time.Duration
	//时间函数 从**此时**开始计算过期
	TimeFuc func() time.Time
	//私钥
	PrivateKey string
	//key
	Key string
	//save cookie
	SendCookie bool
	CookieName string
	//存活时间
	CookieMaxAge   int
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
}
type JwtResponse struct {
	Token        string
	RefreshToken string
}

func (j *JwtHandler) LoginHandler(ctx *zygo.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	//A部分
	signingMethod := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(signingMethod)
	//B部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for k, v := range data {
			claims[k] = v
		}
	}
	//设置起始时间
	if j.TimeFuc == nil {
		j.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFuc().Add(j.TimeOut)
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFuc().Unix()

	//C部分
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		//需要私钥
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}

	jr := &JwtResponse{
		Token: tokenString,
	}
	//refresh token
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	//发送到存储cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = "my_token"
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFuc().Unix())
		}
		ctx.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "HS256", "HS384", "HS512":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	//B部分
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFuc().Add(j.TimeOut).Unix()
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		//需要私钥
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenString, nil
}
