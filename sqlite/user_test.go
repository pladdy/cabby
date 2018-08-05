package sqlite

import (
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestUserServiceRead(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := UserService{DB: ds.DB}

	expected := testUser

	result, err := s.User(testUserEmail, testUserPassword)
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

	_, err = s.User(testUserEmail, testUserPassword)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceValid(t *testing.T) {
	tests := []struct {
		user     cabby.User
		expected bool
	}{
		{user: cabby.User{Email: testUserEmail}, expected: true},
		{user: cabby.User{}, expected: false},
	}

	ds := testDataStore()
	s := UserService{DB: ds.DB}

	for _, test := range tests {
		result := s.Valid(test.user)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
