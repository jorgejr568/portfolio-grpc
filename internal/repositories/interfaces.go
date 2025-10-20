package repositories

type rowScanner interface {
	Scan(dest ...any) error
}
