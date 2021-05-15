package service

import (
	"database/sql"
	"errors"
	"time"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/constant"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/events/producer"
	"github.com/muktiarafi/ticketing-orders/internal/repository"
)

type OrderServiceImpl struct {
	repository.OrderRepository
	repository.TicketRepository
	producer.OrderProducer
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	ticketRepo repository.TicketRepository,
	orderProducer producer.OrderProducer,
) OrderService {
	return &OrderServiceImpl{
		OrderRepository:  orderRepo,
		TicketRepository: ticketRepo,
		OrderProducer:    orderProducer,
	}
}

func (s *OrderServiceImpl) Create(userID int64, ticketID int64) (*entity.Order, error) {
	ticket, err := s.TicketRepository.FindOne(ticketID)
	if err != nil {
		return nil, err
	}
	orders, err := s.OrderRepository.FindReserved(ticket.ID)
	er, ok := err.(*common.Error)
	if ok {
		if er.Err != sql.ErrNoRows {
			return nil, err
		}
	}
	if len(orders) != 0 {
		return nil, &common.Error{
			Op:      "OrderServiceImpl.Create",
			Code:    common.EINVALID,
			Message: "Ticket is already reserved",
			Err:     errors.New("trying to create order with reserved ticket"),
		}

	}

	newOrder := &entity.Order{
		Status:    constant.CREATED,
		UserID:    userID,
		Ticket:    ticket,
		ExpiresAt: time.Now().Add(time.Second * 60),
	}

	newOrder, err = s.OrderRepository.Insert(newOrder)
	if err != nil {
		return nil, err
	}

	if err := s.OrderProducer.Created(newOrder); err != nil {
		return nil, err
	}

	return newOrder, nil
}

func (s *OrderServiceImpl) GetAll(userID int64) ([]*entity.Order, error) {
	return s.OrderRepository.Find(userID)
}

func (s *OrderServiceImpl) Show(userID, orderID int64) (*entity.Order, error) {
	order, err := s.OrderRepository.FindOne(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, &common.Error{
			Op:      "OrderServiceImpl.Show",
			Code:    common.EINVALID,
			Message: "Not Authorized",
			Err:     errors.New("trying to access order not belonged to"),
		}
	}

	return order, nil
}

func (s *OrderServiceImpl) Update(userID, orderID int64) (*entity.Order, error) {
	order, err := s.OrderRepository.FindOne(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, &common.Error{
			Op:      "OrderServiceImpl.Update",
			Code:    common.EINVALID,
			Message: "Not Authorized",
			Err:     errors.New("trying to access order not belonged to"),
		}
	}
	order.Status = constant.CANCELLED

	updatedOrder, err := s.OrderRepository.Update(order)
	if err != nil {
		return nil, err
	}

	if err := s.OrderProducer.Cancelled(updatedOrder); err != nil {
		return nil, err
	}

	return updatedOrder, nil
}
