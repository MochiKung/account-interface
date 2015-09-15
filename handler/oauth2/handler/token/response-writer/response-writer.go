package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
}

type errorResponse struct {
	ErrorTag         string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	HttpStatus       int    `json:"-"`
}

type successResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

func NewResponseWriter(resp http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{resp}
}

func (self *ResponseWriter) WriteSuccess(accessToken string, expiresIn int, refreshToken string, scopes string) {
	resp := &successResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		Scope:        scopes,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		self.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	} else {
		self.Header().Set("Content-Type", "application/json")
		self.Header().Set("Cache-Control", "no-store")
		self.Header().Set("Pragma", "no-cache")
		self.WriteHeader(http.StatusOK)
		self.Write(data)
	}
}

func (self *ResponseWriter) WriteError(resp *errorResponse, description string) {
	var data []byte
	var err error

	if description == "" {
		data, err = json.Marshal(resp)
	} else {
		data, err = json.Marshal(&errorResponse{
			ErrorTag:         resp.ErrorTag,
			ErrorDescription: description,
			ErrorURI:         resp.ErrorURI,
		})
	}
	if err != nil {
		self.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	} else {
		self.Header().Set("Content-Type", "application/json")
		self.Header().Set("Cache-Control", "no-store")
		self.Header().Set("Pragma", "no-cache")
		self.WriteHeader(resp.HttpStatus)
		self.Write(data)
	}
}

var InvalidRequestError errorResponse = errorResponse{
	ErrorTag:   "invalid_request",
	HttpStatus: http.StatusBadRequest,
}

var InvalidClientError errorResponse = errorResponse{
	ErrorTag:         "invalid_client",
	ErrorDescription: "client authentication failed",
	HttpStatus:       http.StatusUnauthorized,
}

var InvalidGrantError errorResponse = errorResponse{
	ErrorTag:         "invalid_grant",
	ErrorDescription: "the provided authorization grant or refresh token is invalid",
	HttpStatus:       http.StatusBadRequest,
}

var UnauthorizedClientError errorResponse = errorResponse{
	ErrorTag:         "unauthorized_client",
	ErrorDescription: "the authenticated client is not authorized to use this authorization grant type",
	HttpStatus:       http.StatusBadRequest,
}

var UnsupportedGrantTypeError errorResponse = errorResponse{
	ErrorTag:         "unsupported_grant_type",
	ErrorDescription: "the authorization grant type is not supported by the authorization server",
	HttpStatus:       http.StatusBadRequest,
}

var InvalidScopeError errorResponse = errorResponse{
	ErrorTag:         "invalid_scope",
	ErrorDescription: "the request scope is invalid",
	HttpStatus:       http.StatusBadRequest,
}
