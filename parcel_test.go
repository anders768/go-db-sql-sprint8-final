package main

import (
	"database/sql"
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
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД: %v", err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка функции Add() в функции TestAddGetDelete(): %v", err)
	require.NotEmpty(t, id, "Идентификатор посылки не должен быть пустым")
	parcel.Number = id // добавляем идентификатор в посылку

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	parcel2, err := store.Get(id)
	require.NoError(t, err, "Ошибка функции Get() в функции TestAddGetDelete(): %v", err)
	require.Equal(t, parcel, parcel2, `Ошибка функции Get() в функции TestAddGetDelete(): 
		неравенство полей добавленной и полученной посылок`) // через сравнение структур

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err, "Ошибка функции Delete() в функции TestAddGetDelete(): %v", err)
	// пытаемся получить удалённую посылку
	_, err = store.Get(id)
	// проверяем, что вернулась ошибка sql.ErrNoRows
	require.ErrorIs(t, err, sql.ErrNoRows, `Функция Get() (в проверке функции Delete()) в функции TestAddGetDelete() 
		должна была вернуть ошибку sql.ErrNoRows, а не %v`, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД: %v", err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка функции Add() в функции TestSetAddress(): %v", err)
	require.NotEmpty(t, id, "Идентификатор посылки не должен быть пустым")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка функции SetAddress() в функции TestSetAddress(): %v", err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	parcel2, err := store.Get(id)
	require.NoError(t, err, "Ошибка функции Get() (в проверке функции SetAddress()) в функции TestSetAddress(): %v", err)
	require.Equal(t, newAddress, parcel2.Address, `Ошибка функции Get() (в проверке функции SetAddress()) в функции 
		TestSetAddress(): адрес посылки не обновился`)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД: %v", err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка функции Add() в функции TestSetAddress(): %v", err)
	require.NotEmpty(t, id, "Идентификатор посылки не должен быть пустым")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "Ошибка функции SetStatus() в функции TestSetStatus(): %v", err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	parcel2, err := store.Get(id)
	require.NoError(t, err, "Ошибка функции Get() (в проверке функции SetStatus()) в функции TestSetStatus(): %v", err)
	require.Equal(t, newStatus, parcel2.Status, `Ошибка функции Get() (в проверке функции SetStatus()) в функции 
		TestSetStatus(): статус посылки не обновился`)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД: %v", err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{ // parcels - слайс из 3 тестовых посылок, которые отличаются только временем создания
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{} // пока пустая мапа для хранения добавленных посылок

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000) // client - целое число в диапазоне [0; 10 млн)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка функции Add() в функции TestGetByClient(): %v", err)
		require.NotEmpty(t, id, "Идентификатор посылки не должен быть пустым")

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в мапу map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err, "Ошибка функции GetByClient() в функции TestGetByClient(): %v", err)
	require.Equal(t, len(parcels), len(storedParcels), `Ошибка функции GetByClient() в функции TestGetByClient(): 
		неравенство количества полученных и добавленных посылок`)

	// check
	for _, parcel := range storedParcels {
		// в parcelMap (мапе) лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels (слайс с полученными) есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		// в parcel - конкретная посылка из полученных
		parcelExpected, ok := parcelMap[parcel.Number] // ищем в мапе добавленных посылку с идентификатором, как у полученной
		require.True(t, ok, "Ошибка проверки посылки в функции TestGetByClient(): полученной нет среди добавленных")
		require.Equal(t, parcelExpected, parcel, `Ошибка проверки посылки в функции TestGetByClient(): 
		поля добавленной и полученной посылок не равны`) // через сравнение структур
	}
}
