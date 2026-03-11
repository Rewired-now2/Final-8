package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec(
		`INSERT INTO parcel(client, status, address, created_at)
		 VALUES (:client, :status, :address, :created_at)`,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
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

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
		 FROM parcel
		 WHERE number = :number`,
		sql.Named("number", number),
	)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
		 FROM parcel
		 WHERE client = :client`,
		sql.Named("client", client),
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
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		`UPDATE parcel
		 SET status = :status
		 WHERE number = :number`,
		sql.Named("status", status),
		sql.Named("number", number),
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var status string
	row := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = :number`,
		sql.Named("number", number),
	)
	if err := row.Scan(&status); err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("изменить адрес можно только для посылки со статусом registered")
	}

	_, err := s.db.Exec(
		`UPDATE parcel
		 SET address = :address
		 WHERE number = :number`,
		sql.Named("address", address),
		sql.Named("number", number),
	)
	return err
}

func (s ParcelStore) Delete(number int) error {
	var status string
	row := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = :number`,
		sql.Named("number", number),
	)
	if err := row.Scan(&status); err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("удалять можно только посылки со статусом registered")
	}

	_, err := s.db.Exec(
		`DELETE FROM parcel WHERE number = :number`,
		sql.Named("number", number),
	)
	return err
}
