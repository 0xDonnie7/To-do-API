package data

import "database/sql"

type Models struct {
	Todos TodosModel
	Users UsersModel
}

func AllModels(db *sql.DB) Models {
	return Models{
		Todos: TodosModel{DB: db},
		Users: UsersModel{DB: db},
	}
}
