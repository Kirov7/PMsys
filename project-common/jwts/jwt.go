package jwts

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JwtToken struct {
	AccessToken  string
	RefreshToken string
	AccessExp    int64
	RefreshExp   int64
}

func CreateToken(val, AcSecret, rfSecret string, exp, rfExp time.Duration, ip string) *JwtToken {
	aExp := time.Now().Add(exp).Unix()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
		"ip":    ip,
	})
	token, _ := claims.SignedString([]byte(AcSecret))

	rExp := time.Now().Add(rfExp).Unix()
	refreshClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"key": val,
		"exp": rExp,
	})
	refreshToken, _ := refreshClaims.SignedString([]byte(rfSecret))

	return &JwtToken{
		AccessToken:  token,
		RefreshToken: refreshToken,
		AccessExp:    aExp,
		RefreshExp:   rExp,
	}
}

func ParseToken(tokenString string, secret string, ip string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		token := claims["token"].(string)
		exp := int64(claims["exp"].(float64))
		if exp <= time.Now().Unix() {
			return "", errors.New("token已过期")
		}
		if claims["ip"] != ip {
			fmt.Println("传入ip", ip)
			fmt.Println("token ip", claims[ip])
			return "", errors.New("ip不合法")
		}
		return token, nil

	} else {
		return "", err
	}
}