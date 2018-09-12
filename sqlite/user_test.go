package sqlite

import (
	"context"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestUserServiceCreateUser(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	expected := cabby.User{Email: "test@test.test", CanAdmin: true}
	pass := "new-user-password"

	err := s.CreateUser(context.Background(), expected, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.User(context.Background(), expected.Email, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareUser(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceCreateUserInvalidUserPassword(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	err := s.CreateUser(context.Background(), cabby.User{}, "password")
	if err == nil {
		t.Error("Expected an err")
	}

	err = s.CreateUser(context.Background(), cabby.User{Email: "foo@foo.com"}, "")
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceCreateUserQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateUser(context.Background(), cabby.User{Email: "foo@foo.com"}, "password")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceCreateUserPasswordQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user_pass")
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateUser(context.Background(), cabby.User{Email: "foo@foo.com"}, "password")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceDeleteUser(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	// create and verify a user
	userEmail := "test@test.test"
	expected := cabby.User{Email: userEmail, CanAdmin: true}
	pass := "new-user-password"

	err := s.CreateUser(context.Background(), expected, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.User(context.Background(), expected.Email, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareUser(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// delete and verify user is gone
	err = s.DeleteUser(context.Background(), userEmail)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err = s.User(context.Background(), expected.Email, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	if result.Email != "" {
		t.Error("Got:", result, `Expected: ""`)
	}
}

func TestUserServiceDeleteUserQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteUser(context.Background(), "foo")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceDeleteUserPasswordQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user_pass")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteUser(context.Background(), "foo")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceUpdateUser(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	// create a user
	expected := cabby.User{Email: "test@test.test", CanAdmin: false}
	pass := "new-user-password"

	err := s.CreateUser(context.Background(), expected, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.User(context.Background(), expected.Email, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareUser(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// update to be admin
	expected = cabby.User{Email: "test@test.test", CanAdmin: true}

	err = s.UpdateUser(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	// check it
	result, err = s.User(context.Background(), expected.Email, pass)
	if err != nil {
		t.Error("Got:", err)
	}

	passed = tester.CompareUser(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceUpdateUserInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	err := s.UpdateUser(context.Background(), cabby.User{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceUpdateUserQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpdateUser(context.Background(), cabby.User{Email: "foo@foo.com"})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceUser(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	expected := tester.User

	result, err := s.User(tester.Context, tester.UserEmail, tester.UserPassword)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareUser(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceUserQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.User(tester.Context, tester.UserEmail, tester.UserPassword)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceUserCollections(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	expected := tester.UserCollectionList

	result, err := s.UserCollections(tester.Context, tester.UserEmail)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Email != expected.Email {
		t.Error("Got:", result.Email, "Expected:", expected.Email)
	}

	for id, ca := range result.CollectionAccessList {
		if ca.CanRead != expected.CollectionAccessList[id].CanRead {
			t.Error("Got:", ca.CanRead, "Expected:", expected.CollectionAccessList[id].CanRead)
		}
	}
}

func TestUserServiceUserCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.UserCollections(tester.Context, tester.UserEmail)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		raw  string
		hash string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"bloop", "bdf0ff3f50f492bd0fb261caf097829138f472dd0ab3b504fe0f01e8c8225664"},
	}

	for _, test := range tests {
		result := hash(test.raw)
		if result != test.hash {
			t.Error("Got:", result, "Expected:", test.hash)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password    string
		expectError bool
	}{
		{"", true},
		{"12345678", false},
	}

	for _, test := range tests {
		result := validatePassword(test.password)

		if test.expectError && result == nil {
			t.Error("Got:", result, "Expected:", test.expectError)
		}
	}
}
