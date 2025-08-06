module github.com/sktylr/routeit/examples/todo

go 1.24.4

// Required while this is not published
replace github.com/sktylr/routeit => ../../src

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/uuid v1.6.0
	github.com/sktylr/routeit v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.40.0
)

require filippo.io/edwards25519 v1.1.0 // indirect
