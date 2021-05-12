package server

import (
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/config"
	"github.com/muktiarafi/ticketing-orders/internal/driver"
	"github.com/muktiarafi/ticketing-orders/internal/events/consumer"
	"github.com/muktiarafi/ticketing-orders/internal/events/producer"
	"github.com/muktiarafi/ticketing-orders/internal/handler"
	custommiddleware "github.com/muktiarafi/ticketing-orders/internal/middleware"
	"github.com/muktiarafi/ticketing-orders/internal/repository"
	"github.com/muktiarafi/ticketing-orders/internal/service"
)

func SetupServer() *echo.Echo {
	e := echo.New()
	p := custommiddleware.NewPrometheus("echo", nil)
	p.Use(e)

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	e.Validator = customValidator
	e.HTTPErrorHandler = common.CustomErrorHandler
	e.Use(middleware.Logger())

	db, err := driver.ConnectSQL(config.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}

	orderRepository := repository.NewOrderRepository(db)
	ticketRepository := repository.NewTicketRepository(db)
	orderService := service.NewOrderService(orderRepository, ticketRepository)

	producerBrokers := []string{config.NewProducerBroker()}
	commonPublisher, err := common.CreatePublisher(producerBrokers, watermill.NewStdLogger(false, false))
	if err != nil {
		log.Fatal(err)
	}
	orderProducer := producer.NewOrderProducer(commonPublisher)

	orderHandler := handler.NewOrderHandler(orderService, orderProducer)
	orderHandler.Route(e)

	consumerBrokers := []string{config.NewConsumerBroker()}
	subscriber, err := common.CreateSubscriber(consumerBrokers, "orders-service", watermill.NewStdLogger(false, false))
	if err != nil {
		log.Fatal(err)
	}

	orderConsumer := consumer.NewOrderConsumer(ticketRepository)
	commonConsumer := common.NewConsumer(subscriber)
	commonConsumer.On(common.TicketCreated, orderConsumer.TicketCreated)
	commonConsumer.On(common.TIcketUpdated, orderConsumer.TicketUpdated)

	return e
}
