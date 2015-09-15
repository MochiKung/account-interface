package verify

import (
	"reflect"
	"strings"

	"github.com/MochiKung/account-interface/encrypt"
	"github.com/MochiKung/account-interface/handler/oauth2"
	"github.com/MochiKung/account-interface/handler/oauth2/database/client"
	"github.com/MochiKung/account-interface/handler/oauth2/database/user"
)

func Init() {
}

func VerifyClientPassword(clientInfo *client.ClientInfo, password string) bool {
	enteredEncryptPassword := encrypt.EncryptText1Way([]byte(password), clientInfo.Salt)
	if !reflect.DeepEqual(clientInfo.EncryptedPassword, enteredEncryptPassword) {
		return false
	}
	return true
}

func VerifyUserPassword(userInfo *user.UserInfo, password string) bool {
	enteredEncryptPassword := encrypt.EncryptText1Way([]byte(password), userInfo.Salt)
	if !reflect.DeepEqual(userInfo.EncryptedPassword, enteredEncryptPassword) {
		return false
	}
	return true
}

func VerifyGrantScopes(clientInfo *client.ClientInfo, grantType string, scopes string) bool {
	var clientScope map[string]bool
	switch grantType {
	case oauth2.ResourceOwnerCredentialsGrant:
		clientScope = clientInfo.GrantResourceOwner
	case oauth2.ClientCredentialsGrant:
		clientScope = clientInfo.GrantClientCredentials
	}

	scopeSlice := strings.Split(scopes, ",")
	for _, scope := range scopeSlice {
		if !clientScope[scope] {
			return false
		}
	}
	return true
}
