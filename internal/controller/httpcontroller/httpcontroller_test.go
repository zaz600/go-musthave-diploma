package httpcontroller_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	. "github.com/zaz600/go-musthave-diploma/api"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
)

const (
	orderProcessedStatus = "PROCESSED"
)

func NewUser() RegisterRequest {
	return RegisterRequest{
		Login:    random.String(5),
		Password: random.String(6),
	}
}

func newAccrualMock() *httptest.Server {
	mu := &sync.Mutex{}
	orderStates := map[string]int{}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if strings.HasPrefix(orderID, "99999") {
			status = Accrual.ResponseStatusINVALID
		}

		if strings.HasPrefix(orderID, "88888") {
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
}

func newRouter(t *testing.T) *chi.Mux {
	t.Helper()

	accrualClient, err := Accrual.NewClientWithResponses(newAccrualMock().URL)
	require.NoError(t, err)

	options := []gophermartservice.Option{gophermartservice.WithAccrualRetryInterval(20 * time.Millisecond)}

	if os.Getenv("TEST_PG") != "" {
		db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/gophermart")
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = db.Close()
		})
		options = append(options, gophermartservice.WithPgStorage(db))
	} else {
		options = append(options, gophermartservice.WithMemoryStorage())
	}

	service := gophermartservice.New(accrualClient, options...)
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
			Headers().ContainsKey("Authorization").
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

		token := register(t, e, user)

		e.POST("/api/user/register").
			WithJSON(user).
			WithHeader("Authorization", token).
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
			Headers().ContainsKey("Authorization").
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

		orderID := goluhn.Generate(15)
		token := register(t, e, user)

		e.POST("/api/user/orders").
			WithHeader("Authorization", token).
			WithText(orderID).
			Expect().
			Status(http.StatusAccepted)
	})

	t.Run("user upload same order", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		orderID := goluhn.Generate(15)
		token := register(t, e, user)
		uploadOrder(t, e, orderID, token)

		e.POST("/api/user/orders").
			WithHeader("Authorization", token).
			WithText(orderID).
			Expect().
			Status(http.StatusOK)
	})

	t.Run("another user upload same order", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		orderID := goluhn.Generate(15)
		token := register(t, e, user)
		uploadOrder(t, e, orderID, token)

		// второй юзер
		user2 := NewUser()
		e2 := httpexpect.New(t, server.URL)
		token2 := register(t, e2, user2)

		e2.POST("/api/user/orders").
			WithHeader("Authorization", token2).
			WithText(orderID).
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
			WithHeader("Authorization", "1234").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("bad request", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)

		e.POST("/api/user/orders").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("invalid order", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)

		e.POST("/api/user/orders").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusBadRequest)

		e.POST("/api/user/orders").
			WithHeader("Authorization", token).
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

		token := register(t, e, user)

		uploadOrder(t, e, goluhn.Generate(15), token)
		uploadOrder(t, e, goluhn.Generate(15), token)
		uploadOrder(t, e, goluhn.Generate(15), token)

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			orders := e.GET("/api/user/orders").
				WithHeader("Authorization", token).
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

		token := register(t, e, user)

		e.GET("/api/user/orders").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("accrual invalid", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)
		uploadOrder(t, e, goluhn.GenerateWithPrefix("99999", 15), token) // invalid in mock

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			order := e.GET("/api/user/orders").
				WithHeader("Authorization", token).
				Expect().
				Status(http.StatusOK).JSON().
				Array().
				Element(0).
				Object()
			t.Logf("orders resp: %#v", order)

			g.Expect(order.Value("status").String().Raw()).Should(Equal("INVALID"))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())

		assertBalance(t, e, 0.0, 0.0, token)
	})

	t.Run("accrual processing", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)
		// Этот ордер в моке сервиса accrual проходит все статусы обработки, при запросе начислений
		uploadOrder(t, e, goluhn.GenerateWithPrefix("88888", 15), token) // REGISTERED -> PROCESSING -> PROCESSING -> PROCESSED in mock

		// Ждем, что ордер дойдет до конца обработки и баллы будут начислены
		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			order := e.GET("/api/user/orders").
				WithHeader("Authorization", token).
				Expect().
				Status(http.StatusOK).JSON().
				Array().
				Element(0).
				Object()
			t.Logf("orders resp: %#v", order.Raw())

			g.Expect(order.Value("status").String().Raw()).Should(Equal(orderProcessedStatus))
		}, 2*time.Second, 10*time.Millisecond).Should(Succeed())

		assertBalance(t, e, 50.0, 0.0, token)
	})
}

