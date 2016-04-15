package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type test struct {
	Namespace  string    `json:"namespace"   db:"namespace"`
	Procedure  string    `json:"procedure"   db:"procedure"`
	Errors     errors    `json:"errors"      db:"errors"`
	StartedAt  time.Time `json:"started_at"  db:"started_at"`
	FinishedAt time.Time `json:"finished_at" db:"finished_at"`
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
