package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"

	"github.com/restuwahyu13/todos/routes"
)

func main() {
	if _, ok := os.LookupEnv("GO_ENV"); !ok {
		if err := SetupEnv(); err != nil {
			log.Fatalf("Env file not found: %s", err.Error())
		}
	}

	db := SetupDatabase()
	router := httprouter.New()

	router.POST("/", routes.NewRouter(db).CreateTodos)
	router.GET("/", routes.NewRouter(db).GetlAllTodos)
	router.GET("/:id", routes.NewRouter(db).GetTodosById)
	router.DELETE("/:id", routes.NewRouter(db).DeleteTodosById)
	router.POST("/:id", routes.NewRouter(db).UpdateTodosById)

	SetupGraceFullShutdown(router, db, viper.GetString("PORT"))
}

func SetupEnv() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	return err
}

func SetupDatabase() *sqlx.DB {
	driver_name := viper.GetString("DB_DRIVER")
	dsn_url := viper.GetString("PG_DSN")
	migration_dir := viper.GetString("DB_DIR_MIGRATION")

	db, _ := sql.Open(driver_name, dsn_url)

	if err := db.Ping(); err != nil {
		defer log.Fatalf("Database connection error: %s", err.Error())
		db.Close()
	} else {
		ctx, close := context.WithTimeout(context.TODO(), time.Duration(time.Second*30))
		defer close()

		log.Println("Database connection success")
		db.Conn(ctx)
	}

	if viper.GetString("GO_ENV") == "development" {
		migrations := migrate.FileMigrationSource{
			Dir: migration_dir,
		}

		if _, err := migrations.FindMigrations(); err != nil {
			log.Fatalf("Find migration database error: %s", err.Error())
		}

		if _, err := migrate.Exec(db, driver_name, migrations, migrate.Up); err != nil {
			log.Fatalf("Execute migration database error: %s", err.Error())
		}
	}

	return sqlx.NewDb(db, driver_name)
}

func SetupGraceFullShutdown(handler *httprouter.Router, db *sqlx.DB, port string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	httpServer := http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		ReadTimeout:    time.Duration(time.Second) * 60,
		WriteTimeout:   time.Duration(time.Second) * 30,
		IdleTimeout:    time.Duration(time.Second) * 15,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		Handler:        handler,
	}

	wg.Add(1)
	go func() {
		if err := httpServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server not running: %s", err.Error())
		}
		log.Printf("HTTP server running on port: %s", port)
		wg.Done()
	}()
	wg.Wait()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	if sig, ok := <-stop; ok {
		log.Printf("Signal received: %v \n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		defer cancel()

		log.Println("Waiting to HTTP server shutdown...")

		if err := db.Close(); err != nil {
			log.Printf("Database shutdown error: %s", err.Error())
		}

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %s", err.Error())
		}

		defer close(stop)
		log.Println("HTTP server shutdown success")

		os.Exit(0)
	}
}
