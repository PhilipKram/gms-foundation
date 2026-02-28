// Package dbutil provides helpers for opening and configuring database/sql
// connection pools with sensible defaults.
//
// The package imports the MySQL driver (github.com/go-sql-driver/mysql)
// as a side effect. Use Open with a different driverName for other databases
// that have been registered separately.
package dbutil
