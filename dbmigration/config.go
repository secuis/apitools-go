package dbmigration

type MigrationConfig struct {
	MigrateOnStart bool
	SqlFileDir     string
	SqlConnStr     string
}
