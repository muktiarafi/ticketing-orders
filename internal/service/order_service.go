package service

import "github.com/muktiarafi/ticketing-orders/internal/entity"

type OrderService interface {
	Create(userID int64, ticketID int64) (*entity.Order, error)
	Find(userID int64) ([]*entity.Order, error)
	Show(userID, orderID int64) (*entity.Order, error)
	Update(userID, orderID int64) (*entity.Order, error)
}
