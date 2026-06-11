package models

// RetrieveAll returns all models for database migrations
func RetrieveAll() []interface{} {
	return []interface{}{
		&Account{},
		&Flake{},
		&Transaction{},
		&TransactionHistory{},
		&OutboxEvent{},
	}
}
