package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
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
	router := http.NewServeMux()

	router.HandleFunc("/", routes.NewRouter(db).TodosRouter)

	SetupGraceFullShutdown(router)
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
		log.Fatalf("Database connection error: %s", err.Error())
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

	defer log.Println("Database connection success")
	return sqlx.NewDb(db, driver_name)
}

func SetupGraceFullShutdown(handler *http.ServeMux) {
	httpServer := http.Server{
		Addr:           fmt.Sprintf(":%s", viper.GetString("PORT")),
		ReadTimeout:    time.Duration(time.Second) * 60,
		WriteTimeout:   time.Duration(time.Second) * 30,
		IdleTimeout:    time.Duration(time.Second) * 120,
		MaxHeaderBytes: 3145728,
		Handler:        handler,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Server is not running: %s", err.Error())
		os.Exit(0)
	}

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	log.Printf("Signal received: %v\n", <-osSignal)

	if sig := <-osSignal; sig == os.Interrupt || sig == syscall.SIGTERM {
		log.Println("Waiting to server shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second)*10)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP server shutdown error: %s", err.Error())
		}

		log.Println("HTTP server shutdown")
	}
}
