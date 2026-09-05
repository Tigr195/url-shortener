// @title           URL Shortener API
// @version         1.0
// @description     Simple URL shortener service
// @host            localhost:8080
// @BasePath        /
package main

import (
	_ "github.com/Tigr195/url-shortener/docs"
	"github.com/Tigr195/url-shortener/internal/cache"
	"github.com/Tigr195/url-shortener/internal/logger"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"

	urlpb "github.com/Tigr195/url-shortener/gen/url"
	grpchandler "github.com/Tigr195/url-shortener/internal/grpc"
	"github.com/Tigr195/url-shortener/internal/handler"
	"github.com/Tigr195/url-shortener/internal/repository"
	"github.com/Tigr195/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Logger
	log := logger.New()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	appPort := os.Getenv("APP_PORT")
	baseURL := os.Getenv("BASE_URL")
	grpcPort := os.Getenv("GRPC_PORT")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Database
	dsn := "host=" + dbHost +
		" port=" + dbPort +
		" dbname=" + dbName +
		" user=" + dbUser +
		" password=" + dbPassword +
		" sslmode=" + dbSSLMode

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("connected to database")

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
	defer redisClient.Close()
	log.Info("connected to redis")

	// Layers
	urlRepo := repository.NewURLRepository(db)
	urlCache := cache.NewURLCache(redisClient)
	urlService := service.NewURLService(urlRepo, urlCache, baseURL)
	urlHandler := handler.NewURLHandler(urlService, log)

	// gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Error("failed to listen grpc", "error", err)
			os.Exit(1)
		}

		grpcServer := grpc.NewServer()
		urlpb.RegisterURLShortenerServer(grpcServer, grpchandler.NewURLGRPCServer(urlService))

		log.Info("starting gRPC server", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", "error", err)
			os.Exit(1)
		}
	}()

	// Router
	r := chi.NewRouter()

	r.Use(handler.LoggerMiddleware(log))
	r.Use(middleware.Recoverer)

	r.Post("/api/shorten", urlHandler.Shorten)
	r.Get("/{code}", urlHandler.Redirect)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Handle("/*", http.FileServer(http.Dir("./frontend")))

	// Start
	log.Info("starting server", "port", appPort)
	if err := http.ListenAndServe(":"+appPort, r); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
