package entity

type Ticket struct {
	ID      int64   `json:"id"`
	Title   string  `json:"title"`
	Price   float64 `json:"price"`
	Version int64   `json:"version"`
}
