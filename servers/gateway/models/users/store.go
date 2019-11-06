package users

import (
	"errors"
	"net/http"
)

//ErrUserNotFound is returned when the user can't be found
var ErrUserNotFound = errors.New("user not found")

// //UserStore represents a connection to our user database
// type UserStore struct {
// 	Store *Store
// }

//Store represents a store for Users
type Store interface {
	// //NewStore returns the Store with an open database connection to do queries and transactions on
	// NewStore() *UserStore

	//GetByID returns the User with the given ID
	GetByID(id int64) (*User, error)

	//GetByEmail returns the User with the given email
	GetByEmail(email string) (*User, error)

	//GetByUserName returns the User with the given Username
	GetByUserName(username string) (*User, error)

	//Insert inserts the user into the database, and returns
	//the newly-inserted User, complete with the DBMS-assigned ID
	Insert(user *User) (*User, error)

	LogSuccessfulSignIns(user *User, r *http.Request)

	//Update applies UserUpdates to the given user ID
	//and returns the newly-updated user
	Update(id int64, updates *Updates) (*User, error)

	//Delete deletes the user with the given ID
	Delete(id int64) error
}
