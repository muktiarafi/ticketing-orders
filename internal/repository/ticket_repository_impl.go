package repository

import (
	"database/sql"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/driver"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
)

type TicketRepositoryImpl struct {
	*driver.DB
}

func NewTicketRepository(db *driver.DB) TicketRepository {
	return &TicketRepositoryImpl{
		DB: db,
	}
}

func (r *TicketRepositoryImpl) Insert(ticket *entity.Ticket) (*entity.Ticket, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `INSERT INTO tickets (id, title, price)
	VALUES ($1, $2, $3)
	RETURNING *`

	newTicket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(ctx, stmt, ticket.ID, ticket.Title, ticket.Price).Scan(
		&newTicket.ID,
		&newTicket.Title,
		&newTicket.Price,
		&newTicket.Version,
	); err != nil {
		return nil, &common.Error{Op: "TicketRepository.Insert", Err: err}
	}

	return newTicket, nil
}

func (r *TicketRepositoryImpl) FindOne(ticketID int64) (*entity.Ticket, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT * FROM tickets
	WHERE id = $1`

	ticket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(ctx, stmt, ticketID).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Price,
		&ticket.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "TicketRepository.FindOne",
				Message: "Ticket Not Found",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "TicketRepository.FindOne", Err: err}
	}

	return ticket, nil
}

func (r *TicketRepositoryImpl) Update(ticket *entity.Ticket) (*entity.Ticket, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `UPDATE tickets
	SET title = $1, price =$2, version = $3
	WHERE id = $4
	RETURNING *`

	updatedTicket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		ticket.Title,
		ticket.Price,
		ticket.Version+1,
		ticket.ID,
	).Scan(
		&updatedTicket.ID,
		&updatedTicket.Title,
		&updatedTicket.Price,
		&updatedTicket.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "TicketRepository.Update",
				Message: "Ticket Not Found",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "TicketRepository.Update", Err: err}
	}

	return updatedTicket, nil
}

func (r *TicketRepositoryImpl) UpdateByEvent(ticket *entity.Ticket) (*entity.Ticket, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `UPDATE tickets
	SET title = $1, price =$2, version = $3
	WHERE id = $4 AND version = $5
	RETURNING *`

	updatedTicket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		ticket.Title,
		ticket.Price,
		ticket.Version,
		ticket.ID,
		ticket.Version-1,
	).Scan(
		&updatedTicket.ID,
		&updatedTicket.Title,
		&updatedTicket.Price,
		&updatedTicket.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "OrderRepository.FindOne",
				Message: "Ticket version is out of sync",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "TicketRepository.Update", Err: err}
	}

	return updatedTicket, nil
}
