package db

import (
	"encoding/base64"
)

var testDB *Handler

func init() {
	const dataSource = `postgres://licensing_testing:licensing_testing@localhost:5432/licensing_testing?sslmode=disable`

	// Clear testing database
	db, err := Open(dataSource)
	if err != nil {
		panic(err)
	}
	_, err = db.db.Exec("DROP SCHEMA IF EXISTS public CASCADE;")
	if err != nil {
		panic(err)
	}
	_, err = db.db.Exec("CREATE SCHEMA public;")
	if err != nil {
		panic(err)
	}
	err = db.db.Close()
	if err != nil {
		panic(err)
	}

	_, err = MigrateUp(dataSource)
	if err != nil {
		panic(err)
	}
	testDB, err = Open(dataSource)
	if err != nil {
		panic(err)
	}
}

func base64Key(str string) *[32]byte {
	bs, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return (*[32]byte)(bs)
}
