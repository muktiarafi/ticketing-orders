package repository

import "github.com/muktiarafi/ticketing-orders/internal/entity"

type OrderRepository interface {
	Insert(order *entity.Order) (*entity.Order, error)
	Find(userID int64) ([]*entity.Order, error)
	FindReserved(userID int64) ([]*entity.Order, error)
	FindOne(orderID int64) (*entity.Order, error)
	FindOneByTicketID(ticketID int64) (*entity.Order, error)
	Update(order *entity.Order) (*entity.Order, error)
}
