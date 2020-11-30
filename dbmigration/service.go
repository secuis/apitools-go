package dbmigration

import (
	"context"

	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/appengine/log"
	"google.golang.org/grpc/codes"
)

type MigrationService struct {
	log *zap.SugaredLogger

	mA Migrator
	UnimplementedMigrationServiceServer
}

func NewMigrationService(log *zap.SugaredLogger, config MigrationConfig) MigrationServiceServer {
	action := NewMigrator(log, config.SqlConnStr, config.SqlFileDir)
	s := MigrationService{
		log: log,
		mA:  action,
	}

	if config.MigrateOnStart {
		migrationStatus, err := action.Migrate()

		if err != nil {
			panic(err)
		}

		log.Infof("Tried to migrate on boot, status now: %v\n", migrationStatus)
	}

	return s
}

func (s MigrationService) DatabaseMigrate(ctx context.Context, request *DatabaseMigrateRequest) (*DatabaseMigrateResponse, error) {
	migratiosnStatus, err := s.mA.Migrate()

	if err != nil {
		log.Warningf(ctx, "could not migrate database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not migrate database")
	}

	return &DatabaseMigrateResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseStatus(ctx context.Context, request *DatabaseStatusRequest) (*DatabaseStatusResponse, error) {
	migratiosnStatus, err := s.mA.Status()

	if err != nil {
		log.Warningf(ctx, "could not get database status, err: %v", err)
		return nil, status.Error(codes.Internal, "could not get database status")
	}

	return &DatabaseStatusResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseRollback(ctx context.Context, request *DatabaseRollbackRequest) (*DatabaseRollbackResponse, error) {
	migratiosnStatus, err := s.mA.Rollback()

	if err != nil {
		log.Warningf(ctx, "could not rollback database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not rollback database")
	}

	return &DatabaseRollbackResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseForceVersion(ctx context.Context, request *DatabaseForceVersionRequest) (*DatabaseForceVersionResponse, error) {
	migratiosnStatus, err := s.mA.ForceVersion(request.GetVersion())

	if err != nil {
		log.Warningf(ctx, "could not force database version, err: %v", err)
		return nil, status.Error(codes.Internal, "could not force database version")
	}

	return &DatabaseForceVersionResponse{Status: migratiosnStatus}, nil
}
