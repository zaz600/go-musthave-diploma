package httpcontroller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/zaz600/go-musthave-diploma/api"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
)

const (
	sessionCookieName    = "GM_LS_SESSION"
	orderProcessedStatus = "PROCESSED"
)

func newRouter(t *testing.T) *chi.Mux {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var accrual float32 = 50.0
		resp := Accrual.Response{
			Accrual: &accrual,
			Order:   "1",
			Status:  Accrual.ResponseStatusPROCESSED,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))

	accrualClient, err := Accrual.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	return httpcontroller.NewRouter(gophermartservice.NewWithMemStorage(accrualClient))
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

	assert.Eventually(t, func() bool {
		json := e.GET("/api/user/orders").
			Expect().
			Status(http.StatusOK).
			ContentType("application/json").
			JSON()

		count := json.Array().Length().Raw()
		order1Status := json.Array().Element(0).Object().Value("status").String().Raw()
		order2Status := json.Array().Element(1).Object().Value("status").String().Raw()
		order3Status := json.Array().Element(2).Object().Value("status").String().Raw()

		return count == 3 && order1Status == orderProcessedStatus && order2Status == orderProcessedStatus && order3Status == orderProcessedStatus
	}, 2*time.Second, 100*time.Millisecond)
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

	uploadOrder(t, e, "92345678905")
	uploadOrder(t, e, "346436439")

	assert.Eventually(t, func() bool {
		json = e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		json.Value("current").Equal(100.0)
		json.Value("withdrawn").Equal(0.0)
		return true
	}, 2*time.Second, 100*time.Millisecond)
}

func TestGophermartController_UserBalanceWithdraw(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	e.POST("/api/user/balance/withdraw").
		Expect().
		Status(http.StatusUnauthorized)

	register(t, e, user)

	e.POST("/api/user/balance/withdraw").
		Expect().
		Status(http.StatusBadRequest)

	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 751.21}`).
		Expect().
		Status(http.StatusPaymentRequired)

	uploadOrder(t, e, "92345678905")
	uploadOrder(t, e, "346436439")

	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 751.21}`).
		Expect().
		Status(http.StatusPaymentRequired)

	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 10.50}`).
		Expect().
		Status(http.StatusOK)

	assert.Eventually(t, func() bool {
		json := e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		json.Value("current").Equal(100.0 - 10.5)
		json.Value("withdrawn").Equal(10.5)
		return true
	}, 2*time.Second, 100*time.Millisecond)
}

func TestGophermartController_UserBalanceWithdrawals(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	e.GET("/api/user/balance/withdrawals").
		Expect().
		Status(http.StatusUnauthorized)

	register(t, e, user)

	e.GET("/api/user/balance/withdrawals").
		Expect().
		Status(http.StatusNoContent)

	uploadOrder(t, e, "92345678905")
	assert.Eventually(t, func() bool {
		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 10.50}`).
			Expect().
			Status(http.StatusOK)
		return true
	}, 2*time.Second, 100*time.Millisecond)

	assert.Eventually(t, func() bool {
		e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			Length().Equal(1)

		return true
	}, 2*time.Second, 100*time.Millisecond)
}

//nolint:funlen
func TestGophermart_SuccessPath(t *testing.T) {
	user := RegisterRequest{
		Login:    "foouser",
		Password: "pass",
	}

	server := httptest.NewServer(newRouter(t))
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	// Регистрируемся
	e.POST("/api/user/register").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Cookie(sessionCookieName).
		Value().
		NotEmpty()

	// Загружаем заказ
	e.POST("/api/user/orders").
		WithText("92345678905").
		Expect().
		Status(http.StatusAccepted)

	// заказ обработан
	assert.Eventually(t, func() bool {
		json := e.GET("/api/user/orders").
			Expect().
			Status(http.StatusOK).
			ContentType("application/json").
			JSON()

		json.Array().Length().Equal(1)
		json.Array().Element(0).Object().Value("status").Equal(orderProcessedStatus)

		return true
	}, 2*time.Second, 100*time.Millisecond)

	// баланс поменялся
	assert.Eventually(t, func() bool {
		json := e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		json.Value("current").Equal(50.0)
		json.Value("withdrawn").Equal(0.0)
		return true
	}, 2*time.Second, 100*time.Millisecond)

	// Запрашиваем списание
	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 10.50}`).
		Expect().
		Status(http.StatusOK)

	// баланс изменился
	assert.Eventually(t, func() bool {
		json := e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		json.Value("current").Equal(50.0 - 10.5)
		json.Value("withdrawn").Equal(10.5)
		return true
	}, 2*time.Second, 100*time.Millisecond)

	// отображается одно списание
	assert.Eventually(t, func() bool {
		e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			Length().Equal(1)
		return true
	}, 2*time.Second, 100*time.Millisecond)
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
