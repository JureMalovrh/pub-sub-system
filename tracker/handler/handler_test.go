package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"pub-sub/tracker/database"
	"pub-sub/tracker/handler"
	"pub-sub/tracker/socket"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

func TestMetricsHandler(t *testing.T) {
	testCases := []struct {
		desc             string
		dataURL          string
		addAccountID     bool
		accountID        string
		expectedCode     int
		expectedResponse string

		returnPerson      database.Person
		returnPersonError error
		databaseCall      bool
		socketCall        bool
	}{
		{
			desc:             "AccountID not present",
			dataURL:          "?data=test",
			addAccountID:     false,
			expectedCode:     400,
			expectedResponse: `{"error": "AccountId not present"}`,
		},
		{
			desc:             "Data not present",
			dataURL:          "",
			addAccountID:     true,
			accountID:        "5555e2d316ca1b6d40aaaaaa",
			expectedCode:     400,
			expectedResponse: `{"error": "Data not present"}`,
		},
		{
			desc:              "User not in database",
			dataURL:           "?data=test",
			addAccountID:      true,
			accountID:         "5555e2d316ca1b6d40aaaaaa",
			expectedCode:      404,
			expectedResponse:  `{"error": "not found"}`,
			returnPerson:      database.Person{},
			returnPersonError: fmt.Errorf("not found"),
			databaseCall:      true,
		},
		{
			desc:              "Some other error from database",
			dataURL:           "?data=test",
			addAccountID:      true,
			accountID:         "5555e2d316ca1b6d40aaaaaa",
			expectedCode:      400,
			expectedResponse:  `{"error": "error"}`,
			returnPerson:      database.Person{},
			returnPersonError: fmt.Errorf("error"),
			databaseCall:      true,
		},
		{
			desc:             "Account is not active",
			dataURL:          "?data=test",
			addAccountID:     true,
			accountID:        "5555e2d316ca1b6d40aaaaaa",
			expectedCode:     200,
			expectedResponse: `{"data": "Account not active"}`,
			returnPerson:     database.Person{ID: "5555e2d316ca1b6d40aaaaaa", IsActive: false},
			databaseCall:     true,
		},
		{
			desc:             "Account is active",
			dataURL:          "?data=test",
			addAccountID:     true,
			accountID:        "5555e2d316ca1b6d40aaaaaa",
			expectedCode:     202,
			expectedResponse: `{"data": "Account acepted"}`,
			returnPerson:     database.Person{ID: "5555e2d316ca1b6d40aaaaaa", IsActive: true},
			databaseCall:     true,
			socketCall:       true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			_ = tC
			url := "/accountId"
			url = fmt.Sprintf("%s%s", url, tC.dataURL)
			req, _ := http.NewRequest("GET", url, nil)
			if tC.addAccountID {
				req = mux.SetURLVars(req, map[string]string{"accountId": tC.accountID})
			}

			mockSocket := socket.NewMockClient(ctrl)
			mockDatabase := database.NewMockStorage(ctrl)
			if tC.databaseCall {
				mockDatabase.EXPECT().GetUserByID(tC.accountID).Return(tC.returnPerson, tC.returnPersonError)
			}
			if tC.socketCall {
				mockSocket.EXPECT().SendMessage(tC.accountID, "test").AnyTimes()
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler.NewAccountHandler(mockDatabase, mockSocket))
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != tC.expectedResponse {
				t.Errorf("Expected %s, got %s", tC.expectedResponse, rr.Body.String())
			}

			if rr.Code != tC.expectedCode {
				t.Errorf("Expected %d, got %d", tC.expectedCode, rr.Code)
			}
		})
	}
}
