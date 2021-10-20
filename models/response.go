package models

type Response struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponseFromApi struct {
	Message  string `json:"message"`
	Response bool   `json:"response"`
}
