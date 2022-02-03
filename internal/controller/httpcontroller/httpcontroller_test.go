package httpcontroller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	. "github.com/zaz600/go-musthave-diploma/api"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
)

const sessionCookieName = "GM_LS_SESSION"

func newRouter(t *testing.T) *chi.Mux {
	t.Helper()
	userService := userservice.NewService(userrepository.NewInmemoryUserRepository())
	sessionService := sessionservice.NewService(sessionrepository.NewInmemorySessionRepository())
	orderService := orderservice.NewService(orderrepository.NewInmemoryOrderRepository())
	return httpcontroller.NewRouter(userService, sessionService, orderService)
}

func TestGophermartController_UserRegister(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	// no request body
	e.POST("/api/user/register").
		Expect().
		Status(http.StatusBadRequest)

	// correct registration
	e.POST("/api/user/register").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Cookie(sessionCookieName).
		Value().
		NotEmpty()

	// повторная рега с тем же логином
	e.POST("/api/user/register").
		WithJSON(user).
		Expect().
		Status(http.StatusConflict)
}

func TestGophermartController_UserLogin(t *testing.T) {
	user := LoginRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	// no request body
	e.POST("/api/user/login").
		Expect().
		Status(http.StatusBadRequest)

	// not registered
	e.POST("/api/user/login").
		WithJSON(user).
		Expect().
		Status(http.StatusUnauthorized)

	// регистрируемся
	register(t, e, RegisterRequest(user))

	// логинимся зарегистрированным
	e.POST("/api/user/login").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Cookie(sessionCookieName).
		Value().
		NotEmpty()

	// wrong password
	user.Password = "1111111111"
	e.POST("/api/user/login").
		WithJSON(user).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestGophermartController_UploadOrder(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	e.POST("/api/user/orders").
		Expect().
		Status(http.StatusUnauthorized)

	e.POST("/api/user/orders").
		WithCookie(sessionCookieName, "123456").
		Expect().
		Status(http.StatusUnauthorized)

	register(t, e, user)

	e.POST("/api/user/orders").
		Expect().
		Status(http.StatusBadRequest)

	e.POST("/api/user/orders").
		WithText("12345").
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/user/orders").
		WithText("12345").
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/user/orders").
		WithText("92345678905").
		Expect().
		Status(http.StatusAccepted)

	// повторная отправка
	e.POST("/api/user/orders").
		WithText("92345678905").
		Expect().
		Status(http.StatusOK)

	// второй юзер
	user.Login += "2"
	e2 := httpexpect.New(t, server.URL)
	register(t, e2, user)

	e2.POST("/api/user/orders").
		WithText("92345678905").
		Expect().
		Status(http.StatusConflict)
}

func TestGophermartController_GetUserOrders(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	e.GET("/api/user/orders").
		Expect().
		Status(http.StatusUnauthorized)

	register(t, e, user)

	e.GET("/api/user/orders").
		Expect().
		Status(http.StatusNoContent)

	uploadOrder(t, e, "92345678905")
	uploadOrder(t, e, "12345678903")
	uploadOrder(t, e, "346436439")

	json := e.GET("/api/user/orders").
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		JSON()

	json.Array().Length().Equal(3)
	json.Array().Element(0).Object().Value("number").Equal("92345678905")
	json.Array().Element(1).Object().Value("number").Equal("12345678903")
	json.Array().Element(2).Object().Value("number").Equal("346436439")
}

func TestGophermartController_GetUserBalance(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	e.GET("/api/user/balance").
		Expect().
		Status(http.StatusUnauthorized)

	register(t, e, user)

	json := e.GET("/api/user/balance").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	json.Value("current").Equal(0.0)
	json.Value("withdrawn").Equal(0.0)
	// TODO
}

func register(t *testing.T, e *httpexpect.Expect, user RegisterRequest) {
	t.Helper()

	e.POST("/api/user/register").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Cookie(sessionCookieName).
		Value().
		NotEmpty()
}

func uploadOrder(t *testing.T, e *httpexpect.Expect, orderID string) {
	t.Helper()

	e.POST("/api/user/orders").
		WithText(orderID).
		Expect().
		Status(http.StatusAccepted)
}
