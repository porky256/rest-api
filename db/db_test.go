package db

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/porky256/rest-api/models"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestDatabasePostgres_AddBook(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error in creating mock: %s", err)
	}
	db := DatabasePostgres{Conn: conn}
	type MockBehavior func(mock sqlmock.Sqlmock, id int, book models.Book)
	tests := []struct {
		name         string
		inputBook    models.Book
		returnId     int
		mockBehavior MockBehavior
		returnErr    bool
	}{
		{
			name:      "OK",
			inputBook: models.Book{Name: "OK", Price: 1, Genre: 1, Amount: 1},
			returnId:  1,
			mockBehavior: func(mock sqlmock.Sqlmock, id int, book models.Book) {
				mock.ExpectBegin()
				mock.ExpectQuery("insert into books").
					WithArgs(book.Name, book.Price, book.Genre, book.Amount).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
				mock.ExpectCommit()
			},
		},
		{
			name:      "Missing name",
			returnId:  0,
			returnErr: true,
			inputBook: models.Book{Price: 1, Genre: 1, Amount: 1},
			mockBehavior: func(mock sqlmock.Sqlmock, id int, book models.Book) {
				mock.ExpectBegin()
				mock.ExpectQuery("insert into books").
					WithArgs(book.Name, book.Price, book.Genre, book.Amount).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(0, errors.New("input error")))
				mock.ExpectRollback()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.returnId, test.inputBook)

			id, err := db.AddBook(test.inputBook)
			if test.returnErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.returnId, id)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDatabasePostgres_GetAllBooks(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error in creating mock: %s", err)
	}
	db := DatabasePostgres{Conn: conn}
	type MockBehavior func(mock sqlmock.Sqlmock, filter map[string][]string)
	tests := []struct {
		name         string
		filter       map[string][]string
		returnBooks  []models.Book
		mockBehavior MockBehavior
		returnErr    bool
	}{
		{
			name:   "OK",
			filter: map[string][]string{},
			returnBooks: []models.Book{
				{ID: 1, Name: "book1", Price: 1, Genre: 1, Amount: 1},
				{ID: 2, Name: "book2", Price: 1, Genre: 2, Amount: 1},
			},
			mockBehavior: func(mock sqlmock.Sqlmock, filter map[string][]string) {
				mock.ExpectBegin()
				mock.ExpectQuery("select").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "genre", "amount"}).
						AddRow(1, "book1", 1, 1, 1).
						AddRow(2, "book2", 1, 2, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:   "Ok with filter",
			filter: map[string][]string{"genre": {"1"}},
			returnBooks: []models.Book{
				{ID: 1, Name: "book1", Price: 1, Genre: 1, Amount: 1},
			},
			mockBehavior: func(mock sqlmock.Sqlmock, filter map[string][]string) {
				mock.ExpectBegin()
				genre, _ := strconv.Atoi(filter["genre"][0])
				mock.ExpectQuery("select").
					WithArgs(genre).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "genre", "amount"}).
						AddRow(1, "book1", 1, 1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:        "Ok with filter but no rows",
			filter:      map[string][]string{"genre": {"3"}},
			returnBooks: []models.Book{},
			mockBehavior: func(mock sqlmock.Sqlmock, filter map[string][]string) {
				mock.ExpectBegin()
				genre, _ := strconv.Atoi(filter["genre"][0])
				mock.ExpectQuery("select").
					WithArgs(genre).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "genre", "amount"}))
				mock.ExpectCommit()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.filter)

			books, err := db.GetAllBooks(test.filter)
			if test.returnErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.returnBooks, books)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDatabasePostgres_GetBookById(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error in creating mock: %s", err)
	}
	db := DatabasePostgres{Conn: conn}
	type MockBehavior func(mock sqlmock.Sqlmock, id int)
	tests := []struct {
		name         string
		returnBook   models.Book
		inputId      int
		mockBehavior MockBehavior
		returnErr    bool
	}{
		{
			name:       "OK",
			returnBook: models.Book{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1},
			inputId:    1,
			mockBehavior: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectBegin()
				mock.ExpectQuery("select").
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "genre", "amount"}).
						AddRow(1, "OK", 1, 1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:      "No such id",
			inputId:   1422,
			returnErr: true,
			mockBehavior: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectBegin()
				mock.ExpectQuery("select").
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "genre", "amount"}).
						AddRow(0, "", 0, 0, 0).RowError(0, errors.New("no such id")))
				mock.ExpectRollback()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.inputId)

			book, err := db.GetBookById(test.inputId)
			if test.returnErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.returnBook, book)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDatabasePostgres_DelBook(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error in creating mock: %s", err)
	}
	db := DatabasePostgres{Conn: conn}
	type MockBehavior func(mock sqlmock.Sqlmock, id int)
	tests := []struct {
		name         string
		inputId      int
		mockBehavior MockBehavior
		returnErr    bool
	}{
		{
			name:    "OK",
			inputId: 1,
			mockBehavior: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectBegin()
				mock.ExpectExec("delete").
					WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:      "No such id",
			inputId:   1422,
			returnErr: true,
			mockBehavior: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectBegin()
				mock.ExpectExec("delete").
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.inputId)

			err := db.DelBook(test.inputId)
			if test.returnErr {
				assert.Error(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDatabasePostgres_UpdateBook(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error in creating mock: %s", err)
	}
	db := DatabasePostgres{Conn: conn}
	type MockBehavior func(mock sqlmock.Sqlmock, id int, book models.Book)
	tests := []struct {
		name         string
		inputBook    models.Book
		inputId      int
		mockBehavior MockBehavior
		returnErr    bool
	}{
		{
			name:      "OK",
			inputBook: models.Book{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1},
			inputId:   1,
			mockBehavior: func(mock sqlmock.Sqlmock, id int, book models.Book) {
				mock.ExpectBegin()
				mock.ExpectExec("update books").
					WithArgs(book.Name, book.Price, book.Genre, book.Amount, id).
					WillReturnResult(sqlmock.NewResult(int64(id), 1))
				mock.ExpectCommit()
			},
		},
		{
			name:      "No such id",
			inputId:   1422,
			returnErr: true,
			inputBook: models.Book{ID: 1422, Name: "OK", Price: 1, Genre: 1, Amount: 1},
			mockBehavior: func(mock sqlmock.Sqlmock, id int, book models.Book) {
				mock.ExpectBegin()
				mock.ExpectExec("update books").
					WithArgs(book.Name, book.Price, book.Genre, book.Amount, id).
					WillReturnResult(sqlmock.NewResult(int64(id), 0))
				mock.ExpectCommit()
			},
		},
		{
			name:      "Duplicate name",
			inputId:   0,
			returnErr: true,
			mockBehavior: func(mock sqlmock.Sqlmock, id int, book models.Book) {
				mock.ExpectBegin()
				mock.ExpectExec("update books set").
					WithArgs(book.Name, book.Price, book.Genre, book.Amount, id).
					WillReturnResult(sqlmock.NewResult(int64(id), 0)).
					WillReturnError(errors.New("duplicate name"))
				mock.ExpectRollback()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.inputId, test.inputBook)

			err := db.UpdateBook(test.inputId, test.inputBook)
			if test.returnErr {
				assert.Error(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
