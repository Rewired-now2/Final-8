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
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, InitDB(db))
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	t.Cleanup(func() {
		err := store.Delete(id)
		require.NoError(t, err)
	})

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Address, got.Address)
	require.Equal(t, parcel.Status, got.Status)
}

func TestSetAddressAndStatus(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, InitDB(db))
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Delete(id)) })

	newAddress := "new address"
	require.NoError(t, store.SetAddress(id, newAddress))

	newStatus := ParcelStatusSent
	require.NoError(t, store.SetStatus(id, newStatus))

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
	require.Equal(t, newStatus, got.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, InitDB(db))
	store := NewParcelStore(db)
	client := randRange.Intn(10_000_000)

	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	parcelMap := map[int]Parcel{}

	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]

		t.Cleanup(func(id int) func() {
			return func() { require.NoError(t, store.Delete(id)) }
		}(id))
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, p := range storedParcels {
		orig, ok := parcelMap[p.Number]
		require.True(t, ok)
		require.Equal(t, orig.Client, p.Client)
		require.Equal(t, orig.Address, p.Address)
		require.Equal(t, orig.Status, p.Status)
	}
}
