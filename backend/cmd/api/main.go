package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"flowpilot/backend/internal/ai"
	"flowpilot/backend/internal/config"
	"flowpilot/backend/internal/document"
	"flowpilot/backend/internal/httpapi"
	"flowpilot/backend/internal/rag"
	"flowpilot/backend/internal/storage/postgres"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("FlowPilot API stopped: %v", err)
	}
}

func run() error {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if cfg.OpenAIAPIKey == "" {
		return errors.New("OPENAI_API_KEY is required")
	}

	// Root context is cancelled when the process receives a shutdown signal.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Initialize PostgreSQL connection pool.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database connection pool: %w", err)
	}
	defer pool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()

	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	log.Println("Connected to PostgreSQL")

	// Initialize OpenAI embedding provider.
	embeddingProvider, err := ai.NewOpenAIEmbeddingProvider(
		cfg.OpenAIAPIKey,
		ai.WithEmbeddingDimensions(cfg.EmbeddingDimensions),
	)
	if err != nil {
		return fmt.Errorf("initialize embedding provider: %w", err)
	}

	// Initialize OpenAI chat provider.
	chatProvider, err := ai.NewOpenAIChatProvider(
		cfg.OpenAIAPIKey,
	)
	if err != nil {
		return fmt.Errorf("initialize chat provider: %w", err)
	}

	// Initialize PostgreSQL vector store.
	vectorStore, err := postgres.NewVectorStore(
		pool,
		cfg.EmbeddingDimensions,
	)
	if err != nil {
		return fmt.Errorf("initialize vector store: %w", err)
	}

	// Initialize document extraction and chunking.
	documentService := document.NewService(document.NewParser())
	chunker, err := rag.NewTextChunker(1000, 200)
	if err != nil {
		return fmt.Errorf("create text chunker: %w", err)
	}

	// Initialize retrieval and grounded answer generation.
	retriever, err := rag.NewRAGRetriever(
		embeddingProvider,
		vectorStore,
	)
	if err != nil {
		return fmt.Errorf("initialize RAG retriever: %w", err)
	}

	answerer, err := rag.NewRAGAnswerer(
		retriever,
		chatProvider,
	)
	if err != nil {
		return fmt.Errorf("initialize RAG answerer: %w", err)
	}

	// Assemble the complete ingestion and question-answering pipeline.
	pipeline, err := rag.NewPipeline(
		documentService,
		chunker,
		embeddingProvider,
		vectorStore,
		answerer,
	)
	if err != nil {
		return fmt.Errorf("initialize RAG pipeline: %w", err)
	}

	// Existing summarization service remains unconfigured.
	// RAG ingestion and question-answering are available.
	router := httpapi.NewRouterWithRAGPipeline(
		documentService,
		nil,
		pipeline,
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("FlowPilot API listening on %s", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server: %w", err)
		}

	case <-ctx.Done():
		log.Println("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("gracefully shut down HTTP server: %w", err)
		}

		log.Println("HTTP server stopped")
	}

	return nil
}
