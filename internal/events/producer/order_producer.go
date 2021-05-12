package producer

import "github.com/muktiarafi/ticketing-orders/internal/entity"

type OrderProducer interface {
	Created(order *entity.Order) error
	Cancelled(order *entity.Order) error
}
