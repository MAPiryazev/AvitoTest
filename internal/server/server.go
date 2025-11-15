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
func NewServer() (*Server, error) {
	// Загружаем конфигурацию БД
	dbCfg, err := config.LoadDBConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load DB config: %w", err)
	}

	// Инициализируем подключение к БД
	db, err := config.InitDB(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init DB: %w", err)
	}

	// Создаем репозитории
	userRepo := postgres.NewUserRepositoryPG(db)
	teamRepo := postgres.NewTeamRepositoryPG(db)
	prRepo := postgres.NewPullRequestRepositoryPG(db)

	repo := &repository.Repository{
		UserRepo: userRepo,
		TeamRepo: teamRepo,
		PRRepo:   prRepo,
	}

	// Создаем сервисы
	svc := service.NewService(repo)

	// Создаем хендлеры
	h := handler.NewHandler(svc)

	// Загружаем конфигурацию API
	apiCfg, err := config.LoadAPIConfig()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to load API config: %w", err)
	}

	// Создаем HTTP роутер
	router := api.Handler(h)

	// Создаем HTTP сервер
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

// Start запускает HTTP сервер
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown корректно останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	// Закрываем HTTP сервер
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	// Закрываем подключение к БД
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close DB connection: %w", err)
	}

	log.Println("Server shutdown complete")
	return nil
}
