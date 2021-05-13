package repository

import (
	"database/sql"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/driver"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
)

type OrderRepositoryImpl struct {
	*driver.DB
}

func NewOrderRepository(db *driver.DB) OrderRepository {
	return &OrderRepositoryImpl{
		DB: db,
	}
}

func (r *OrderRepositoryImpl) Insert(order *entity.Order) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `INSERT INTO orders (status, expires_at, user_id, ticket_id)
	VALUES ($1, $2, $3, $4)
	RETURNING *`

	newOrder := new(entity.Order)
	var ticketID int64
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		order.Status,
		order.ExpiresAt,
		order.UserID,
		order.Ticket.ID,
	).Scan(
		&newOrder.ID,
		&newOrder.Status,
		&newOrder.ExpiresAt,
		&newOrder.UserID,
		&ticketID,
		&newOrder.Version,
	); err != nil {
		return nil, &common.Error{Op: "OrderRepository.Insert", Err: err}
	}

	ticketStmt := `SELECT * FROM tickets
	WHERE id = $1`

	ticket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(ctx, ticketStmt, ticketID).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Price,
		&ticket.Version,
	); err != nil {
		return nil, &common.Error{Op: "OrderRepository.Insert", Err: err}
	}

	newOrder.Ticket = ticket

	return newOrder, nil
}

func (r *OrderRepositoryImpl) Find(userID int64) ([]*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT o.id, status, expires_at, user_id, o.version, t.id, title, price, t.version
	FROM orders AS o JOIN tickets AS t
	ON o.ticket_id = t.id
	WHERE o.user_id = $1`

	rows, err := r.SQL.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, &common.Error{Op: "OrderRepository.Find", Err: err}
	}
	defer rows.Close()

	order := new(entity.Order)
	ticket := new(entity.Ticket)
	orders := make([]*entity.Order, 0)
	for rows.Next() {
		if err := rows.Scan(
			&order.ID,
			&order.Status,
			&order.ExpiresAt,
			&order.UserID,
			&order.Version,
			&ticket.ID,
			&ticket.Title,
			&ticket.Price,
			&ticket.Version,
		); err != nil {
			return nil, &common.Error{Op: "OrderRepository.Find", Err: err}
		}
		order.Ticket = ticket

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) FindReserved(ticketID int64) ([]*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT o.id, status, expires_at, user_id, o.version, t.id, title, price, t.version
	FROM orders AS o JOIN tickets AS t
	ON o.ticket_id = t.id
	WHERE t.id = $1 AND status IN ('CREATED', 'PENDING', 'COMPLETED')`

	rows, err := r.SQL.QueryContext(ctx, stmt, ticketID)
	if err != nil {
		return nil, &common.Error{Op: "OrderRepositoryImpl.FindReserved", Err: err}
	}
	defer rows.Close()

	order := new(entity.Order)
	ticket := new(entity.Ticket)
	orders := make([]*entity.Order, 0)
	for rows.Next() {
		if err := rows.Scan(
			&order.ID,
			&order.Status,
			&order.ExpiresAt,
			&order.UserID,
			&order.Version,
			&ticket.ID,
			&ticket.Title,
			&ticket.Price,
			&ticket.Version,
		); err != nil {
			return nil, &common.Error{Op: "OrderRepositoryImpl.FindReserved", Err: err}
		}
		order.Ticket = ticket
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) FindOne(orderID int64) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT o.id, status, expires_at, user_id, o.version, t.id, title, price, t.version
	FROM orders AS o JOIN tickets AS t
	ON o.ticket_id = t.id
	WHERE o.id = $1`

	order := new(entity.Order)
	ticket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(ctx, stmt, orderID).Scan(
		&order.ID,
		&order.Status,
		&order.ExpiresAt,
		&order.UserID,
		&order.Version,
		&ticket.ID,
		&ticket.Title,
		&ticket.Price,
		&ticket.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "OrderRepository.FindOne",
				Message: "Order Not Found",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "OrderRepository.FindOne", Err: err}
	}
	order.Ticket = ticket

	return order, nil
}

func (r *OrderRepositoryImpl) FindOneByTicketID(ticketID int64) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT o.id, status, expires_at, user_id, o.version, t.id, title, price, t.version
	FROM orders AS o JOIN tickets AS t
	ON o.ticket_id = t.id
	WHERE t.id = $1`

	order := new(entity.Order)
	ticket := new(entity.Ticket)
	if err := r.SQL.QueryRowContext(ctx, stmt, ticketID).Scan(
		&order.ID,
		&order.Status,
		&order.ExpiresAt,
		&order.UserID,
		&order.Version,
		&ticket.ID,
		&ticket.Title,
		&ticket.Price,
		&ticket.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "OrderRepository.FindOne",
				Message: "Order Not Found",
				Err:     err,
			}
		}

		return nil, &common.Error{Op: "OrderRepository.FindOne", Err: err}
	}
	order.Ticket = ticket

	return order, nil
}

func (r *OrderRepositoryImpl) Update(order *entity.Order) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `UPDATE orders
	SET status = $1, version = $2
	WHERE id = $3
	RETURNING id, status, expires_at, user_id, version`

	updatedOrder := new(entity.Order)
	if err := r.SQL.QueryRowContext(ctx, stmt, order.Status, order.Version+1, order.ID).Scan(
		&updatedOrder.ID,
		&updatedOrder.Status,
		&updatedOrder.ExpiresAt,
		&updatedOrder.UserID,
		&updatedOrder.Version,
	); err != nil {
		return nil, &common.Error{Op: "OrderRepository.Update", Err: err}
	}
	updatedOrder.Ticket = order.Ticket

	return updatedOrder, nil
}

func (r *OrderRepositoryImpl) UpdateOnEvent(order *entity.Order) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `UPDATE orders
	SET status = $1, version = $2
	WHERE id = $3 AND version = $4
	RETURNING id, status, expires_at, user_id, version`

	updatedOrder := new(entity.Order)
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		order.Status,
		order.Version,
		order.ID,
		order.Version-1,
	).Scan(
		&updatedOrder.ID,
		&updatedOrder.Status,
		&updatedOrder.ExpiresAt,
		&updatedOrder.UserID,
		&updatedOrder.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Message: "Order not found. Version probably out of sync",
				Op:      "OrderRepository.UpdateOnEvent",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "orderRepository.UpdateOnEvent", Err: err}
	}
	updatedOrder.Ticket = order.Ticket

	return updatedOrder, nil
}
