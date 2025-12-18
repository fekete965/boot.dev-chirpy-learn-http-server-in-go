package models

type EnvVars struct {
	DbUrl string
	JWTSecret string
	Platform string
	PolkaWebhookSecret string
	Port int
}
