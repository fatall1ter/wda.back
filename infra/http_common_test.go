package infra

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.countmax.ru/countmax/wda.back/internal/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func TestServer_getSessionID(t *testing.T) {
	s := &Server{log: zap.NewNop().Sugar()}
	tests := []struct {
		name  string
		c     echo.Context
		want  string
		want1 session.TokenSource
	}{
		{"bearer", makeEchoContext("token", false), "token", session.FromBearer},
		{"cookie", makeEchoContext("tokenpoken", true), "tokenpoken", session.FromCookie},
		{"bearer_complex", makeEchoContext("BearerTokenBearer", false), "BearerTokenBearer", session.FromBearer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := s.getSessionID(tt.c)
			if got != tt.want {
				t.Errorf("Server.getSessionID().SessionID = %s, want %s", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Server.getSessionID().SessionSource = %d, want %d", got1, tt.want1)
			}
		})
	}
}

func makeEchoContext(token string, inCookie bool) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	//
	if inCookie {
		cookie := &http.Cookie{
			Name:     oryKratosSessCookieID,
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		}
		req.AddCookie(cookie)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return e.NewContext(req, rec)
}
