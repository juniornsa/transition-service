package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"upworkfixmux/models"
)

func JSON(w http.ResponseWriter, code int, r *models.Response) {

	data, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(500)
		errResponse := &models.Response{
			Error: true, Message: "Something went wrong!",
		}
		bytes, _ := json.Marshal(errResponse)
		_, err := w.Write(bytes)
		if err != nil {
			log.Println(err.Error())
		}
		return
	}

	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err.Error())
	}
}
