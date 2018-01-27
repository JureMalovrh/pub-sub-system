package database

import (
	"fmt"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

//Person definition
type Person struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bson:"name,omitempty"`
	IsActive bool          `bson:"isActive,omitempty"`
}

//Storage interface definition
type Storage interface {
	GetUserByID(userID string) (Person, error)
}

//UserStorage definition
type UserStorage struct {
	Collection *mgo.Collection
}

//NewUserStorage returns new UserStorage, that handles db calls
func NewUserStorage(db *mgo.Session, table, collection string) Storage {
	dbCollection := db.DB(table).C(collection)
	return &UserStorage{
		Collection: dbCollection,
	}
}

//GetUserByID returns user from DB
func (us *UserStorage) GetUserByID(userID string) (Person, error) {
	person := Person{}
	if !bson.IsObjectIdHex(userID) {
		return Person{}, fmt.Errorf("ObjectID not valid")
	}

	err := us.Collection.FindId(bson.ObjectIdHex(userID)).One(&person)
	if err != nil {
		log.Printf("method GetUserByID, error %s", err)
		return Person{}, err
	}
	return person, nil
}
