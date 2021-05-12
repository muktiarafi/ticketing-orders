package service

import (
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/model"
)

type TicketService interface {
	Create(int64, *model.TicketDTO) (*entity.Ticket, error)
	Find() ([]*entity.Ticket, error)
	FindOne(ticketID int64) (*entity.Ticket, error)
	Update(userID, ticketID int64, ticketDTO *model.TicketDTO) (*entity.Ticket, error)
}
