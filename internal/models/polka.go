package models

import "github.com/google/uuid"

type webhookEventData struct {
	UserID uuid.UUID `json:"user_id"`
}

type WebhookResource struct {
	Event string `json:"event"`
	Data webhookEventData `json:"data"`
}
