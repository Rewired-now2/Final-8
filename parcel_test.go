package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, stored.Client)
	assert.Equal(t, parcel.Status, stored.Status)
	assert.Equal(t, parcel.Address, stored.Address)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, got.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newStatus := "sent"
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newStatus, stored.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	for i := 0; i < len(parcels); i++ {
		parcels[i].Client = client
	}

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	for _, p := range storedParcels {
		orig, ok := parcelMap[p.Number]
		assert.True(t, ok, "посылка с ID %d не найдена в map", p.Number)
		assert.Equal(t, orig.Client, p.Client)
		assert.Equal(t, orig.Status, p.Status)
		assert.Equal(t, orig.Address, p.Address)
		assert.Equal(t, orig.CreatedAt, p.CreatedAt)
	}
}
