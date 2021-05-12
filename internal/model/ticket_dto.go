package model

type TicketDTO struct {
	Title string  `json:"title" validate:"required,min=4"`
	Price float64 `json:"price" validate:"required,gt=0"`
}
