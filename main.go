package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
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

type ParcelStore interface {
	Add(p Parcel) (int, error)
	Get(number int) (Parcel, error)
	GetByClient(client int) ([]Parcel, error)
	SetStatus(number int, status string) error
	SetAddress(number int, address string) error
	Delete(number int) error
}

type SQLParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS parcels (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	return &SQLParcelStore{db: db}
}

func (s *SQLParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		"INSERT INTO parcels (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (s *SQLParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcels WHERE number = :number",
		sql.Named("number", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

func (s *SQLParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcels WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}

func (s *SQLParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		"UPDATE parcels SET status = @status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	return err
}

func (s *SQLParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec(
		"UPDATE parcels SET address = :address WHERE number = :number",
		sql.Named("address", address),
		sql.Named("number", number))
	return err
}

func (s *SQLParcelStore) Delete(number int) error {
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return fmt.Errorf("посылку можно удалить только если статус 'registered'")
	}
	_, err = s.db.Exec(
		"DELETE FROM parcels WHERE number = :number",
		sql.Named("number", number))
	return err
}

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}
	parcel.Number = id
	fmt.Printf("Новая посылка №%d зарегистрирована на адрес %s\n", parcel.Number, parcel.Address)
	return parcel, nil
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}
	fmt.Printf("Посылки клиента %d:\n", client)
	for _, p := range parcels {
		fmt.Printf("№%d: адрес %s, статус %s, зарегистрирована %s\n", p.Number, p.Address, p.Status, p.CreatedAt)
	}
	fmt.Println()
	return nil
}

func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}
	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		fmt.Printf("Посылка №%d уже доставлена, статус не изменился\n", number)
		return nil
	}
	fmt.Printf("У посылки №%d новый статус: %s\n", number, nextStatus)
	return s.store.SetStatus(number, nextStatus)
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}
	if parcel.Status != ParcelStatusRegistered {
		return fmt.Errorf("можно менять адрес только для посылки со статусом 'registered'")
	}
	return s.store.SetAddress(number, address)
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	db, err := sql.Open("sqlite", "./tracker.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		log.Fatal(err)
	}

	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	if err := service.ChangeAddress(p.Number, newAddress); err != nil {
		fmt.Println("Ошибка при смене адреса:", err)
	} else {
		fmt.Printf("Адрес посылки №%d успешно изменён на %s\n", p.Number, newAddress)
	}
}
