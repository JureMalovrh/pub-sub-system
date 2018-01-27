package database_test

import (
	"fmt"
	"pub-sub/tracker/database"
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

/*
[database]
server = "mongodb://database"
port = "27017"
table = "tracker"
collection = "user"

*/

func connectToDB() *mgo.Session {
	session, err := mgo.Dial("mongodb://127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	demoData(session)
	return session
}

func demoData(session *mgo.Session) {
	c := session.DB("tracker_test").C("user")
	err := c.Insert(
		&database.Person{ID: bson.ObjectIdHex("5555e2d316ca1b6d40aaaaaa"), Name: "test user 1", IsActive: true},
		&database.Person{ID: bson.ObjectIdHex("5555e2d316ca1b6d40aaaaab"), Name: "test user 2", IsActive: false},
	)
	if err != nil {
		panic(err)
	}
}

func dropData(session *mgo.Session) {
	c := session.DB("tracker_test").C("user")
	err := c.DropCollection()
	if err != nil {
		panic(err)
	}
}

func Test(t *testing.T) {
	testCases := []struct {
		desc           string
		id             string
		expectedError  error
		expectedPerson database.Person
	}{
		{
			desc:           "Valid user id, correctly returned from DB",
			id:             "5555e2d316ca1b6d40aaaaaa",
			expectedPerson: database.Person{ID: bson.ObjectIdHex("5555e2d316ca1b6d40aaaaaa"), Name: "test user 1", IsActive: true},
		},
		{
			desc:          "User not found",
			id:            "5555e2d316ca1b6d40aaaaac",
			expectedError: fmt.Errorf("not found"),
		},
		{
			desc:          "Non objectId queried",
			id:            "nonobjid",
			expectedError: fmt.Errorf("ObjectID not valid"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			session := connectToDB()
			defer dropData(session)

			userStorage := database.NewUserStorage(session, "tracker_test", "user")
			person, err := userStorage.GetUserByID(tC.id)

			if err != nil && err.Error() != tC.expectedError.Error() {
				t.Errorf("Expected %s, got %s", tC.expectedError, err)
			}

			eq := reflect.DeepEqual(person, tC.expectedPerson)
			if !eq {
			}
		})
	}
}
