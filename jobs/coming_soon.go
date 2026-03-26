package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeComingSoon = "order:coming_soon"

type ComingSoonPayload struct {
	OrderID int64 `json:"order_id"`
}

func NewComingSoonTask(orderID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(ComingSoonPayload{
		OrderID: orderID,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeComingSoon, payload), nil
}

func HandleComingSoon(ctx context.Context, t *asynq.Task) error {
	var p ComingSoonPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	fmt.Printf("Processing order %d\n", p.OrderID)

	return nil
}
