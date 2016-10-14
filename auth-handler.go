package apperix

import (
	"fmt"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/crypto/bcrypt"
)

func authReadHandler(client *Client, request *Request, service *Service) Response {
	response := ResponseJson {}
	if request.Parameters["username"] == nil ||
		//missing username
		len(request.Parameters["username"]) < 1 ||
		len(request.Parameters["username"][0]) < 1 {
		response.ReplyClientError("NO_USERNAME", "Missing username argument")
		return &response
	}
	if request.Parameters["password"] == nil ||
		//missing password
		len(request.Parameters["password"]) < 1 ||
		len(request.Parameters["password"][0]) < 1 {
		response.ReplyClientError("NO_PASSWORD", "Missing password argument")
		return &response
	}
	username := request.Parameters["username"][0]
	password := request.Parameters["password"][0]
	account, err := service.FindUserByUsername(username)
	if err != nil {
		//wrong username
		response.ReplyForbidden("Wrong username or password")
		return &response
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		//wrong password
		response.ReplyForbidden("Wrong username or password")
		return &response
	}
	//generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": account.Identifier.String(),
		"iat": time.Now().UTC().Format("2006-01-02T15:04:05-0700"),
		"lft": service.Config.AccessTokenLiveTime().Seconds(),
	})
	tokenString, err := token.SignedString(service.Config.JwtSignatureSecret())
	if err != nil {
		panic(fmt.Errorf("Could not sign token: %s", err))
	}
	response.Data("access-token", tokenString)
	response.Data("life-time", service.Config.AccessTokenLiveTime().Seconds())
	return &response
}