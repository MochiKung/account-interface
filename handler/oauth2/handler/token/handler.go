package token

import (
	"log"
	"net/http"
	"time"

	"github.com/MochiKung/account-interface/handler/oauth2"
	"github.com/MochiKung/account-interface/handler/oauth2/database/access-token"
	"github.com/MochiKung/account-interface/handler/oauth2/database/client"
	"github.com/MochiKung/account-interface/handler/oauth2/database/user"
	"github.com/MochiKung/account-interface/handler/oauth2/handler/token/response-writer"
	"github.com/MochiKung/account-interface/handler/oauth2/string-generator"
	"github.com/MochiKung/account-interface/handler/oauth2/verify"
)

const (
	PrefixPath = oauth2.PrefixPath + "/token"
	expiresIn  = 3600
)

var ()

func init() {
}

type Handler struct {
}

func New() *Handler {
	self := &Handler{}
	return self
}

func (self *Handler) ServeHTTP(httpRes http.ResponseWriter, req *http.Request) {
	resp := response.NewResponseWriter(httpRes)
	if req.Method != "POST" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	for _, value := range req.Form {
		if len(value) > 1 {
			resp.WriteError(&response.InvalidRequestError, "request parameters must not be included more than once")
			return
		}
	}
	grantType := req.Form.Get("grant_type")
	switch grantType {
	case oauth2.ResourceOwnerCredentialsGrant:
		serveResourceOwnerCredentials(resp, req)
	case oauth2.ClientCredentialsGrant:
		serveClientCredentials(resp, req)
	default:
		resp.WriteError(&response.UnsupportedGrantTypeError, "")
	}
}

func serveClientCredentials(resp *response.ResponseWriter, req *http.Request) {
	grantType := req.Form.Get("grant_type")
	scopes := req.Form.Get("scope")
	username, password, ok := req.BasicAuth()
	if !ok {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	clientInfo, err := client.GetClientInfo(username)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if clientInfo == nil {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	// verify password
	if !verify.VerifyClientPassword(clientInfo, password) {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	// verify client grant
	if clientInfo.GrantClientCredentials == nil {
		resp.WriteError(&response.UnauthorizedClientError, "")
		return
	}

	// verify client grant scope
	if !verify.VerifyGrantScopes(clientInfo, grantType, scopes) {
		resp.WriteError(&response.InvalidScopeError, "")
		return
	}

	// generate new token
	token := stringgenerator.RandomString(8)
	tokenInfo := &accesstoken.TokenInfo{}
	tokenInfo.Token = token
	tokenInfo.Client = username
	tokenInfo.User = ""
	tokenInfo.Scopes = scopes
	expireTime := time.Now().Add(time.Duration(expiresIn) * time.Second)
	tokenInfo.ExpireTime = &expireTime
	for err := accesstoken.PutTokenInfo(tokenInfo); err != nil && err.Error() == "duplicate token"; {
		tokenInfo.Token = stringgenerator.RandomString(8)
		err = accesstoken.PutTokenInfo(tokenInfo)
	}
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	resp.WriteSuccess(tokenInfo.Token, expiresIn, "", scopes)
}

func serveResourceOwnerCredentials(resp *response.ResponseWriter, req *http.Request) {
	grantType := req.Form.Get("grant_type")
	username := req.Form.Get("username")
	password := req.Form.Get("password")
	scopes := req.Form.Get("scope")
	clientUsername, clientPassword, ok := req.BasicAuth()
	if !ok {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	clientInfo, err := client.GetClientInfo(clientUsername)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if clientInfo == nil {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	if username == "" || password == "" {
		resp.WriteError(&response.InvalidGrantError, "")
		return
	}

	userInfo, err := user.GetUserInfo(username)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userInfo == nil {
		resp.WriteError(&response.InvalidGrantError, "")
		return
	}

	// verify client password
	if !verify.VerifyClientPassword(clientInfo, clientPassword) {
		resp.WriteError(&response.InvalidClientError, "")
		return
	}

	// verify resource owner credential
	if !verify.VerifyUserPassword(userInfo, password) {
		resp.WriteError(&response.InvalidGrantError, "")
		return
	}

	// verify client grant
	if clientInfo.GrantResourceOwner == nil {
		resp.WriteError(&response.UnauthorizedClientError, "")
		return
	}

	// verify client grant scope
	if !verify.VerifyGrantScopes(clientInfo, grantType, scopes) {
		resp.WriteError(&response.InvalidScopeError, "")
		return
	}

	// generate new token
	token := stringgenerator.RandomString(8)
	tokenInfo := &accesstoken.TokenInfo{}
	tokenInfo.Token = token
	tokenInfo.Client = clientUsername
	tokenInfo.User = username
	tokenInfo.Scopes = scopes
	expireTime := time.Now().Add(time.Duration(expiresIn) * time.Second)
	tokenInfo.ExpireTime = &expireTime
	for err := accesstoken.PutTokenInfo(tokenInfo); err != nil && err.Error() == "duplicate token"; {
		tokenInfo.Token = stringgenerator.RandomString(8)
		err = accesstoken.PutTokenInfo(tokenInfo)
	}
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	resp.WriteSuccess(tokenInfo.Token, expiresIn, "", scopes)
}
