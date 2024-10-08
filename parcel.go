package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("insert into parcel (client, status, address, created_at) values (:Client, :Status, :Address, :CreatedAt)",
		sql.Named("Client", p.Client),
		sql.Named("Status", p.Status),
		sql.Named("Address", p.Address),
		sql.Named("CreatedAt", p.CreatedAt))
	// верните идентификатор последней добавленной записи
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Возвращаем идентификатор
	return int(lastID), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// Выполняем запрос на получение данных по указанному number
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number)

	// Заполняем объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	// Если возникла ошибка (например, записи нет), возвращаем её
	if err != nil {
		if err == sql.ErrNoRows {
			// Если запись не найдена
			return Parcel{}, fmt.Errorf("parcel with number %d not found", number)
		}
		// Любая другая ошибка
		return Parcel{}, err
	}

	// Возвращаем заполненный объект Parcel
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Выполняем запрос на получение данных по указанному client
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Срез для хранения результатов
	var parcels []Parcel

	// Проходим по всем строкам и заполняем срез Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}

	// Проверяем на ошибки, которые могли произойти при итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Возвращаем срез с результатами
	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// Выполняем запрос на обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return err // Возвращаем ошибку, если обновление не удалось
	}

	// Если всё прошло успешно, возвращаем nil
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Проверяем статус посылки по её номеру
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d not found", number)
		}
		return err
	}

	// Проверяем, что статус "registered"
	if status != "registered" {
		return fmt.Errorf("cannot change address, parcel status is not 'registered'")
	}

	// Если статус "registered", обновляем адрес
	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return err
	}

	// Возвращаем nil, если операция выполнена успешно
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// Проверяем статус посылки по её номеру
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d not found", number)
		}
		return err
	}

	// Проверяем, что статус "registered"
	if status != "registered" {
		return err
	}

	// Если статус "registered", удаляем запись
	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	if err != nil {
		return err
	}

	// Возвращаем nil, если операция выполнена успешно
	return nil
}
