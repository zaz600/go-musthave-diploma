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

var user = RegisterRequest{
	Login:    "foouser",
	Password: "pass",
}

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
	t.Run("success registration", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/register").
			WithJSON(user).
			Expect().
			Status(http.StatusOK).
			ContentType("application/json").
			Cookie(sessionCookieName).
			Value().
			NotEmpty()
	})

	t.Run("no request body", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/register").
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/register").
			WithJSON(user).
			Expect().
			Status(http.StatusConflict)
	})
}

func TestGophermartController_UserLogin(t *testing.T) {
	t.Run("success login", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

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
	})

	t.Run("no request body", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/login").
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("unknown user", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/login").
			WithJSON(user).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("wrong password", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		// регистрируемся
		register(t, e, RegisterRequest(user))

		user.Password = "1111111111"
		e.POST("/api/user/login").
			WithJSON(user).
			Expect().
			Status(http.StatusUnauthorized)
	})
}

//nolint:funlen
func TestGophermartController_UploadOrder(t *testing.T) {
	t.Run("success upload order", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/orders").
			WithText("92345678905").
			Expect().
			Status(http.StatusAccepted)
	})

	t.Run("user upload same order", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		uploadOrder(t, e, "92345678905")

		e.POST("/api/user/orders").
			WithText("92345678905").
			Expect().
			Status(http.StatusOK)
	})

	t.Run("another user upload same order", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		uploadOrder(t, e, "92345678905")

		// второй юзер
		user2 := RegisterRequest{
			Login:    "baruser",
			Password: "pass",
		}
		e2 := httpexpect.New(t, server.URL)
		register(t, e2, user2)

		e2.POST("/api/user/orders").
			WithText("92345678905").
			Expect().
			Status(http.StatusConflict)
	})

	t.Run("not authorized", func(t *testing.T) {
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
	})

	t.Run("bad request", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/orders").
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("invalid order", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/orders").
			Expect().
			Status(http.StatusBadRequest)

		e.POST("/api/user/orders").
			WithText("12345").
			Expect().
			Status(http.StatusUnprocessableEntity)
	})
}

func TestGophermartController_GetUserOrders(t *testing.T) {
	t.Run("success get user orders", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		uploadOrder(t, e, "92345678905")
		uploadOrder(t, e, "12345678903")
		uploadOrder(t, e, "346436439")

		assert.Eventually(t, func() bool {
			orders := e.GET("/api/user/orders").
				Expect().
				Status(http.StatusOK).
				ContentType("application/json").
				JSON()

			orders.Array().Length().Equal(3)
			orders.Array().Element(0).Object().Value("status").Equal(orderProcessedStatus)
			orders.Array().Element(1).Object().Value("status").Equal(orderProcessedStatus)
			orders.Array().Element(2).Object().Value("status").Equal(orderProcessedStatus)
			return true
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("not authorized", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.GET("/api/user/orders").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("no orders", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.GET("/api/user/orders").
			Expect().
			Status(http.StatusNoContent)
	})
}

func TestGophermartController_GetUserBalance(t *testing.T) {
	t.Run("balance changed after upload order", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		balance := e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		balance.Value("current").Equal(0.0)
		balance.Value("withdrawn").Equal(0.0)

		uploadOrder(t, e, "92345678905")
		uploadOrder(t, e, "346436439")

		assert.Eventually(t, func() bool {
			balance := e.GET("/api/user/balance").
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()

			balance.Value("current").Equal(100.0)
			balance.Value("withdrawn").Equal(0.0)
			return true
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("not authorized", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.GET("/api/user/balance").
			Expect().
			Status(http.StatusUnauthorized)
	})
}

//nolint:funlen
func TestGophermartController_UserBalanceWithdraw(t *testing.T) {
	t.Run("success balance withdraw", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		uploadOrder(t, e, "92345678905")
		uploadOrder(t, e, "346436439")
		assertBalance(t, e, 100.0, 0)

		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 10.50}`).
			Expect().
			Status(http.StatusOK)

		assertBalance(t, e, 100.0-10.5, 10.5)
	})

	t.Run("not authorized", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/balance/withdraw").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("invalid request", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/balance/withdraw").
			Expect().
			Status(http.StatusBadRequest)

		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 751.21}`).
			Expect().
			Status(http.StatusPaymentRequired)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 751.21}`).
			Expect().
			Status(http.StatusPaymentRequired)

		uploadOrder(t, e, "92345678905")
		uploadOrder(t, e, "346436439")
		assertBalance(t, e, 100.0, 0)

		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 751.21}`).
			Expect().
			Status(http.StatusPaymentRequired)
		assertBalance(t, e, 100.0, 0)
	})
}

func TestGophermartController_UserBalanceWithdrawals(t *testing.T) {
	t.Run("success get user withdrawals", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		uploadOrder(t, e, "92345678905")
		assertBalance(t, e, 50.0, 0)

		e.POST("/api/user/balance/withdraw").
			WithText(`{"order": "92345678905", "sum": 10.50}`).
			Expect().
			Status(http.StatusOK)

		assert.Eventually(t, func() bool {
			e.GET("/api/user/balance/withdrawals").
				Expect().
				Status(http.StatusOK).
				JSON().
				Array().
				Length().Equal(1)

			return true
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("not authorized", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("no withdrawals", func(t *testing.T) {
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusNoContent)
	})
}

//nolint:funlen
func TestGophermart_SuccessPath(t *testing.T) {
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
	}, 2*time.Second, 10*time.Millisecond)

	// баланс поменялся
	assertBalance(t, e, 50.0, 0.0)

	// Запрашиваем списание
	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 10.50}`).
		Expect().
		Status(http.StatusOK)

	// баланс изменился
	assertBalance(t, e, 50.0-10.5, 10.5)

	// отображается одно списание
	assert.Eventually(t, func() bool {
		e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			Length().Equal(1)
		return true
	}, 2*time.Second, 10*time.Millisecond)
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

func assertBalance(t *testing.T, e *httpexpect.Expect, current float32, withdrawn float32) {
	assert.Eventually(t, func() bool {
		balance := e.GET("/api/user/balance").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		balance.Value("current").Equal(current)
		balance.Value("withdrawn").Equal(withdrawn)
		return true
	}, 2*time.Second, 10*time.Millisecond)
}
