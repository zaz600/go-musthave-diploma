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
	"github.com/stretchr/testify/suite"
	. "github.com/zaz600/go-musthave-diploma/api"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
)

const (
	orderProcessedStatus = "PROCESSED"
)

type HTTPControllerTestSuite struct {
	suite.Suite
	server *httptest.Server
	user   RegisterRequest
}

func (suite *HTTPControllerTestSuite) SetupTest() {
	suite.server = httptest.NewServer(newRouter(suite.T()))
	suite.user = NewUser()
}

func (suite *HTTPControllerTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *HTTPControllerTestSuite) TestRegistration_Success() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	e.POST("/api/user/register").
		WithJSON(suite.user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Headers().ContainsKey("Authorization").
		NotEmpty()
}

func (suite *HTTPControllerTestSuite) TestRegistration_BadRequest() {
	e := httpexpect.New(suite.T(), suite.server.URL)
	e.POST("/api/user/register").
		Expect().
		Status(http.StatusBadRequest)
}

func (suite *HTTPControllerTestSuite) TestRegistration_Duplicate() {
	e := httpexpect.New(suite.T(), suite.server.URL)
	token := register(suite.T(), e, suite.user)

	e.POST("/api/user/register").
		WithJSON(suite.user).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusConflict)
}

func (suite *HTTPControllerTestSuite) TestLogin_Success() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	// регистрируемся
	register(suite.T(), e, suite.user)

	// логинимся зарегистрированным
	e.POST("/api/user/login").
		WithJSON(suite.user).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		Headers().ContainsKey("Authorization").
		NotEmpty()
}

func (suite *HTTPControllerTestSuite) TestLogin_BadRequest() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	e.POST("/api/user/login").
		Expect().
		Status(http.StatusBadRequest)
}

func (suite *HTTPControllerTestSuite) TestLogin_UnknownUser() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	user := NewUser()

	e.POST("/api/user/login").
		WithJSON(user).
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestLogin_WrongPassword() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	// регистрируемся
	register(suite.T(), e, suite.user)

	e.POST("/api/user/login").
		WithJSON(LoginRequest{Login: suite.user.Login, Password: "1111111111"}).
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_Success() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	orderID := random.OrderID()
	token := register(suite.T(), e, suite.user)

	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText(orderID).
		Expect().
		Status(http.StatusAccepted)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_UserUploadSameOrder() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	orderID := random.OrderID()
	token := register(suite.T(), e, suite.user)

	uploadOrder(suite.T(), e, orderID, token)

	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText(orderID).
		Expect().
		Status(http.StatusOK)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_AnotherUserUploadSameOrder() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	orderID := random.OrderID()
	token := register(suite.T(), e, suite.user)

	uploadOrder(suite.T(), e, orderID, token)

	// второй юзер
	user2 := NewUser()
	e2 := httpexpect.New(suite.T(), suite.server.URL)
	token2 := register(suite.T(), e2, user2)

	e2.POST("/api/user/orders").
		WithHeader("Authorization", token2).
		WithText(orderID).
		Expect().
		Status(http.StatusConflict)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_NotAuthorized() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	e.POST("/api/user/orders").
		Expect().
		Status(http.StatusUnauthorized)

	e.POST("/api/user/orders").
		WithHeader("Authorization", "1234").
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_BadRequest() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	token := register(suite.T(), e, suite.user)

	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusBadRequest)
}

func (suite *HTTPControllerTestSuite) TestUploadOrder_InvalidOrder() {
	e := httpexpect.New(suite.T(), suite.server.URL)

	token := register(suite.T(), e, suite.user)

	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText("12345").
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (suite *HTTPControllerTestSuite) TestGetUserOrders_Success() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	uploadOrder(t, e, random.OrderID(), token)
	uploadOrder(t, e, random.OrderID(), token)
	uploadOrder(t, e, random.OrderID(), token)

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
}

func (suite *HTTPControllerTestSuite) TestGetUserOrders_NotAuthorized() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	e.GET("/api/user/orders").
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestGetUserOrders_NoOrders() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	e.GET("/api/user/orders").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusNoContent)
}

func (suite *HTTPControllerTestSuite) TestGetUserOrders_AccrualInvalid() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)
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
}

// TestGetUserOrders_AccrualProcessed проверяет, что заказ после его загрузки дошел до финального статуса
// Для заказов, номер которых начинается на 88888, мок меняет статус при каждом запроса статуса.
// REGISTERED -> PROCESSING -> PROCESSING -> PROCESSED
func (suite *HTTPControllerTestSuite) TestGetUserOrders_AccrualProcessed() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)
	// Этот ордер в моке сервиса accrual проходит все статусы обработки, при запросе начислений
	uploadOrder(t, e, goluhn.GenerateWithPrefix("88888", 15), token)

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
}

func (suite *HTTPControllerTestSuite) TestGetUserBalance_BalanceChangedAfterUploadOrder() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	balance := e.GET("/api/user/balance").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	balance.Value("current").Equal(0.0)
	balance.Value("withdrawn").Equal(0.0)

	uploadOrder(t, e, random.OrderID(), token)
	uploadOrder(t, e, random.OrderID(), token)

	assertBalance(t, e, 100.0, 0, token)
}

func (suite *HTTPControllerTestSuite) TestGetUserBalance_NotAuthorized() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	e.GET("/api/user/balance").
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdraw_Success() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)
	uploadOrder(t, e, random.OrderID(), token)
	uploadOrder(t, e, random.OrderID(), token)
	assertBalance(t, e, 100.0, 0, token)

	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, random.OrderID())).
		Expect().
		Status(http.StatusOK)

	assertBalance(t, e, 100.0-10.5, 10.5, token)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdraw_NotAuthorized() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	e.POST("/api/user/balance/withdraw").
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdraw_BadRequest() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusBadRequest)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdraw_PaymentRequired() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 751.21}`, random.OrderID())).
		Expect().
		Status(http.StatusPaymentRequired)

	uploadOrder(t, e, random.OrderID(), token)
	uploadOrder(t, e, random.OrderID(), token)
	assertBalance(t, e, 100.0, 0, token)

	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 751.21}`, random.OrderID())).
		Expect().
		Status(http.StatusPaymentRequired)
	assertBalance(t, e, 100.0, 0, token)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdrawals_Success() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	uploadOrder(t, e, random.OrderID(), token)
	assertBalance(t, e, 50.0, 0, token)

	e.POST("/api/user/balance/withdraw").
		WithHeader("Authorization", token).
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, random.OrderID())).
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
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdrawals_NotAuthorized() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	e.GET("/api/user/balance/withdrawals").
		Expect().
		Status(http.StatusUnauthorized)
}

func (suite *HTTPControllerTestSuite) TestUserBalanceWithdrawals_NoWithdrawals() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	token := register(t, e, suite.user)

	e.GET("/api/user/balance/withdrawals").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusNoContent)
}

func (suite *HTTPControllerTestSuite) TestSuccessPath() {
	t := suite.T()
	e := httpexpect.New(t, suite.server.URL)

	// Регистрируемся
	token := register(t, e, suite.user)

	// Загружаем заказ
	e.POST("/api/user/orders").
		WithHeader("Authorization", token).
		WithText(random.OrderID()).
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
		WithText(fmt.Sprintf(`{"order": "%s", "sum": 10.50}`, random.OrderID())).
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

func TestGophermartControllerSuite(t *testing.T) {
	suite.Run(t, new(HTTPControllerTestSuite))
}

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
