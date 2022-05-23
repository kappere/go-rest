// 文档https://go-oauth2.github.io/zh/
// oauth2错误码一览
// var StatusCodes = map[error]int{
// 	ErrInvalidRequest:          400,
// 	ErrUnauthorizedClient:      401,
// 	ErrAccessDenied:            403,
// 	ErrUnsupportedResponseType: 401,
// 	ErrInvalidScope:            400,
// 	ErrServerError:             500,
// 	ErrTemporarilyUnavailable:  503,
// 	ErrInvalidClient:           401,
// 	ErrInvalidGrant:            401,
// 	ErrUnsupportedGrantType:    401,
// }
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-oauth2/oauth2/server"
	"github.com/kappere/go-rest/core/db"
	"github.com/kappere/go-rest/core/rest"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gorm.io/gorm"
)

const (
	SQL_SELECT_CLIENT = "select client_id as id, client_secret as secret from oauth_client_details "
	SQL_SELECT_TOKEN  = "select "
	TABLE_CLIENT      = "oauth_client_details"
	TABLE_TOKEN       = "oauth_access_token"
)

type DbClientStore struct {
	Db *gorm.DB
}

func (store *DbClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	var client OauthClientDetails
	store.Db.Table(TABLE_CLIENT).Where("id = ?", id).Take(&client)
	if client.ID == "" {
		return nil, nil
	}
	return &client, nil
}

type DbTokenStore struct {
	Db         *gorm.DB
	TokenMutex sync.Mutex
}

// create and store the new token information
func (store *DbTokenStore) Create(info oauth2.TokenInfo) error {
	store.TokenMutex.Lock()
	defer store.TokenMutex.Unlock()
	var token OauthAccessToken
	r := store.Db.Table(TABLE_TOKEN).Where("client_id = ?", info.GetClientID()).Take(&token)
	if r.Error == nil {
		store.Db.Table(TABLE_TOKEN).Where("client_id = ?", info.GetClientID()).Update("access", info.GetAccess())
	} else if r.Error != nil && r.Error.Error() == "record not found" {
		oauthToken := newOauthAccessToken(info)
		r := store.Db.Table(TABLE_TOKEN).Create(&oauthToken)
		if r.Error != nil {
			return r.Error
		}
	} else {
		return r.Error
	}
	return nil
}

// delete the authorization code
func (store *DbTokenStore) RemoveByCode(code string) error {
	store.Db.Table(TABLE_TOKEN).Where("code = ?", code).Delete(&OauthAccessToken{})
	return nil
}

// use the access token to delete the token information
func (store *DbTokenStore) RemoveByAccess(access string) error {
	store.Db.Table(TABLE_TOKEN).Where("access = ?", access).Delete(&OauthAccessToken{})
	return nil
}

// use the refresh token to delete the token information
func (store *DbTokenStore) RemoveByRefresh(refresh string) error {
	store.Db.Table(TABLE_TOKEN).Where("refresh = ?", refresh).Delete(&OauthAccessToken{})
	return nil
}

// use the authorization code for token information data
func (store *DbTokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	var token OauthAccessToken
	r := store.Db.Table(TABLE_TOKEN).Where("code = ?", code).Take(&token)
	if r.Error != nil {
		return nil, nil
	}
	return newModelsToken(&token), nil
}

