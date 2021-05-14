package consumer

import (
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-common/types"
	"github.com/muktiarafi/ticketing-orders/internal/constant"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/events/producer"
	"github.com/muktiarafi/ticketing-orders/internal/repository"
)

type OrderConsumer struct {
	producer.OrderProducer
	repository.OrderRepository
	repository.TicketRepository
}

func NewOrderConsumer(producer producer.OrderProducer, orderRepo repository.OrderRepository, ticketRepo repository.TicketRepository) *OrderConsumer {
	return &OrderConsumer{
		OrderProducer:    producer,
		OrderRepository:  orderRepo,
		TicketRepository: ticketRepo,
	}
}

func (c *OrderConsumer) TicketCreated(msg *message.Message) error {
	log.Println("received event from topic:", common.TicketCreated)
	ticketCreatedData := new(types.TicketCreatedEvent)
	if err := ticketCreatedData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return &common.Error{Op: "OrderConsumer.TicketCreated", Err: err}
	}

	ticket := &entity.Ticket{
		ID:    ticketCreatedData.ID,
		Title: ticketCreatedData.Title,
		Price: ticketCreatedData.Price,
	}

	if _, err := c.TicketRepository.Insert(ticket); err != nil {
		msg.Nack()
		return &common.Error{Op: "OrderConsumer.TicketCreated", Err: err}
	}

	msg.Ack()

	return nil
}

func (c *OrderConsumer) TicketUpdated(msg *message.Message) error {
	log.Println("received event from topic:", common.TIcketUpdated)
	ticketUpdatedData := new(types.TicketUpdatedEvent)
	if err := ticketUpdatedData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return &common.Error{Op: "OrderConsumer.TicketUpdated", Err: err}
	}

	ticket := &entity.Ticket{
		ID:      ticketUpdatedData.ID,
		Title:   ticketUpdatedData.Title,
		Price:   ticketUpdatedData.Price,
		Version: ticketUpdatedData.Version,
	}

	if _, err := c.TicketRepository.UpdateByEvent(ticket); err != nil {
		er, _ := err.(*common.Error)
		if er.Code == common.ECONCLICT {
			msg.Ack()
		} else {
			msg.Nack()
		}
		return &common.Error{Op: "OrderConsumer.TicketUpdated", Err: err}
	}

	msg.Ack()

	return nil
}

func (c *OrderConsumer) ExpirationComplete(msg *message.Message) error {
	log.Println("received event from topic:", common.ExpirationComplete)
	expirationCompleteData := new(types.ExpirationCompleteEvent)
	if err := expirationCompleteData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return err
	}

	order, err := c.OrderRepository.FindOne(expirationCompleteData.OrderID)
	if err != nil {
		er, _ := err.(*common.Error)
		if er.Code == common.ENOTFOUND {
			msg.Ack()
		} else {
			msg.Nack()
		}
		return err
	}

	if order.Status == constant.COMPLETED {
		msg.Ack()
		return nil
	}

	order.Status = constant.CANCELLED
	order.Version++
	if _, err := c.OrderRepository.Update(order); err != nil {
		msg.Nack()
		return err
	}

	if err := c.OrderProducer.Cancelled(order); err != nil {
		msg.Nack()
		return err
	}

	msg.Ack()

	return nil
}

func (c *OrderConsumer) PaymentCreated(msg *message.Message) error {
	log.Println("received event from topic:", common.PaymentCreated)
	paymentCreatedEventData := new(types.PaymentCreatedEvent)
	if err := paymentCreatedEventData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return err
	}

	order, err := c.OrderRepository.FindOne(paymentCreatedEventData.OrderID)
	if err != nil {
		er, _ := err.(*common.Error)
		if er.Code == common.ENOTFOUND {
			msg.Ack()
		} else {
			msg.Nack()
		}
	}

	order.Status = constant.COMPLETED
	order.Version++
	if _, err := c.OrderRepository.Update(order); err != nil {
		msg.Nack()
		return err
	}

	msg.Ack()

	return nil
}
