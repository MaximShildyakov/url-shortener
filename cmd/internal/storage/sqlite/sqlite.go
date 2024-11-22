package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/MaximShildyakov/url-shortener/cmd/internal/storage"
	"github.com/mattn/go-sqlite3"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct{
	db *sql.DB
}

func New(storagePath string) (*Storage, error){
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil{
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil{
		return nil, fmt.Errorf("#{op}: #{err}")
	}

	_, err = stmt.Exec()
	if err != nil{
		return nil, fmt.Errorf("#{op}: #{err}")
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error){
	const fn = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil{
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil{
		// TODO: refactoring
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique{
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil{
		return 0, fmt.Errorf("%s: failed to get last inserted id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error){
	const fn = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil{
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error{
	const fn = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	result, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get affected rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
	}

	return nil
}

// type Storage struct {
// 	db         *sql.DB
// 	saveStmt   *sql.Stmt
// 	getStmt    *sql.Stmt
// 	deleteStmt *sql.Stmt
// }

// func New(storagePath string) (*Storage, error) {
// 	const fn = "storage.sqlite.New"

// 	db, err := sql.Open("sqlite3", storagePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	// Создаем таблицу
// 	if err := createTable(db); err != nil {
// 		return nil, err
// 	}

// 	// Подготавливаем все statements
// 	saveStmt, err := db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: prepare save stmt: %w", fn, err)
// 	}

// 	getStmt, err := db.Prepare("SELECT url FROM url WHERE alias = ?")
// 	if err != nil {
// 		saveStmt.Close()
// 		return nil, fmt.Errorf("%s: prepare get stmt: %w", fn, err)
// 	}

// 	deleteStmt, err := db.Prepare("DELETE FROM url WHERE alias = ?")
// 	if err != nil {
// 		saveStmt.Close()
// 		getStmt.Close()
// 		return nil, fmt.Errorf("%s: prepare delete stmt: %w", fn, err)
// 	}

// 	return &Storage{
// 		db:         db,
// 		saveStmt:   saveStmt,
// 		getStmt:    getStmt,
// 		deleteStmt: deleteStmt,
// 	}, nil
// }

// func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
// 	const fn = "storage.sqlite.SaveURL"

// 	res, err := s.saveStmt.Exec(urlToSave, alias)
// 	if err != nil {
// 		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
// 			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
// 		}
// 		return 0, fmt.Errorf("%s: %w", fn, err)
// 	}

// 	id, err := res.LastInsertId()
// 	if err != nil {
// 		return 0, fmt.Errorf("%s: failed to get last inserted id: %w", fn, err)
// 	}

// 	return id, nil
// }

// func (s *Storage) GetURL(alias string) (string, error) {
// 	const fn = "storage.sqlite.GetURL"

// 	var resURL string
// 	err := s.getStmt.QueryRow(alias).Scan(&resURL)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return "", storage.ErrURLNotFound
// 		}
// 		return "", fmt.Errorf("%s: execute statement: %w", fn, err)
// 	}

// 	return resURL, nil
// }

// func (s *Storage) DeleteURL(alias string) error {
// 	const fn = "storage.sqlite.DeleteURL"

// 	result, err := s.deleteStmt.Exec(alias)
// 	if err != nil {
// 		return fmt.Errorf("%s: execute statement: %w", fn, err)
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("%s: failed to get affected rows: %w", fn, err)
// 	}

// 	if rowsAffected == 0 {
// 		return fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
// 	}

// 	return nil
// }

// // Важно добавить метод Close для освобождения ресурсов
// func (s *Storage) Close() error {
// 	if err := s.saveStmt.Close(); err != nil {
// 		return err
// 	}
// 	if err := s.getStmt.Close(); err != nil {
// 		return err
// 	}
// 	if err := s.deleteStmt.Close(); err != nil {
// 		return err
// 	}
// 	return s.db.Close()
// }

// // Вспомогательная функция для создания таблицы
// func createTable(db *sql.DB) error {
// 	_, err := db.Exec(`
// 	CREATE TABLE IF NOT EXISTS url(
// 		id INTEGER PRIMARY KEY,
// 		alias TEXT NOT NULL UNIQUE,
// 		url TEXT NOT NULL);
// 	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
// 	`)
// 	return err
// }