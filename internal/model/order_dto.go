package model

type OrderDTO struct {
	TicketID int64 `json:"ticketId" validate:"required"`
}
