package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/driver"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/repository"
	"github.com/muktiarafi/ticketing-orders/internal/service"
	"github.com/ory/dockertest/v3"
)

var (
	pool     *dockertest.Pool
	resource *dockertest.Resource
)

var router *echo.Echo
var ticketRepo repository.TicketRepository

func TestMain(m *testing.M) {
	db := &driver.DB{
		SQL: newTestDatabase(),
	}

	router = echo.New()
	router.Use(middleware.Logger())

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	router.Validator = customValidator
	router.HTTPErrorHandler = common.CustomErrorHandler

	ticketRepo = repository.NewTicketRepository(db)
	orderRepository := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepository, ticketRepo)

	orderPublisher := &OrderPublisherStub{}
	orderHandler := NewOrderHandler(orderService, orderPublisher)
	orderHandler.Route(router)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func newTestDatabase() *sql.DB {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err = pool.Run("postgres", "alpine", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=postgres"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB
	if err = pool.Retry(func() error {
		db, err = sql.Open(
			"pgx",
			fmt.Sprintf("host=localhost port=%s dbname=postgres user=postgres password=secret", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}

		migrationFilePath := filepath.Join("..", "..", "db", "migrations")
		return driver.Migration(migrationFilePath, db)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return db
}

func assertResponseCode(t testing.TB, want, got int) {
	t.Helper()

	if got != want {
		t.Errorf("Expected status code %d, but got %d instead", want, got)
	}
}

func signIn(userPayload *common.UserPayload) *http.Cookie {

	token, _ := common.CreateToken(userPayload)

	cookie := http.Cookie{
		Name:    "session",
		Value:   token,
		Expires: time.Now().Add(10 * time.Minute),
		Path:    "/auth",
	}

	return &cookie
}

type OrderPublisherStub struct{}

func (p *OrderPublisherStub) Created(order *entity.Order) error {
	fmt.Println("Order publisher publish order created event")

	return nil
}

func (p *OrderPublisherStub) Cancelled(order *entity.Order) error {
	fmt.Println("Order publisher publish order cancelled event")

	return nil
}

type TicketHelper struct {
	repository.TicketRepository
}
