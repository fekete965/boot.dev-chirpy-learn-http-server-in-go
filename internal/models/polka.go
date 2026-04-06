package models

import "github.com/google/uuid"

type WebhookEventData struct {
	UserID uuid.UUID `json:"user_id"`
}

type WebhookResource struct {
	Event string           `json:"event"`
	Data  WebhookEventData `json:"data"`
}
