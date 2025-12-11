package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Postgres *sql.DB
var MongoDB *mongo.Database

func ConnectPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	name := os.Getenv("POSTGRES_DB")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	// Validate
	if host == "" || port == "" || user == "" || name == "" {
		return fmt.Errorf("missing PostgreSQL environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	if err = db.Ping(); err != nil {
		return err
	}

	Postgres = db
	log.Println("Postgres connected")
	return nil
}


func ConnectMongo() error {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	if err = client.Ping(ctx, nil); err != nil {
		return err
	}
	MongoDB = client.Database(dbName)
	log.Println("MongoDB connected")
	return nil
}
