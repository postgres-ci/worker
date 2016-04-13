package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type test struct {
	Namespace  string    `db:"namespace"`
	Procedure  string    `db:"procedure"`
	Errors     errors    `db:"errors"`
	StartedAt  time.Time `db:"started_at"`
	FinishedAt time.Time `db:"finished_at"`
}

type err struct {
	Message string `json:"message"`
	Comment string `json:"comment"`
}

type tests []test

func (t tests) Value() (driver.Value, error) {

	return json.Marshal(t)
}

type errors []err

func (e *errors) Scan(src interface{}) error {

	var source []byte

	switch src.(type) {

	case string:

		source = []byte(src.(string))

	case []byte:

		source = src.([]byte)
	}

	return json.Unmarshal(source, &e)
}
