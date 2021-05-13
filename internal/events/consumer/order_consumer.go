package consumer

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-common/types"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/repository"
)

type OrderConsumer struct {
	repository.TicketRepository
}

func NewOrderConsumer(ticketRepo repository.TicketRepository) *OrderConsumer {
	return &OrderConsumer{
		TicketRepository: ticketRepo,
	}
}

func (c *OrderConsumer) TicketCreated(msg *message.Message) error {
	fmt.Println("received event from topic ", common.TicketCreated)
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
	fmt.Println("received event from topic ", common.TIcketUpdated)
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
