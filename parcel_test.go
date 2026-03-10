package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Address, got.Address)
	require.Equal(t, parcel.Status, got.Status)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, _ := store.Add(parcel)

	newAddress := "new address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	got, _ := store.Get(id)
	require.Equal(t, newAddress, got.Address)
	_ = store.Delete(id)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, _ := store.Add(parcel)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	got, _ := store.Get(id)
	require.Equal(t, ParcelStatusSent, got.Status)
	_ = store.Delete(id)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	client := randRange.Intn(10_000_000)

	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	parcelMap := map[int]Parcel{}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Client = client
		id, _ := store.Add(parcels[i])
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	for _, p := range storedParcels {
		orig, ok := parcelMap[p.Number]
		require.True(t, ok)
		require.Equal(t, orig.Client, p.Client)
		require.Equal(t, orig.Address, p.Address)
		require.Equal(t, orig.Status, p.Status)
		_ = store.Delete(p.Number)
	}
}
