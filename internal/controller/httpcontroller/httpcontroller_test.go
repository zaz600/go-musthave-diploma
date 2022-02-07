package httpcontroller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	. "github.com/zaz600/go-musthave-diploma/api"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

const (
	sessionCookieName    = "GM_LS_SESSION"
	orderProcessedStatus = "PROCESSED"
)

func NewUser() RegisterRequest {
	return RegisterRequest{
		Login:    random.String(5),
		Password: random.String(6),
	}
}

func newRouter(t *testing.T) *chi.Mux {
	t.Helper()

	mu := &sync.Mutex{}
	orderStates := map[string]int{}

	accrualMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/orders/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		orderID := strings.Replace(r.URL.Path, "/api/orders/", "", 1)

		mu.Lock()
		defer mu.Unlock()
		orderStates[orderID]++

		var accrual float32 = 50.0
		status := Accrual.ResponseStatusPROCESSED
		if orderID == "999999998" {
			status = Accrual.ResponseStatusINVALID
		}

		if orderID == "888888880" {
			switch orderStates[orderID] {
			case 1:
				status = Accrual.ResponseStatusREGISTERED
				accrual = 0
			case 2, 3:
				status = Accrual.ResponseStatusPROCESSING
				accrual = 0
			default:
				status = Accrual.ResponseStatusPROCESSED
			}
		}

		resp := Accrual.Response{
			Accrual: &accrual,
			Order:   Accrual.Order(orderID),
			Status:  status,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))

	accrualClient, err := Accrual.NewClientWithResponses(accrualMockServer.URL)
	require.NoError(t, err)
	service := gophermartservice.New(accrualClient,
		gophermartservice.WithStorage(gophermartservice.Memory),
		gophermartservice.WithAccrualRetryInterval(20*time.Millisecond),
	)
	return httpcontroller.NewRouter(service)
}

func TestGophermartController_UserRegister(t *testing.T) {
	t.Run("success registration", func(t *testing.T) {
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		e.POST("/api/user/login").
			WithJSON(user).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("wrong password", func(t *testing.T) {
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		uploadOrder(t, e, "92345678905")

		// второй юзер
		user2 := NewUser()
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
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.POST("/api/user/orders").
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("invalid order", func(t *testing.T) {
		user := NewUser()
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

//nolint:funlen
func TestGophermartController_GetUserOrders(t *testing.T) {
	t.Run("success get user orders", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		uploadOrder(t, e, "92345678905")
		uploadOrder(t, e, "12345678903")
		uploadOrder(t, e, "346436439")

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			orders := e.GET("/api/user/orders").
				Expect().
				Status(http.StatusOK).
				ContentType("application/json").
				JSON().
				Array()
			t.Logf("orders resp: %#v", orders.Raw())

			g.Expect(orders.Length().Raw()).To(Equal(float64(3)))
			g.Expect(orders.Element(0).Object().Value("status").String().Raw()).Should(Equal(orderProcessedStatus))
			g.Expect(orders.Element(1).Object().Value("status").String().Raw()).Should(Equal(orderProcessedStatus))
			g.Expect(orders.Element(2).Object().Value("status").String().Raw()).Should(Equal(orderProcessedStatus))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
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
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)

		e.GET("/api/user/orders").
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("accrual invalid", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		uploadOrder(t, e, "999999998") // invalid in mock

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			order := e.GET("/api/user/orders").
				Expect().
				Status(http.StatusOK).JSON().
				Array().
				Element(0).
				Object()
			t.Logf("orders resp: %#v", order)

			g.Expect(order.Value("status").String().Raw()).Should(Equal("INVALID"))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())

		assertBalance(t, e, 0.0, 0.0)
	})

	t.Run("accrual processing", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		register(t, e, user)
		// Этот ордер в моке сервиса accrual проходит все статусы обработки, при запросе начислений
		uploadOrder(t, e, "888888880") // REGISTERED -> PROCESSING -> PROCESSING -> PROCESSED in mock

		// Ждем, что ордер дойдет до конца обработки и баллы будут начислены
		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			order := e.GET("/api/user/orders").
				Expect().
				Status(http.StatusOK).JSON().
				Array().
				Element(0).
				Object()
			t.Logf("orders resp: %#v", order.Raw())

			g.Expect(order.Value("status").String().Raw()).Should(Equal(orderProcessedStatus))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())

		assertBalance(t, e, 50.0, 0.0)
	})
}

func TestGophermartController_GetUserBalance(t *testing.T) {
	t.Run("balance changed after upload order", func(t *testing.T) {
		user := NewUser()
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

		assertBalance(t, e, 100.0, 0)
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
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
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
		user := NewUser()
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

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			withdrawals := e.GET("/api/user/balance/withdrawals").
				Expect().
				Status(http.StatusOK).
				JSON().
				Array()

			t.Logf("withdrawals resp: %#v", withdrawals)

			g.Expect(withdrawals.Length().Raw()).Should(Equal(float64(1)))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
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
		user := NewUser()
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
	user := NewUser()
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

	g := NewGomegaWithT(t)
	g.Eventually(func(g Gomega) {
		orders := e.GET("/api/user/orders").
			Expect().
			Status(http.StatusOK).
			ContentType("application/json").
			JSON().
			Array()

		t.Logf("orders resp: %#v", orders)

		g.Expect(orders.Length().Raw()).Should(Equal(float64(1)))
		g.Expect(orders.Element(0).Object().Value("status").Raw()).Should(Equal(orderProcessedStatus))
	}, 2*time.Second, 10*time.Millisecond).Should(Succeed())

	// баланс поменялся
	assertBalance(t, e, 50.0, 0.0)

	// Запрашиваем списание
	e.POST("/api/user/balance/withdraw").
		WithText(`{"order": "92345678905", "sum": 10.50}`).
		Expect().
		Status(http.StatusOK)

	// баланс изменился
	assertBalance(t, e, 50.0-10.5, 10.5)

	g.Eventually(func(g Gomega) {
		withdrawals := e.GET("/api/user/balance/withdrawals").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		t.Logf("withdrawals resp: %#v", withdrawals)

		g.Expect(withdrawals.Length().Raw()).Should(Equal(float64(1)))
	}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
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
	g := NewGomegaWithT(t)
	g.Eventually(func(g Gomega) {
		balance := e.GET("/api/user/balance").Expect().Status(http.StatusOK).JSON().Object()
		t.Logf("balance response: %#v", balance.Raw())

		g.Expect(balance.Value("current").Number().Raw()).Should(Equal(float64(current)))
		g.Expect(balance.Value("withdrawn").Number().Raw()).Should(Equal(float64(withdrawn)))
	}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
}
