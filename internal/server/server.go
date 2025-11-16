package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"AvitoTest/internal/api"
	"AvitoTest/internal/config"
	"AvitoTest/internal/handler"
	"AvitoTest/internal/repository"
	"AvitoTest/internal/repository/postgres"
	"AvitoTest/internal/service"
)

type Server struct {
	httpServer *http.Server
	db         interface {
		Close() error
	}
}

// NewServer создает и инициализирует сервер со всеми зависимостями
func NewServer(envPath string) (*Server, error) {
	dbCfg, err := config.LoadDBConfig(envPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфиг БД: %w", err)
	}

	db, err := config.InitDB(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации БД: %w", err)
	}

	userRepo := postgres.NewUserRepositoryPG(db)
	teamRepo := postgres.NewTeamRepositoryPG(db)
	prRepo := postgres.NewPullRequestRepositoryPG(db)

	repo := &repository.Repository{
		UserRepo: userRepo,
		TeamRepo: teamRepo,
		PRRepo:   prRepo,
	}

	svc := service.NewService(repo)

	h := handler.NewHandler(svc)

	apiCfg, err := config.LoadAPIConfig(envPath)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ошибка загрузки api конфига: %w", err)
	}

	router := api.Handler(h)

	httpServer := &http.Server{
		Addr:         ":" + apiCfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		db:         db,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("запуск сервера на  %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("остановка сервера...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("не удалось остановить сервер: %w", err)
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть подключение к БД: %w", err)
	}

	log.Println("Остановка сервиса успешна")
	return nil
}
