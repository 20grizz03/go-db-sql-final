package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	res, err := store.Add(parcel)
	require.Nil(t, err, "error not nil")
	assert.NotNil(t, res, "id not found")
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res2, err := store.Get(res)
	require.Nil(t, err, "error not nil")
	assert.Equal(t, res2.Status, parcel.Status, "objects do not match")
	assert.Equal(t, res2.Client, parcel.Client, "objects do not match")
	assert.Equal(t, res2.Address, parcel.Address, "objects do not match")
	assert.Equal(t, res2.CreatedAt, parcel.CreatedAt, "objects do not match")
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	err = store.Delete(res)
	require.Nil(t, err, "error not nil")
	_, err = store.Get(res)
	assert.NotNil(t, err, "error not nil")
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := store.Add(parcel)
	require.Nil(t, err, "error not nil")
	assert.NotNil(t, res, "id not found")
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(res, newAddress)
	require.Nil(t, err, "error not nil")
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	res2, err := store.Get(res)
	assert.Equal(t, res2.Address, newAddress, "Adresses not match")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := store.Add(parcel)
	require.Nil(t, err, "error not nil")
	assert.NotNil(t, res, "id not found")
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := "sent"
	err = store.SetStatus(res, newStatus)
	require.Nil(t, err, "error not nil")
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	res2, err := store.Get(res)
	assert.Equal(t, res2.Status, newStatus, "status not match")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	require.Nil(t, err, "failed to open database")
	defer db.Close() // Закрываем базу данных после завершения теста

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// Задаём всем посылкам один и тот же идентификатор клиента
	client := rand.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// Добавляем посылки в базу данных и сохраняем их идентификаторы в карте
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.Nil(t, err, "error while adding parcel")

		// Обновляем идентификатор посылки
		parcels[i].Number = id

		// Сохраняем добавленную посылку в карту для дальнейшей проверки
		parcelMap[id] = parcels[i]
	}

	// Получаем список посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client) // Здесь исправлено на GetByClient
	require.Nil(t, err, "error while getting parcels by client")

	// Убеждаемся, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(storedParcels), len(parcelMap), "number of stored parcels does not match")

	// Проверяем, что все полученные посылки есть в parcelMap и их поля совпадают
	for _, parcel := range storedParcels {
		// Убеждаемся, что посылка есть в карте по её идентификатору
		expectedParcel, exists := parcelMap[parcel.Number]
		assert.True(t, exists, "parcel not found in map")

		// Проверяем, что поля совпадают
		assert.Equal(t, expectedParcel.Client, parcel.Client, "client does not match")
		assert.Equal(t, expectedParcel.Status, parcel.Status, "status does not match")
		assert.Equal(t, expectedParcel.Address, parcel.Address, "address does not match")
	}
}
