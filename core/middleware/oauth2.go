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

type DbClientStore struct {
	Db *gorm.DB
}

func (store *DbClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	var client OauthClientDetails
	store.Db.Where("id = ?", id).Take(&client)
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
	r := store.Db.Where("client_id = ?", info.GetClientID()).Take(&token)
	if r.Error == nil && r.RowsAffected >= 1 {
		now := time.Now()
		store.Db.Model(&OauthAccessToken{}).
			Where("client_id = ?", info.GetClientID()).
			Updates(OauthAccessToken{
				Access:         info.GetAccess(),
				AccessCreateAt: &now,
			})
	} else {
		oauthToken := newOauthAccessToken(info)
		r := store.Db.Create(&oauthToken)
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

// delete the authorization code
func (store *DbTokenStore) RemoveByCode(code string) error {
	r := store.Db.Where("code = ?", code).Delete(&OauthAccessToken{})
	return r.Error
}

// use the access token to delete the token information
func (store *DbTokenStore) RemoveByAccess(access string) error {
	r := store.Db.Where("access = ?", access).Delete(&OauthAccessToken{})
	return r.Error
}

// use the refresh token to delete the token information
func (store *DbTokenStore) RemoveByRefresh(refresh string) error {
	r := store.Db.Where("refresh = ?", refresh).Delete(&OauthAccessToken{})
	return r.Error
}

// use the authorization code for token information data
func (store *DbTokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	var token OauthAccessToken
	r := store.Db.Where("code = ?", code).Take(&token)
	if r.Error != nil {
		return nil, nil
	}
	return newModelsToken(&token), nil
}

// use the access token for token information data
func (store *DbTokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {
	var token OauthAccessToken
	r := store.Db.Where("access = ?", access).Take(&token)
	if r.Error != nil {
		return nil, nil
	}
	return newModelsToken(&token), nil
}

// use the refresh token for token information data
func (store *DbTokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	var token OauthAccessToken
	r := store.Db.Where("refresh = ?", refresh).Take(&token)
	if r.Error != nil {
		return nil, nil
	}
	return newModelsToken(&token), nil
}

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
		grantType, tgr, err := srv.ValidationTokenRequest(c.Request)
		if err != nil {
			errorToken(c, srv, err)
			c.Abort()
			return
		}

		client, err := srv.Manager.GetClient(tgr.ClientID)
		if err == nil {
			tgr.UserID = client.GetUserID()
		}

		tokenInfo, err := srv.GetAccessToken(grantType, tgr)
		if err != nil {
			errorToken(c, srv, err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, rest.Success(srv.GetTokenData(tokenInfo)))
	})
	return func(c *rest.Context) {
		tokenInfo, err := srv.ValidationBearerToken(c.Request)
		if err != nil {
			c.JSON(http.StatusOK, rest.ErrorWithCode(err.Error(), rest.STATUS_NO_AUTHORIZATION))
			c.Abort()
			return
		}
		// add oauth info to context
		c.Set("oauth_client_id", tokenInfo.GetClientID())
		c.Set("oauth_user_id", tokenInfo.GetUserID())
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

func (OauthClientDetails) TableName() string {
	return "oauth_client_details"
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
	ClientID         string
	UserID           string
	RedirectURI      string
	Scope            string
	Code             string
	CodeCreateAt     *time.Time
	CodeExpiresIn    time.Duration
	Access           string
	AccessCreateAt   *time.Time
	AccessExpiresIn  time.Duration
	Refresh          string
	RefreshCreateAt  *time.Time
	RefreshExpiresIn time.Duration
}

func (OauthAccessToken) TableName() string {
	return "oauth_access_token"
}

func newOauthAccessToken(info oauth2.TokenInfo) *OauthAccessToken {
	CodeCreateAt := info.GetCodeCreateAt()
	AccessCreateAt := info.GetAccessCreateAt()
	RefreshCreateAt := info.GetRefreshCreateAt()

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
		ClientID:         info.GetClientID(),
		UserID:           info.GetUserID(),
		RedirectURI:      info.GetRedirectURI(),
		Scope:            info.GetScope(),
		Code:             info.GetCode(),
		CodeCreateAt:     PCodeCreateAt,
		CodeExpiresIn:    info.GetCodeExpiresIn() / time.Second,
		Access:           info.GetAccess(),
		AccessCreateAt:   PAccessCreateAt,
		AccessExpiresIn:  info.GetAccessExpiresIn() / time.Second,
		Refresh:          info.GetRefresh(),
		RefreshCreateAt:  PRefreshCreateAt,
		RefreshExpiresIn: info.GetRefreshExpiresIn() / time.Second,
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
		ClientID:         token.ClientID,
		UserID:           token.UserID,
		RedirectURI:      token.RedirectURI,
		Scope:            token.Scope,
		Code:             token.Code,
		CodeCreateAt:     CodeCreateAt,
		CodeExpiresIn:    token.CodeExpiresIn * time.Second,
		Access:           token.Access,
		AccessCreateAt:   AccessCreateAt,
		AccessExpiresIn:  token.AccessExpiresIn * time.Second,
		Refresh:          token.Refresh,
		RefreshCreateAt:  RefreshCreateAt,
		RefreshExpiresIn: token.RefreshExpiresIn * time.Second,
	}
}
