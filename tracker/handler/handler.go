package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"pub-sub/tracker/database"
	"pub-sub/tracker/socket"
)

//AccountCallResponse response definition
type AccountCallResponse struct {
	StatusCode   int
	Error        string
	ResponseText string
}

func encodeJSON(data AccountCallResponse, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.StatusCode)
	if data.Error != "" {
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, data.Error)))
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"data": "%s"}`, data.ResponseText)))
}

func parseURL(r *http.Request) (string, string, error) {
	vars := mux.Vars(r)
	accountID, ok := vars["accountId"]
	if !ok {
		return "", "", fmt.Errorf("AccountId not present")
	}

	data := r.URL.Query().Get("data")
	if data == "" {
		return "", "", fmt.Errorf("Data not present")
	}
	return accountID, data, nil
}

//NewAccountHandler returns new HTTP handler for account action
func NewAccountHandler(db database.Storage, publisher socket.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, data, err := parseURL(r)

		if err != nil {
			encodeJSON(AccountCallResponse{StatusCode: http.StatusBadRequest, Error: err.Error()}, w)
			return
		}

		account, err := db.GetUserByID(accountID)
		if err != nil {
			if strings.Contains("not found", err.Error()) {
				encodeJSON(AccountCallResponse{StatusCode: http.StatusNotFound, Error: err.Error()}, w)
				return
			}
			encodeJSON(AccountCallResponse{StatusCode: http.StatusBadRequest, Error: err.Error()}, w)
			return
		}
		if !account.IsActive {
			encodeJSON(AccountCallResponse{StatusCode: http.StatusOK, ResponseText: "Account not active"}, w)
			return
		}
		go publisher.SendMessage(accountID, data)

		encodeJSON(AccountCallResponse{StatusCode: http.StatusAccepted, ResponseText: "Account acepted"}, w)
		//time.Sleep(1 * time.Second)
	}
}
