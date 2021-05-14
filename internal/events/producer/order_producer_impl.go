package producer

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-common/types"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
)

type OrderProducerImpl struct {
	message.Publisher
}

func NewOrderProducer(publisher message.Publisher) OrderProducer {
	return &OrderProducerImpl{
		Publisher: publisher,
	}
}

func (p *OrderProducerImpl) Created(order *entity.Order) error {
	orderCreatedEventData := types.OrderCreatedEvent{
		ID:          order.ID,
		Status:      order.Status,
		Version:     order.Version,
		UserID:      order.UserID,
		ExpiresAt:   order.ExpiresAt.Format(time.RFC3339),
		TicketID:    order.Ticket.ID,
		TicketPrice: order.Ticket.Price,
	}
	orderBytes, err := orderCreatedEventData.Marshal()
	if err != nil {
		return &common.Error{Op: "OrderProducer.Created", Err: err}
	}

	msg := message.NewMessage(watermill.NewUUID(), orderBytes)
	return p.Publish(common.OrderCreated, msg)
}

func (p *OrderProducerImpl) Cancelled(order *entity.Order) error {
	orderCancelledData := types.OrderCancelledEvent{
		ID:       order.ID,
		Version:  order.Version,
		TicketID: order.Ticket.ID,
	}
	orderBytes, err := orderCancelledData.Marshal()
	if err != nil {
		return &common.Error{Op: "OrderProducer.Cancelled", Err: err}
	}

	msg := message.NewMessage(watermill.NewUUID(), orderBytes)
	return p.Publish(common.OrderCancelled, msg)
}
