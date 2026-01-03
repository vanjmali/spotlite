package payload

type SendExpiryEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}
