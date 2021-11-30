package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/porky256/rest-api/models"
	"log"
	"strconv"
)

const (
	Host = "db"
	Port = 5432
)

type Database interface {
	GetAllBooks(filter map[string][]string) ([]models.Book, error)
	AddBook(book models.Book) (int, error)
	DelBook(id int) error
	UpdateBook(id int, book models.Book) error
	GetBookById(id int) (models.Book, error)
}

type DatabasePostgres struct {
	Conn *sql.DB
}

func Initialize(username, password, database string) (DatabasePostgres, error) {
	db := DatabasePostgres{}
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", Host, Port, username, password, database)
	conn, err := sql.Open("postgres", connectionString)

	if err != nil {
		return db, err
	}

	db.Conn = conn
	err = db.Conn.Ping()

	if err != nil {
		return db, err
	}

	log.Println("Database connection established")
	return db, nil
}

func (db *DatabasePostgres) GetAllBooks(filter map[string][]string) ([]models.Book, error) {

	tx, err := db.Conn.Begin()
	if err != nil {
		return []models.Book{}, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	list := []models.Book{}
	var rows *sql.Rows
	var query string
	var queryinfo []interface{}
	if len(filter) > 0 {
		_, hasName := filter["name"]
		_, hasGenre := filter["genre"]
		if hasName && hasGenre {
			query = "select * from books where genre=$1 and name=$2 and amount>0 order by id desc;"
			genre, err := strconv.Atoi(filter["genre"][0])
			if err != nil {
				return list, err
			}
			queryinfo = append(queryinfo, genre, filter["name"][0])
		} else if hasName {
			query = "select * from books where name=$1 and amount>0 order by id desc;"
			queryinfo = append(queryinfo, filter["name"][0])
		} else if hasGenre {
			query = "select * from books where genre=$1 and amount>0 order by id desc;"
			genre, err := strconv.Atoi(filter["genre"][0])
			if err != nil {
				return list, err
			}
			queryinfo = append(queryinfo, genre)
		}
	} else {
		query = "select * from books where amount>0 order by id desc;"
	}
	rows, err = tx.Query(query, queryinfo...)
	if err != nil {
		return list, err
	}

	if rows != nil {
		for rows.Next() {
			var book models.Book
			err := rows.Scan(&book.ID, &book.Name, &book.Price, &book.Genre, &book.Amount)
			if err != nil {
				return list, err
			}
			list = append(list, book)
		}
	}
	return list, err
}
func (db *DatabasePostgres) AddBook(book models.Book) (int, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	var id int
	query := "insert into books (name,price,genre,amount) values ($1, $2, $3, $4) returning id;"
	err = tx.QueryRow(query, book.Name, book.Price, book.Genre, book.Amount).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DatabasePostgres) DelBook(id int) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	query := "delete from books where id =$1;"
	res, err := tx.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DatabasePostgres) UpdateBook(id int, book models.Book) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()
	query := "update books set name=$1, price=$2, genre=$3, amount=$4 where id=$5;"
	res, err := tx.Exec(query, book.Name, book.Price, book.Genre, book.Amount, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DatabasePostgres) GetBookById(id int) (models.Book, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return models.Book{}, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	var book models.Book
	query := "select * from books where id =$1;"
	row := tx.QueryRow(query, id)
	err = row.Scan(&book.ID, &book.Name, &book.Price, &book.Genre, &book.Amount)
	return book, err
}
