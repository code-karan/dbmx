package model

type Postgres struct {
	ID       int64
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Env      string
	Colour   *string
}

type GenericResponse struct {
	OK           bool          `json:"ok"`
	Data         []interface{} `json:"data"`
	RowsAffected int64         `json:"rowsAffected"`
	Message      string        `json:"message"`
}
