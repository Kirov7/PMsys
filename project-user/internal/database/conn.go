package gorms

type DbConn interface {
	Begin()
	Rollback()
	Commit()
}
