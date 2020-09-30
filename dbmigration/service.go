package dbmigration

import (
	"context"
	devv1 "dev.azure.com/securitasintelligentservices/insights/_git/sispbgo.git/sis/rp/dev/v1"
	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/appengine/log"
	"google.golang.org/grpc/codes"
)

type MigrationService struct {
	log *zap.SugaredLogger

	mA Migrator
}

func NewMigrationService(log *zap.SugaredLogger, config MigrationConfig) devv1.MigrationServiceServer {
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

func (s MigrationService) DatabaseMigrate(ctx context.Context, request *devv1.DatabaseMigrateRequest) (*devv1.DatabaseMigrateResponse, error) {
	migratiosnStatus, err := s.mA.Migrate()

	if err != nil {
		log.Warningf(ctx, "could not migrate database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not migrate database")
	}

	return &devv1.DatabaseMigrateResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseStatus(ctx context.Context, request *devv1.DatabaseStatusRequest) (*devv1.DatabaseStatusResponse, error) {
	migratiosnStatus, err := s.mA.Status()

	if err != nil {
		log.Warningf(ctx, "could not get database status, err: %v", err)
		return nil, status.Error(codes.Internal, "could not get database status")
	}

	return &devv1.DatabaseStatusResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseRollback(ctx context.Context, request *devv1.DatabaseRollbackRequest) (*devv1.DatabaseRollbackResponse, error) {
	migratiosnStatus, err := s.mA.Rollback()

	if err != nil {
		log.Warningf(ctx, "could not rollback database, err: %v", err)
		return nil, status.Error(codes.Internal, "could not rollback database")
	}

	return &devv1.DatabaseRollbackResponse{Status: migratiosnStatus}, nil
}

func (s MigrationService) DatabaseForceVersion(ctx context.Context, request *devv1.DatabaseForceVersionRequest) (*devv1.DatabaseForceVersionResponse, error) {
	migratiosnStatus, err := s.mA.ForceVersion(request.GetVersion())

	if err != nil {
		log.Warningf(ctx, "could not force database version, err: %v", err)
		return nil, status.Error(codes.Internal, "could not force database version")
	}

	return &devv1.DatabaseForceVersionResponse{Status: migratiosnStatus}, nil
}
