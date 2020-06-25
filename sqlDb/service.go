package sqlDb

import (
	"context"
	"dev.azure.com/securitasintelligentservices/insights/_git/sispbgo.git/sis/rp/devpb"
	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/appengine/log"
	"google.golang.org/grpc/codes"
)

type Migrationservice struct {
	log *zap.SugaredLogger

	mR MigrationRepository
}

func NewMigrationService(log *zap.SugaredLogger, config MigrationConfig) devpb.MigrationServer {
	repo := NewMigrationRepository(log, config.sqlConnStr, config.sqlFileDir)
	s := Migrationservice{
		log: log,
		mR:  repo,
	}

	if config.migrateOnStart {
		migrationStatus, err := repo.Migrate()

		if err != nil {
			panic(err)
		}

		log.Infof("Tried to migrate on boot, status now: %v\n", migrationStatus)
	}

	return s
}

func (d Migrationservice) DatabaseMigrate(ctx context.Context, request *devpb.DatabaseMigrateRequest) (*devpb.DatabaseMigrateResponse, error) {
	migratiosnStatus, err := d.mR.Migrate()

	if err != nil {
		log.Warningf(ctx, "could not migrate database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not migrate database")
	}

	return &devpb.DatabaseMigrateResponse{Status: migratiosnStatus}, nil
}

func (d Migrationservice) DatabaseStatus(ctx context.Context, request *devpb.DatabaseStatusRequest) (*devpb.DatabaseStatusResponse, error) {
	migratiosnStatus, err := d.mR.Status()

	if err != nil {
		log.Warningf(ctx, "could not get database status, err: %v", err)
		return nil, status.Error(codes.Internal, "could not get database status")
	}

	return &devpb.DatabaseStatusResponse{Status: migratiosnStatus}, nil
}

func (d Migrationservice) DatabaseRollback(ctx context.Context, request *devpb.DatabaseRollbackRequest) (*devpb.DatabaseRollbackResponse, error) {
	migratiosnStatus, err := d.mR.Rollback()

	if err != nil {
		log.Warningf(ctx, "could not rollback database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not rollback database")
	}

	return &devpb.DatabaseRollbackResponse{Status: migratiosnStatus}, nil
}

func (d Migrationservice) DatabaseForceVersion(ctx context.Context, request *devpb.DatabaseForceVersionRequest) (*devpb.DatabaseForceVersionResponse, error) {
	migratiosnStatus, err := d.mR.ForceVersion(request.GetVersion())

	if err != nil {
		log.Warningf(ctx, "could not force database version, err: %v", err)
		return nil, status.Error(codes.Internal, "could not force database version")
	}

	return &devpb.DatabaseForceVersionResponse{Status: migratiosnStatus}, nil
}
