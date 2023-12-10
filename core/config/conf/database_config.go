package conf

type DatabaseConfig struct {
	Dialector interface{}
	Dsn       string
}
