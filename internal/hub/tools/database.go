package tools

type DatabaseInterface interface {
	ConnectDatabase() error
}

func GetDatabase() (DatabaseInterface, error) {
	var database DatabaseInterface = &PostgreSQL{}

	var err error = database.ConnectDatabase()
	if err != nil {
		return nil, err
	}

	return database, nil
}
