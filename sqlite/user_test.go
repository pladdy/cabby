package sqlite

import (
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestUserServiceRead(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := UserService{DB: ds.DB}

	expected := tester.User

	result, err := s.User(expected.Email, tester.UserPassword)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Email != expected.Email {
		t.Error("Got:", result.Email, "Expected:", expected.Email)
	}
	if result.CanAdmin != expected.CanAdmin {
		t.Error("Got:", result.CanAdmin, "Expected:", expected.CanAdmin)
	}
}

func TestUserServiceReadQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := UserService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.User(tester.UserEmail, tester.UserPassword)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceExists(t *testing.T) {
	tests := []struct {
		user     cabby.User
		expected bool
	}{
		{user: cabby.User{Email: tester.UserEmail}, expected: true},
		{user: cabby.User{}, expected: false},
	}

	ds := testDataStore()
	s := UserService{DB: ds.DB}

	for _, test := range tests {
		result := s.Exists(test.user)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
