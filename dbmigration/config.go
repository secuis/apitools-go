package dbmigration

type MigrationConfig struct {
	migrateOnStart bool
	sqlFileDir     string
	sqlConnStr     string
}
