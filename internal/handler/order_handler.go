package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/events/producer"
	"github.com/muktiarafi/ticketing-orders/internal/model"
	"github.com/muktiarafi/ticketing-orders/internal/service"
)

type OrderHandler struct {
	service.OrderService
	producer.OrderProducer
}

func NewOrderHandler(orderSrv service.OrderService, orderProducer producer.OrderProducer) *OrderHandler {
	return &OrderHandler{
		OrderService:  orderSrv,
		OrderProducer: orderProducer,
	}
}

func (h *OrderHandler) Route(e *echo.Echo) {
	orders := e.Group("/api/orders", common.RequireAuth)
	orders.POST("", h.Create)
	orders.GET("", h.GetAll)
	orders.GET("/:orderID", h.Show)
	orders.PUT("/:orderID", h.Update)
}

func (h *OrderHandler) Create(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	const op = "OrderHandler.Create"
	if !ok {
		return &common.Error{
			Op:  op,
			Err: errors.New("missing payload in context"),
		}
	}

	orderDTO := new(model.OrderDTO)
	if err := c.Bind(orderDTO); err != nil {
		return &common.Error{Op: op, Err: err}
	}

	if err := c.Validate(orderDTO); err != nil {
		return err
	}

	order, err := h.OrderService.Create(int64(userPayload.ID), orderDTO.TicketID)
	if err != nil {
		return err
	}

	if err := h.OrderProducer.Created(order); err != nil {
		return err
	}

	return common.NewResponse(http.StatusCreated, "Created", order).SendJSON(c)
}

func (h *OrderHandler) GetAll(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	const op = "OrderHandler.GetAll"
	if !ok {
		return &common.Error{
			Op:  op,
			Err: errors.New("missing payload in context"),
		}
	}

	orders, err := h.OrderService.Find(int64(userPayload.ID))
	if err != nil {
		return err
	}

	return common.NewResponse(http.StatusOK, "OK", orders).SendJSON(c)
}

func (h *OrderHandler) Show(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	const op = "OrderHandler.Show"
	if !ok {
		return &common.Error{
			Op:  op,
			Err: errors.New("missing payload in context"),
		}
	}

	orderIDParam := c.Param("orderID")
	orderID, err := strconv.ParseInt(orderIDParam, 10, 64)
	if err != nil {
		return &common.Error{
			Code:    common.EINVALID,
			Op:      "OrderHandler.Show",
			Message: "Invalid order Id",
			Err:     err,
		}
	}
	order, err := h.OrderService.Show(int64(userPayload.ID), orderID)
	if err != nil {
		return err
	}

	return common.NewResponse(http.StatusOK, "OK", order).SendJSON(c)
}

func (h *OrderHandler) Update(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	const op = "OrderHandler.Show"
	if !ok {
		return &common.Error{
			Op:  op,
			Err: errors.New("missing payload in context"),
		}
	}

	orderIDParam := c.Param("orderID")
	orderID, err := strconv.ParseInt(orderIDParam, 10, 64)
	if err != nil {
		return &common.Error{
			Code:    common.EINVALID,
			Op:      "OrderHandler.Show",
			Message: "Invalid order Id",
			Err:     err,
		}
	}

	order, err := h.OrderService.Update(int64(userPayload.ID), orderID)
	if err != nil {
		return err
	}

	if err := h.OrderProducer.Cancelled(order); err != nil {
		return err
	}

	return common.NewResponse(http.StatusOK, "OK", order).SendJSON(c)
}
