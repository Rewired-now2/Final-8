package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка подключения к базе:", err)
		return
	}
	defer db.Close()

	if err := InitDB(db); err != nil {
		fmt.Println("Ошибка инициализации базы:", err)
		return
	}

	store := NewParcelStore(db)
	clientID := 1

	p1 := Parcel{
		Client:    clientID,
		Status:    ParcelStatusRegistered,
		Address:   "Псков, д. Пушкина, ул. Колотушкина, д. 5",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id1, err := store.Add(p1)
	if err != nil {
		fmt.Println("Ошибка добавления посылки:", err)
		return
	}
	p1.Number = id1
	fmt.Printf(
		"Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		p1.Number, p1.Address, p1.Client, p1.CreatedAt,
	)

	if err := store.SetStatus(id1, ParcelStatusSent); err != nil {
		fmt.Println("Ошибка изменения статуса:", err)
	} else {
		got, _ := store.Get(id1)
		fmt.Printf("У посылки № %d новый статус: %s\n", got.Number, got.Status)
	}

	parcels, _ := store.GetByClient(clientID)
	fmt.Printf("Посылки клиента №%d:\n", clientID)
	for _, p := range parcels {
		fmt.Printf(
			"Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			p.Number, p.Address, p.Client, p.CreatedAt, p.Status,
		)
	}

	p2 := Parcel{
		Client:    clientID,
		Status:    ParcelStatusRegistered,
		Address:   "Псков, д. Пушкина, ул. Колотушкина, д. 5",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id2, err := store.Add(p2)
	if err != nil {
		fmt.Println("Ошибка добавления посылки:", err)
		return
	}
	p2.Number = id2
	fmt.Printf("\nНовая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		p2.Number, p2.Address, p2.Client, p2.CreatedAt)

	parcels, _ = store.GetByClient(clientID)
	fmt.Printf("Посылки клиента №%d:\n", clientID)
	for _, p := range parcels {
		fmt.Printf(
			"Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			p.Number, p.Address, p.Client, p.CreatedAt, p.Status,
		)
	}
}
