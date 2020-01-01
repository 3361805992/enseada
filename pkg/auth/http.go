package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/enseadaio/enseada/internal/auth"
	authv1beta1api "github.com/enseadaio/enseada/internal/auth/v1beta1"
	"github.com/enseadaio/enseada/internal/middleware"
	"github.com/enseadaio/enseada/internal/utils"
	authv1beta1 "github.com/enseadaio/enseada/rpc/auth/v1beta1"
	session "github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/random"
	"github.com/ory/fosite"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func mountRoutes(e *echo.Echo, s *auth.Store, op fosite.OAuth2Provider, enf *casbin.Enforcer, sm echo.MiddlewareFunc) {
	g := e.Group("/oauth")
	g.Use(sm)
	g.GET("/authorize", authorizationPage())
	g.POST("/authorize", authorize(op, s))
	g.POST("/token", token(op, s))
	g.POST("/token/introspect", introspect(op))

	acl := &authv1beta1api.ACLService{
		Logger:   s.Logger,
		Enforcer: enf,
	}

	aclhandler := authv1beta1.NewAclAPIServer(acl, middleware.AuthTwirpHooks(s.Logger, s, op))
	h := echo.WrapHandler(middleware.WithAuthorizationStrategy(aclhandler))
	e.Any(aclhandler.PathPrefix()+"*", h)
}

func authorizationPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		s := session.Default(c)
		e := s.Flashes("errors")
		params := echo.Map{
			"ClientID":     utils.QueryWithDefault(c, "client_id", ""),
			"RedirectURI":  utils.QueryWithDefault(c, "redirect_uri", ""),
			"State":        utils.QueryWithDefault(c, "state", random.String(32)),
			"Scope":        utils.QueryWithDefault(c, "scope", ""),
			"Audience":     utils.QueryWithDefault(c, "audience", ""),
			"ResponseType": utils.QueryWithDefault(c, "response_type", "code"),
		}
		if len(e) > 0 {
			params["Errors"] = e
		}

		return c.Render(http.StatusOK, "login", params)
	}
}

func authorize(oauth fosite.OAuth2Provider, store *auth.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		resw := c.Response()
		ctx := req.Context()

		ar, err := oauth.NewAuthorizeRequest(ctx, req)
		if err != nil {
			c.Logger().Error(err)
			oauth.WriteAuthorizeError(resw, ar, err)
			return nil
		}

		for _, scope := range ar.GetRequestedScopes() {
			ar.GrantScope(scope)
		}

		username := strings.TrimSpace(req.FormValue("username"))
		password := strings.TrimSpace(req.FormValue("password"))

		err = store.Authenticate(ctx, username, password)
		if err != nil {
			if strings.Contains(req.Header.Get("accept"), "html") {
				s := session.Default(c)
				s.AddFlash("Invalid username of password", "errors")
				if err := s.Save(); err != nil {
					return err
				}
				return c.Redirect(http.StatusSeeOther, c.Request().Header.Get("Referer"))
			}
			oauth.WriteAuthorizeError(resw, ar, fosite.ErrAccessDenied)
			return nil
		}

		u, err := store.FindUserByUsername(ctx, username)
		if err != nil {
			return err
		}

		os := auth.NewSession(u)
		res, err := oauth.NewAuthorizeResponse(ctx, ar, os)
		if err != nil {
			c.Logger().Error(err)
			oauth.WriteAuthorizeError(resw, ar, err)
			return nil
		}

		oauth.WriteAuthorizeResponse(resw, ar, res)
		return nil
	}
}

func token(oauth fosite.OAuth2Provider, store *auth.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		resw := c.Response()
		ctx := req.Context()

		os := auth.NewSession(nil)
		c.Logger().Debug(req)
		ar, err := oauth.NewAccessRequest(ctx, req, os)
		if err != nil {
			rfce := errors.Cause(err).(*fosite.RFC6749Error)
			if strings.Contains(rfce.Debug, "password") {
				c.Logger().Error("authentication failed")
				oauth.WriteAccessError(resw, ar, fosite.ErrAccessDenied)
				return nil

			}
			c.Logger().Error(err)
			oauth.WriteAccessError(resw, ar, err)
			return nil
		}

		// If this is a client_credentials grant, grant all scopes the client is allowed to perform.
		if ar.GetGrantTypes().Exact("client_credentials") {
			for _, scope := range ar.GetRequestedScopes() {
				if fosite.HierarchicScopeStrategy(ar.GetClient().GetScopes(), scope) {
					ar.GrantScope(scope)
				}
			}
		}

		// If this is a password grant, populate the session.
		if ar.GetGrantTypes().Exact("password") {
			username := strings.TrimSpace(req.FormValue("username"))
			u, err := store.FindUserByUsername(ctx, username)
			if err != nil {
				return err
			}

			ar.SetSession(auth.NewSession(u))
		}

		for _, scope := range ar.GetRequestedScopes() {
			ar.GrantScope(scope)
		}

		res, err := oauth.NewAccessResponse(ctx, ar)
		if err != nil {
			c.Logger().Error(err)
			oauth.WriteAccessError(resw, ar, err)
			return nil
		}

		oauth.WriteAccessResponse(resw, ar, res)
		return nil
	}
}

func introspect(oauth fosite.OAuth2Provider) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		req := c.Request()
		resw := c.Response().Writer
		os := auth.NewSession(nil)

		ir, err := oauth.NewIntrospectionRequest(ctx, req, os)
		if err != nil {
			c.Logger().Error(err)
			oauth.WriteIntrospectionError(resw, err)
			return nil
		}

		oauth.WriteIntrospectionResponse(resw, ir)
		return nil
	}
}
