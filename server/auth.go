package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ag-computational-bio/BioDataDBModels/go/client"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
)

//AuthHandler Basic for performing authentication
type AuthHandler struct {
	Oauth2Conf        *oauth2.Config
	Oauth2StateString string
}

// Init Initializes the auth handler object
func (handler *AuthHandler) Init() {
	clientID := viper.GetString("Auth.ClientID")
	callbackURL := viper.GetString("Auth.CallbackURL")
	AuthURL := viper.GetString("Auth.AuthURL")
	TokenURL := viper.GetString("Auth.TokenURL")

	oauth2Conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: os.Getenv("Oauth2ClientSecret"),
		RedirectURL:  callbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
		Scopes: []string{"profile", "email"},
	}
	handler.Oauth2Conf = oauth2Conf

	//TODO change
	handler.Oauth2StateString = "test"
}

// Callback callback handling for oauth2 login
func (handler *AuthHandler) Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	token, err := handler.GetAccessToken(state, code)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
	}

	marshalledToken, err := json.Marshal(token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
	}

	tokenString := base64.StdEncoding.EncodeToString(marshalledToken)

	c.SetCookie("token", tokenString, 15*60*60, "/", "", true, true)

	c.Redirect(http.StatusMovedPermanently, "/index")
	c.Abort()
}

// GetAccessToken Returns the token of a login request
func (handler *AuthHandler) GetAccessToken(state string, code string) (*oauth2.Token, error) {
	if state != handler.Oauth2StateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := handler.Oauth2Conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

// Auth Used to authenticate call and login
func (handler *AuthHandler) Auth(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, handler.Oauth2Conf.AuthCodeURL(handler.Oauth2StateString, oauth2.AccessTypeOffline))
	c.Abort()
}

//UpdateToken Updates the token if available
func (handler *AuthHandler) UpdateToken(c *gin.Context) {
	if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/auth/callback" {
		c.Next()
	}

	rawTokenCookie, err := c.Request.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	token := decodeToken(rawTokenCookie, c)

	tokensource := handler.Oauth2Conf.TokenSource(context.TODO(), token)
	updatedToken, err := tokensource.Token()
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
	}

	updatedRawToken, err := json.Marshal(updatedToken)
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	rawTokenCookie.Value = string(updatedRawToken)

	c.Next()
}

// OutGoingContextFromToken Creates the required outgoing context for a call
func (handler *AuthHandler) OutGoingContextFromToken(token string, tokentype client.TokenType) context.Context {
	mdMap := make(map[string]string)
	mdMap[string(tokentype)] = token
	tokenMetadata := metadata.New(mdMap)

	outgoingContext := metadata.NewOutgoingContext(context.TODO(), tokenMetadata)
	return outgoingContext
}

// GetAccessTokenFromGinContext Returns the access token of a gin context from cookie
// Token needs to be stored as "token"
func (handler *AuthHandler) GetAccessTokenFromGinContext(c *gin.Context) string {
	rawTokenCookie, err := c.Request.Cookie("token")
	if err != nil && err != http.ErrNoCookie {
		log.Println(err.Error())
		c.AbortWithError(400, err)
	}

	if err == http.ErrNoCookie {
		log.Println("cookie not found")
		return ""
	}

	unescapedBase64Data, err := url.QueryUnescape(rawTokenCookie.Value)
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	rawBytesDecoded, err := base64.StdEncoding.DecodeString(unescapedBase64Data)
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	var token oauth2.Token

	err = json.Unmarshal([]byte(rawBytesDecoded), &token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
	}

	return token.AccessToken
}

func decodeToken(rawTokenCookie *http.Cookie, c *gin.Context) *oauth2.Token {
	unescapedBase64Data, err := url.QueryUnescape(rawTokenCookie.Value)
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	rawBytesDecoded, err := base64.StdEncoding.DecodeString(unescapedBase64Data)
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	var token oauth2.Token
	err = json.Unmarshal(rawBytesDecoded, &token)
	if err != nil && err != http.ErrNoCookie {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/index")
		c.AbortWithError(400, err)
	}

	return &token
}
