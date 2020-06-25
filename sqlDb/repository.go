package sqlDb

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

type MigrationRepository struct {
	migrator *migrate.Migrate
	fileDir  string
	log      *zap.SugaredLogger
}

func NewMigrationRepository(log *zap.SugaredLogger, sqlConnStr, fileDir string) MigrationRepository {
	m, err := migrate.New(fmt.Sprintf("file://%s", fileDir), sqlConnStr)

	if err != nil {
		panic(err)
	}

	s := MigrationRepository{
		migrator: m,
		fileDir:  fileDir,
		log:      log,
	}

	return s
}

func (mr MigrationRepository) Migrate() (*devpb.MigrationStatus, error) {
	mr.log.Infof("Migrating...")

	status, err := mr.Status()

	if err != nil {
		return status, err
	}

	if status.Dirty {
		spew.Dump(status)
		return status, errors.New("database status is dirty, this should never happen")
	}

	totalSteps := 0

	migrationErr := mr.migrator.Steps(1)
	for migrationErr == nil {
		totalSteps++
		fmt.Printf("Migrated up one step... Total: %d\n", totalSteps)

		migrationErr = mr.migrator.Steps(1)
	}

	if migrationErr.Error() != "file does not exist" {
		fmt.Println("Rolling back one version due to error")
		errMsg := fmt.Sprintf("failed to run migrations, err: %v\n", migrationErr)

		_ = mr.migrator.Force(int(status.Version) + totalSteps - 1)

		return status, errors.New(errMsg)
	}

	fmt.Printf("Migrated a total of %d steps\n", totalSteps)
	return mr.Status()
}

func (mr MigrationRepository) Status() (*devpb.MigrationStatus, error) {
	version, dirty, err := mr.migrator.Version()
	latestVersion := mr.getLatestVersion()

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

func (mr MigrationRepository) ForceVersion(version int32) (*devpb.MigrationStatus, error) {
	err := mr.migrator.Force(int(version))

	if err != nil {
		return &devpb.MigrationStatus{}, err
	}

	return mr.Status()
}

func (mr MigrationRepository) Rollback() (*devpb.MigrationStatus, error) {
	status, err := mr.Status()

	if err != nil {
		return status, err
	}

	if status.Dirty {
		spew.Dump(status)
		return status, errors.New("database status is dirty, this should never happen")
	}

	totalSteps := 0

	migrationErr := mr.migrator.Steps(-1)
	for migrationErr == nil {
		totalSteps--
		fmt.Printf("Rolled back one step... Total: %d\n", totalSteps)

		migrationErr = mr.migrator.Steps(-1)
	}

	if migrationErr.Error() != "file does not exist" {
		fmt.Println("Rolling back one version due to error")
		errMsg := fmt.Sprintf("failed to run migrations, err: %v\n", migrationErr)

		_, _ = mr.ForceVersion(int32(status.Version) + int32(totalSteps) + 1)

		return status, errors.New(errMsg)
	}

	return mr.Status()
}

func (mr MigrationRepository) getLatestVersion() int32 {
	files, err := ioutil.ReadDir(mr.fileDir)
	if err != nil {
		return 0
	}

	count := 0

	for range files {
		count++
	}

	return int32(count / 2)
}