func TestGophermartController_GetUserBalance(t *testing.T) {
	t.Run("balance changed after upload order", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)

		balance := e.GET("/api/user/balance").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		balance.Value("current").Equal(0.0)
		balance.Value("withdrawn").Equal(0.0)

		uploadOrder(t, e, goluhn.Generate(15), token)
		uploadOrder(t, e, goluhn.Generate(15), token)

		assertBalance(t, e, 100.0, 0, token)
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

		token := register(t, e, user)
		uploadOrder(t, e, goluhn.Generate(15), token)
		uploadOrder(t, e, goluhn.Generate(15), token)
		assertBalance(t, e, 100.0, 0, token)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, goluhn.Generate(15))).
			Expect().
			Status(http.StatusOK)

		assertBalance(t, e, 100.0-10.5, 10.5, token)
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

		token := register(t, e, user)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusBadRequest)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			WithText(`{"order": "92345678905", "sum": 751.21}`).
			Expect().
			Status(http.StatusPaymentRequired)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			WithText(fmt.Sprintf(`{"order": "%s", "sum": 751.21}`, goluhn.Generate(15))).
			Expect().
			Status(http.StatusPaymentRequired)

		uploadOrder(t, e, goluhn.Generate(15), token)
		uploadOrder(t, e, goluhn.Generate(15), token)
		assertBalance(t, e, 100.0, 0, token)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			WithText(fmt.Sprintf(`{"order": "%s", "sum": 751.21}`, goluhn.Generate(15))).
			Expect().
			Status(http.StatusPaymentRequired)
		assertBalance(t, e, 100.0, 0, token)
	})
}

func TestGophermartController_UserBalanceWithdrawals(t *testing.T) {
	t.Run("success get user withdrawals", func(t *testing.T) {
		user := NewUser()
		server := httptest.NewServer(newRouter(t))
		defer server.Close()
		e := httpexpect.New(t, server.URL)

		token := register(t, e, user)

		uploadOrder(t, e, goluhn.Generate(15), token)
		assertBalance(t, e, 50.0, 0, token)

		e.POST("/api/user/balance/withdraw").
			WithHeader("Authorization", token).
			WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, goluhn.Generate(15))).
			Expect().
			Status(http.StatusOK)

		g := NewGomegaWithT(t)
		g.Eventually(func(g Gomega) {
			withdrawals := e.GET("/api/user/balance/withdrawals").
				WithHeader("Authorization", token).
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

		token := register(t, e, user)

		e.GET("/api/user/balance/withdrawals").
			WithHeader("Authorization", token).
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
	token := register(t, e, user)

	// Загружаем заказ
	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText(goluhn.Generate(15)).
		Expect().
		Status(http.StatusAccepted)

	g := NewGomegaWithT(t)
	g.Eventually(func(g Gomega) {
		orders := e.GET("/api/user/orders").
			WithHeader("Authorization", token).
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
	assertBalance(t, e, 50.0, 0.0, token)

	// Запрашиваем списание
	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, goluhn.Generate(15))).
		Expect().
		Status(http.StatusOK)

	// баланс изменился
	assertBalance(t, e, 50.0-10.5, 10.5, token)

	g.Eventually(func(g Gomega) {
		withdrawals := e.GET("/api/user/balance/withdrawals").
			WithHeader("Authorization", token).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		t.Logf("withdrawals resp: %#v", withdrawals)

		g.Expect(withdrawals.Length().Raw()).Should(Equal(float64(1)))
	}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
}

func register(t *testing.T, e *httpexpect.Expect, user RegisterRequest) string {
	t.Helper()

	authHeader := e.POST("/api/user/register").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Header("Authorization").NotEmpty().Raw()
	return authHeader
}

func uploadOrder(t *testing.T, e *httpexpect.Expect, orderID string, token string) {
	t.Helper()

	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText(orderID).
		Expect().
		Status(http.StatusAccepted)
}

func assertBalance(t *testing.T, e *httpexpect.Expect, current float32, withdrawn float32, token string) {
	g := NewGomegaWithT(t)
	g.Eventually(func(g Gomega) {
		balance := e.GET("/api/user/balance").
			WithHeader("Authorization", token).
			Expect().Status(http.StatusOK).JSON().Object()
		t.Logf("balance response: %#v", balance.Raw())

		g.Expect(balance.Value("current").Number().Raw()).Should(Equal(float64(current)))
		g.Expect(balance.Value("withdrawn").Number().Raw()).Should(Equal(float64(withdrawn)))
	}, 2*time.Second, 10*time.Millisecond).Should(Succeed())
}
