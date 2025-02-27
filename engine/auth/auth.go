package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/deluan/navidrome/consts"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

var (
	once      sync.Once
	JwtSecret []byte
	TokenAuth *jwtauth.JWTAuth
)

func InitTokenAuth(ds model.DataStore) {
	once.Do(func() {
		secret, err := ds.Property(nil).DefaultGet(consts.JWTSecretKey, "not so secret")
		if err != nil {
			log.Error("No JWT secret found in DB. Setting a temp one, but please report this error", err)
		}
		JwtSecret = []byte(secret)
		TokenAuth = jwtauth.New("HS256", JwtSecret, nil)
	})
}

func CreateToken(u *model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = consts.JWTIssuer
	claims["sub"] = u.UserName
	claims["adm"] = u.IsAdmin

	return TouchToken(token)
}

func TouchToken(token *jwt.Token) (string, error) {
	expireIn := time.Now().Add(consts.JWTTokenExpiration).Unix()
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expireIn

	return token.SignedString(JwtSecret)
}

func Validate(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return JwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), err
}
