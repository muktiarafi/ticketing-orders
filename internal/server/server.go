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

	producerBrokers := []string{config.NewProducerBroker()}
	commonPublisher, err := common.NewPublisher(producerBrokers, watermill.NewStdLogger(false, false))
	if err != nil {
		log.Fatal(err)
	}
	orderProducer := producer.NewOrderProducer(commonPublisher)
	orderService := service.NewOrderService(orderRepository, ticketRepository, orderProducer)

	orderHandler := handler.NewOrderHandler(orderService)
	orderHandler.Route(e)

	subscriberConfig := &common.SubscriberConfig{
		Brokers:       []string{config.NewConsumerBroker()},
		ConsumerGroup: "orders-service",
		FromBeginning: true,
		LoggerAdapter: watermill.NewStdLogger(false, false),
	}
	subscriber, err := common.NewSubscriber(subscriberConfig)
	if err != nil {
		log.Fatal(err)
	}

	orderConsumer := consumer.NewOrderConsumer(orderProducer, orderRepository, ticketRepository)
	commonConsumer := common.NewConsumer(subscriber)
	commonConsumer.On(common.TicketCreated, orderConsumer.TicketCreated)
	commonConsumer.On(common.TIcketUpdated, orderConsumer.TicketUpdated)
	commonConsumer.On(common.ExpirationComplete, orderConsumer.ExpirationComplete)
	commonConsumer.On(common.PaymentCreated, orderConsumer.PaymentCreated)

	return e
}
