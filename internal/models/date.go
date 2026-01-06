package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Date - кастомный тип для работы с датами без времени
type Date struct {
	time.Time
}

// MarshalJSON - сериализация в JSON как "YYYY-MM-DD"
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}

// UnmarshalJSON - десериализация из JSON формата "YYYY-MM-DD"
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// Убираем кавычки
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	parsed, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

// Value - для записи в базу данных
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

// Scan - для чтения из базы данных
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into Date", value)
}
