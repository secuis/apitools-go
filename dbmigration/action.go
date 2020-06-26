package dbmigration

import (
	"dev.azure.com/securitasintelligentservices/insights/_git/sispbgo.git/sis/rp/devpb"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"io/ioutil"
)

type Migrator struct {
	migrator *migrate.Migrate
	fileDir  string
	log      *zap.SugaredLogger
}

func NewMigrator(log *zap.SugaredLogger, sqlConnStr, fileDir string) Migrator {
	m, err := migrate.New(fmt.Sprintf("file://%s", fileDir), sqlConnStr)

	if err != nil {
		panic(err)
	}

	s := Migrator{
		migrator: m,
		fileDir:  fileDir,
		log:      log,
	}

	return s
}

func (m Migrator) Migrate() (*devpb.MigrationStatus, error) {
	m.log.Infof("Migrating...")

	status, err := m.Status()

	if err != nil {
		return status, err
	}

	if status.Dirty {
		spew.Dump(status)
		return status, errors.New("database status is dirty, this should never happen")
	}

	totalSteps := 0

	migrationErr := m.migrator.Steps(1)
	for migrationErr == nil {
		totalSteps++
		fmt.Printf("Migrated up one step... Total: %d\n", totalSteps)

		migrationErr = m.migrator.Steps(1)
	}

	if migrationErr.Error() != "file does not exist" {
		fmt.Println("Rolling back one version due to error")
		errMsg := fmt.Sprintf("failed to run migrations, err: %v\n", migrationErr)

		_ = m.migrator.Force(int(status.Version) + totalSteps - 1)

		return status, errors.New(errMsg)
	}

	fmt.Printf("Migrated m total of %d steps\n", totalSteps)
	return m.Status()
}

func (m Migrator) Status() (*devpb.MigrationStatus, error) {
	version, dirty, err := m.migrator.Version()
	latestVersion := m.getLatestVersion()

	if err != nil {
		if err == migrate.ErrNilVersion {
			return &devpb.MigrationStatus{
				Version:       0,
				LatestVersion: latestVersion,
				UpToDate:      false,
				Dirty:         false,
			}, nil
		}
		return &devpb.MigrationStatus{}, err
	}

	upToDate := int32(version) == latestVersion

	return &devpb.MigrationStatus{
		Version:       uint32(version),
		LatestVersion: latestVersion,
		UpToDate:      upToDate,
		Dirty:         dirty,
	}, nil
}

func (m Migrator) ForceVersion(version int32) (*devpb.MigrationStatus, error) {
	err := m.migrator.Force(int(version))

	if err != nil {
		return &devpb.MigrationStatus{}, err
	}

	return m.Status()
}

func (m Migrator) Rollback() (*devpb.MigrationStatus, error) {
	status, err := m.Status()

	if err != nil {
		return status, err
	}

	if status.Dirty {
		spew.Dump(status)
		return status, errors.New("database status is dirty, this should never happen")
	}

	totalSteps := 0

	migrationErr := m.migrator.Steps(-1)
	for migrationErr == nil {
		totalSteps--
		fmt.Printf("Rolled back one step... Total: %d\n", totalSteps)

		migrationErr = m.migrator.Steps(-1)
	}

	if migrationErr.Error() != "file does not exist" {
		fmt.Println("Rolling back one version due to error")
		errMsg := fmt.Sprintf("failed to run migrations, err: %v\n", migrationErr)

		_, _ = m.ForceVersion(int32(status.Version) + int32(totalSteps) + 1)

		return status, errors.New(errMsg)
	}

	return m.Status()
}

func (m Migrator) getLatestVersion() int32 {
	files, err := ioutil.ReadDir(m.fileDir)
	if err != nil {
		return 0
	}

	count := 0

	for range files {
		count++
	}

	return int32(count / 2)
}
