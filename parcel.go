package main

import (
	"database/sql"
	"errors"
	"time"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelStore struct {
	db *sql.DB
}

func InitDB(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)
	`)
	return err
}

func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

func (s *ParcelStore) Add(p Parcel) (int, error) {
	if p.CreatedAt == "" {
		p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	result, err := s.db.Exec(
		`INSERT INTO parcel(client, status, address, created_at) VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`,
		number,
	)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, errors.New("посылка не найдена")
		}
		return p, err
	}
	return p, nil
}

func (s *ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`,
		status, number,
	)
	return err
}

func (s *ParcelStore) SetAddress(number int, address string) error {
	var status string
	row := s.db.QueryRow(`SELECT status FROM parcel WHERE number = ?`, number)
	if err := row.Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("посылка не найдена")
		}
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("изменить адрес можно только для посылки со статусом registered")
	}

	_, err := s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ?`,
		address, number,
	)
	return err
}

func (s *ParcelStore) Delete(number int) error {
	var status string
	row := s.db.QueryRow(`SELECT status FROM parcel WHERE number = ?`, number)
	if err := row.Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("посылка не найдена")
		}
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("удалять можно только посылки со статусом registered")
	}

	_, err := s.db.Exec(`DELETE FROM parcel WHERE number = ?`, number)
	return err
}
