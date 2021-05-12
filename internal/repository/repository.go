package repository

import (
	"context"
	"time"
)

func newDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}
