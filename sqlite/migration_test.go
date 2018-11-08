package sqlite

import "testing"

func TestMigrationServiceUp(t *testing.T) {
	tearDownSQLite()
	ds := testDataStore()
	s := ds.MigrationService()

	err := s.Up()
	if err != nil {
		t.Error(err)
	}

	// try if the table is gone
	ds.DB.Exec("drop table schema_version")
	err = s.Up()
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestMigrationServiceCurrentVersion(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.MigrationService()

	version, err := s.CurrentVersion()
	if version != 1 {
		t.Error("Got:", version, "Expected:", 1, "Error:", err)
	}
}

func TestMigrationServiceCurrentVersionFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.MigrationService()

	ds.DB.Exec("drop table schema_version")
	ds.DB.Exec("create table schema_version (id int)")

	_, err := s.CurrentVersion()
	if err == nil {
		t.Error("Expected an error")
	}
}