// use the access token for token information data
func (store *DbTokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {
	var token OauthAccessToken
	r := store.Db.Table(TABLE_TOKEN).Where("access = ?", access).Take(&token)
	if r.Error != nil {
		return nil, nil
	}
	return newModelsToken(&token), nil
}

// use the refresh token for token information data
func (store *DbTokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) { return nil, nil }

func OAuth2ClientTokenMiddleware(oauth2Conf *rest.OAuth2Config, engine *rest.Engine) rest.HandlerFunc {
	if db.Db == nil {
		panic("database not inititialized")
	}
	manager := manage.NewManager()
	// client接口
	manager.MapClientStorage(&DbClientStore{Db: db.Db})
	// access_token生成
	manager.MapAccessGenerate(generates.NewAccessGenerate())
	// access_token存储
	manager.MapTokenStorage(&DbTokenStore{Db: db.Db, TokenMutex: sync.Mutex{}})
	cfg := &manage.Config{
		// 访问令牌过期时间（默认为2小时）
		AccessTokenExp: time.Duration(oauth2Conf.Expire * int(time.Second)),
		// RefreshTokenExp:   time.Duration(oauth2Conf.Expire * int(time.Second)),
		// IsGenerateRefresh: true,
	}
	manager.SetClientTokenCfg(cfg)

	srv := server.NewServer(server.NewConfig(), manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	engine.GET(oauth2Conf.TokenUri, func(c *rest.Context) {
		grantType, tokenGenerateRequest, err := srv.ValidationTokenRequest(c.Request)
		if err != nil {
			errorToken(c, srv, err)
			c.Abort()
			return
		}

		tokenInfo, err := srv.GetAccessToken(grantType, tokenGenerateRequest)
		if err != nil {
			errorToken(c, srv, err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, rest.Success(srv.GetTokenData(tokenInfo)))
	})
	return func(c *rest.Context) {
		err := srv.HandleAuthorizeRequest(c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusOK, rest.ErrorWithCode(err.Error(), rest.STATUS_NO_AUTHORIZATION))
			c.Abort()
			return
		}
		c.Next()
	}
}

func errorToken(c *rest.Context, srv *server.Server, err error) {
	errData, statusCode, _ := srv.GetErrorData(err)
	c.JSON(http.StatusOK, rest.ErrorWithCode(errData["error"].(string)+":"+errData["error_description"].(string), statusCode))
}

// OauthClientDetails client model
type OauthClientDetails struct {
	ID     string
	Secret string
	Domain string
	UserID string
}

// GetID client id
func (c *OauthClientDetails) GetID() string {
	return c.ID
}

// GetSecret client domain
func (c *OauthClientDetails) GetSecret() string {
	return c.Secret
}

// GetDomain client domain
func (c *OauthClientDetails) GetDomain() string {
	return c.Domain
}

// GetUserID user id
func (c *OauthClientDetails) GetUserID() string {
	return c.UserID
}

// OauthAccessToken token model
type OauthAccessToken struct {
	ClientID         *string
	UserID           *string
	RedirectURI      *string
	Scope            *string
	Code             *string
	CodeCreateAt     *time.Time
	CodeExpiresIn    *time.Duration
	Access           *string
	AccessCreateAt   *time.Time
	AccessExpiresIn  *time.Duration
	Refresh          *string
	RefreshCreateAt  *time.Time
	RefreshExpiresIn *time.Duration
}

func newOauthAccessToken(info oauth2.TokenInfo) *OauthAccessToken {
	ClientID := info.GetClientID()
	UserID := info.GetUserID()
	RedirectURI := info.GetRedirectURI()
	Scope := info.GetScope()
	Code := info.GetCode()
	CodeCreateAt := info.GetCodeCreateAt()
	CodeExpiresIn := info.GetCodeExpiresIn()
	Access := info.GetAccess()
	AccessCreateAt := info.GetAccessCreateAt()
	AccessExpiresIn := info.GetAccessExpiresIn()
	Refresh := info.GetRefresh()
	RefreshCreateAt := info.GetRefreshCreateAt()
	RefreshExpiresIn := info.GetRefreshExpiresIn()

	var PCodeCreateAt *time.Time = nil
	if CodeCreateAt.Nanosecond() > 0 {
		PCodeCreateAt = &CodeCreateAt
	}

	var PAccessCreateAt *time.Time = nil
	if AccessCreateAt.Nanosecond() > 0 {
		PAccessCreateAt = &AccessCreateAt
	}

	var PRefreshCreateAt *time.Time = nil
	if RefreshCreateAt.Nanosecond() > 0 {
		PRefreshCreateAt = &RefreshCreateAt
	}
	return &OauthAccessToken{
		ClientID:         &ClientID,
		UserID:           &UserID,
		RedirectURI:      &RedirectURI,
		Scope:            &Scope,
		Code:             &Code,
		CodeCreateAt:     PCodeCreateAt,
		CodeExpiresIn:    &CodeExpiresIn,
		Access:           &Access,
		AccessCreateAt:   PAccessCreateAt,
		AccessExpiresIn:  &AccessExpiresIn,
		Refresh:          &Refresh,
		RefreshCreateAt:  PRefreshCreateAt,
		RefreshExpiresIn: &RefreshExpiresIn,
	}
}

func newModelsToken(token *OauthAccessToken) oauth2.TokenInfo {
	var CodeCreateAt time.Time
	if token.CodeCreateAt != nil {
		CodeCreateAt = *token.CodeCreateAt
	}
	var AccessCreateAt time.Time
	if token.AccessCreateAt != nil {
		AccessCreateAt = *token.AccessCreateAt
	}
	var RefreshCreateAt time.Time
	if token.RefreshCreateAt != nil {
		RefreshCreateAt = *token.RefreshCreateAt
	}
	return &models.Token{
		ClientID:         *token.ClientID,
		UserID:           *token.UserID,
		RedirectURI:      *token.RedirectURI,
		Scope:            *token.Scope,
		Code:             *token.Code,
		CodeCreateAt:     CodeCreateAt,
		CodeExpiresIn:    *token.CodeExpiresIn,
		Access:           *token.Access,
		AccessCreateAt:   AccessCreateAt,
		AccessExpiresIn:  *token.AccessExpiresIn,
		Refresh:          *token.Refresh,
		RefreshCreateAt:  RefreshCreateAt,
		RefreshExpiresIn: *token.RefreshExpiresIn,
	}
}
