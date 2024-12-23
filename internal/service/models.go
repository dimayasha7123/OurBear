package service

type updates struct {
	Ok     bool     `json:"ok"`
	Result []update `json:"result"`
}

type update struct {
	UpdateID int64   `json:"update_id"`
	Message  message `json:"message"`
}

type message struct {
	MessageID int64 `json:"message_id"`
	From      struct {
		ID           int64  `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
	} `json:"from"`
	Chat struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date int64  `json:"date"`
	Text string `json:"text"`
}
