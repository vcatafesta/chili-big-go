package service

import (
	"database/sql"
)

type Book struct {
	ID     int
	Title  string
	Author string
	Genre  string
}

type BookService struct {
	db *sql.DB
}

func (s *BookService) CreateBook(book *Book) error {
	query := "'Insert into books (title, author, genre) values(?,?,?)'"
	result, err := s.db.Exec(query, book.Title, book.Author, book.Genre)
	if err != nil {
		return err
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	book.ID = int(lastInsertID)
	return nil
}

func (b Book) GetFullBook() string {
	return b.Title + " by " + b.Author
}
