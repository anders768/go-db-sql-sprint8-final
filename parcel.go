package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB // единственное поле - соединение с базой данных
}

func NewParcelStore(db *sql.DB) ParcelStore { // функция-конструктор
	return ParcelStore{db: db} // возвращает новый экземпляр структуры ParcelStore
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec(`INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status,
		:address, :created_at)`,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	// значение в столбец number должно добавиться само, т.к. там - автоинкрементный первичный ключ
	if err != nil { // если ошибка добавления строки
		return 0, err
	}

	number, err := res.LastInsertId() // получаем идентификатор последней добавленной строки
	if err != nil {                   // если ошибка получения идентификатора
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	return int(number), nil // number имеет тип int64, преобразуем (а что делать, если не влезет по битам?)
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	// заполните объект Parcel данными из таблицы
	p := Parcel{}

	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt) // напихиваем считанное в поля p
	if err != nil {                                                            // Если ошибка считывания результата запроса -
		return Parcel{}, err // возвращаем пустой Parcel и ошибку.
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	// заполните срез Parcel данными из таблицы
	var res []Parcel // пока что это - nil-слайс, т.к. не инициализирован

	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil { // Если ошибка создания и выполнения запроса -
		return nil, err // возвращаем nil-[]Parcel и ошибку.
	}
	defer rows.Close()

	for rows.Next() {
		p := Parcel{} // создаём пустую посылку (её поля будут заполнены из значений считанной строки)

		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt) // напихиваем считанное в поля p
		if err != nil {                                                            // Если ошибка считывания результата запроса -
			return nil, err // возвращаем nil-[]Parcel и ошибку (не уверен, что это правильно).
		}

		res = append(res, p) // пихаем посылку в слайс посылок
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	// Вопрос: какая функция должна определять, что обновление статуса возможно - эта или внешняя?
	// Будем считать, что внешняя.-)

	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil { // если ошибка обновления
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered

	// получаем статус (для проверки на registered)
	var status string

	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&status)
	if err != nil { // Если ошибка считывания статуса -
		return err // возвращаем ошибку. Надо ли ошибку обернуть (для большей понятности)?
	}

	if status != ParcelStatusRegistered { // Если статус не равен "registered" -
		return errors.New("адрес не изменён, т.к. статус не \"registered\"!") // возвращаем ошибку.
	}

	// обновляем адрес
	_, err = s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil { // Если ошибка обновления адреса -
		return err // возвращаем её. Надо ли ошибку обернуть (для большей понятности)?
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	// получаем статус (для проверки на registered)
	var status string

	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&status)
	if err != nil { // Если ошибка считывания статуса -
		return err // возвращаем ошибку. Надо ли ошибку обернуть (для большей понятности)?
	}

	// удаляем строку
	if status == ParcelStatusRegistered { // если статус "registered" - можно удалять
		_, err = s.db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
		if err != nil { // Если ошибка удаления строки -
			return err // возвращаем ошибку.
		}
	}
	return nil
}
