package constants

import "time"

const AUTHORIZATION_HEADER = "Authorization"
const BEARER_TOKEN_PREFIX = "Bearer "
const POLKA_WEBHOOK_TOKEN_PREFIX = "ApiKey "

var (
	DEFAULT_EXPIRES_IN       time.Duration = time.Hour
	REFRESH_TOKEN_EXPIRES_IN time.Duration = 60 * 24 * time.Hour
)
