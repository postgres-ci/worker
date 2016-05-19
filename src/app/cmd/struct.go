package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type plpgsqlTest struct {
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

type plpgsqlTests []plpgsqlTest

type test struct {
	Function string  `json:"function"`
	Errors   errors  `json:"errors"`
	Duration float64 `json:"duration"`
}

func (p plpgsqlTests) Value() (driver.Value, error) {

	var tests []test

	for _, plpgsql := range p {

		tests = append(tests, test{
			Function: fmt.Sprintf("%s.%s", plpgsql.Namespace, plpgsql.Procedure),
			Errors:   plpgsql.Errors,
			Duration: plpgsql.FinishedAt.Sub(plpgsql.StartedAt).Seconds(),
		})
	}

	return json.Marshal(tests)
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
