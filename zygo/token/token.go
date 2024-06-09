package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
	"web/zygo"
)

const JWTTOKEN = "my_token"

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
	PrivateKey []byte
	//key
	Key        []byte
	RefreshKey string
	//save cookie
	SendCookie bool
	CookieName string
	//存活时间
	CookieMaxAge   int
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool

	//获取认证字段
	Header string
	//认证错误处理
	AuthHandler func(ctx *zygo.Context)
}
type JwtResponse struct {
	Token        string
	RefreshToken string
}

//登录  用户认证（用户名密码） -> 用户id 将id生成jwt，并且保存到cookie或者进行返回

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
			j.CookieName = JWTTOKEN
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
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	//B部分
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFuc().Add(j.RefreshTimeOut).Unix()
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

// 退出登陆
func (j *JwtHandler) LogoutHandler(ctx *zygo.Context) error {
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTTOKEN
		}
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		return nil
	}
	return nil
}

// 刷新token
func (j *JwtHandler) RefreshHandler(ctx *zygo.Context) (*JwtResponse, error) {
	rToken, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	//解析token

	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if j.usingPublicKeyAlgo() {
			return j.PrivateKey, nil
		} else {
			return j.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}

	//B部分
	claims := t.Claims.(jwt.MapClaims)
	//未过期的情况下 重新生成token和refreshToken
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
		tokenString, tokenErr = t.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}

	jr := &JwtResponse{
		Token: tokenString,
	}
	//refresh token
	refreshToken, err := j.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	//发送到存储cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTTOKEN
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFuc().Unix())
		}
		ctx.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

// jwt登陆拦截器
func (j *JwtHandler) AuthIntserception(next zygo.HandlerFunc) zygo.HandlerFunc {
	return func(ctx *zygo.Context) {

	}
}
