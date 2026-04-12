package internal

// SubscribeRequest matches the swagger SubscribeRequest definition.
// Used for POST /api/subscribe.
type SubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Repo  string `json:"repo" binding:"required"`
}

// SubscriptionResponse matches the swagger Subscription definition.
// Used for GET /api/subscriptions response items.
type SubscriptionResponse struct {
	Email       string `json:"email"`
	Repo        string `json:"repo"`
	Confirmed   bool   `json:"confirmed"`
	LastSeenTag string `json:"last_seen_tag"`
}

// ErrorResponse is a generic error envelope for API responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse is a generic success envelope for API responses.
type MessageResponse struct {
	Message string `json:"message"`
}

// ToSubscriptionResponse converts a Subscription model to its API response DTO.
func ToSubscriptionResponse(s *Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		Email:       s.Email,
		Repo:        s.Repo,
		Confirmed:   s.Confirmed,
		LastSeenTag: s.LastSeenTag,
	}
}

// ToSubscriptionListResponse converts a slice of Subscription models to response DTOs.
func ToSubscriptionListResponse(subs []Subscription) []SubscriptionResponse {
	result := make([]SubscriptionResponse, len(subs))
	for i := range subs {
		result[i] = ToSubscriptionResponse(&subs[i])
	}
	return result
}
