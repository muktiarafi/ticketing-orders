package repository

import (
	"github.com/muktiarafi/ticketing-orders/internal/entity"
)

type TicketRepository interface {
	Insert(ticket *entity.Ticket) (*entity.Ticket, error)
	FindOne(ticketId int64) (*entity.Ticket, error)
	Update(ticket *entity.Ticket) (*entity.Ticket, error)
	UpdateByEvent(ticket *entity.Ticket) (*entity.Ticket, error)
}
