package server

import (
	"net/http"

	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) check(c echo.Context) error {
	s.logger.Infof("Server GET check %s", s.urlWithBase(c))

	request := c.Request()
	ctx := request.Context()

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := api.CheckAuthorization(ctx, request.Method, s.urlWithBase(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kid := authRes.KID

	if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}
