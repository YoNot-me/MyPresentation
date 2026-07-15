package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"presentator/internal/core/entity"
	"presentator/internal/core/server"
	adminRepository "presentator/internal/features/admin/repository"
	adminService "presentator/internal/features/admin/service"
	adminTransport "presentator/internal/features/admin/transport"
	authRepository "presentator/internal/features/auth/repository"
	authService "presentator/internal/features/auth/service"
	"presentator/internal/features/auth/token"
	authTransport "presentator/internal/features/auth/transport"
	fileservingRepository "presentator/internal/features/fileserving/repository"
	fileservingService "presentator/internal/features/fileserving/service"
	fileservingTransport "presentator/internal/features/fileserving/transport"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 2*60*time.Second)
	defer cancel()

	env, db, logger, rdb, err := initFeatures(ctx)
	if err != nil {
		log.Printf("Smth get wrong to init features: %v", err)
		return
	}

	fileServingTrans, adminTrans, authTrans, jwtService := initServices(db, env, rdb, logger)

	srv, err := server.ServerInit(env, db, logger, fileServingTrans, authTrans, adminTrans, jwtService)
	if err != nil {
		log.Printf("Server init error: %v", err)
		return
	}

	defer func(Logger *zap.Logger) {
		err = Logger.Sync()
		if err != nil {
			return
		}
	}(srv.Logger)
	defer srv.DB.Close()
	defer func(rdb *redis.Client) {
		closeErr := rdb.Close()
		if closeErr != nil {
			return
		}
	}(rdb)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		srv.Logger.Info("Server starting")

		if err = srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel()
			srv.Logger.Info("Closing server", zap.Error(err))

			quit <- os.Interrupt
		}
	}()

	<-quit
	defer srv.Logger.Info("Server stopped gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Http.Shutdown(shutdownCtx); err != nil {
		srv.Logger.Error("Failed to gracefully shutdown server: ", zap.Error(err))
	}

	srv.Logger.Info("Server closed")
}

func initServices(
	db *pgxpool.Pool,
	env *entity.Config,
	rdb *redis.Client,
	log *zap.Logger,
) (
	*fileservingTransport.FileServingTransport,
	*adminTransport.AdminTransport,
	*authTransport.AuthTransport,
	*JWT.ServingJWT) {

	jwtService := JWT.NewServingJWT(rdb, env, log)

	authRepo := authRepository.NewAuthRepo(rdb, db)
	authSrv := authService.NewAuthService(log, jwtService, authRepo)
	authTrans := authTransport.NewAuthTransport(jwtService, log, authSrv)

	adminRepo := adminRepository.NewAdminRepo(rdb, db)
	adminSrv := adminService.NewAdminService(log, adminRepo, jwtService)
	adminTrans := adminTransport.NewAdminTransport(log, adminSrv)

	fileservingRepo := fileservingRepository.NewServingRepo(db)
	fileServiceSrv := fileservingService.NewServingService(log, jwtService, fileservingRepo)
	fileServingTrans := fileservingTransport.NewServingTransport(fileServiceSrv, log)

	return fileServingTrans, adminTrans, authTrans, jwtService
}
