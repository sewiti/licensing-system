package db

import (
	"encoding/base64"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
)

func newMock() (*Handler, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}
	sq := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(db)
	return &Handler{
		db: db,
		sq: sq,
	}, mock, nil
}

func base64Key(str string) []byte {
	bs, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return bs
}
