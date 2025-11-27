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

var (
	Postgres *sql.DB
	MongoDB  *mongo.Database
)

// Koneksi PostgreSQL
func ConnectPostgres() {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	pass := os.Getenv("PG_PASS")
	name := os.Getenv("PG_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal konek PostgreSQL:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("PostgreSQL tidak merespon:", err)
	}

	Postgres = db
	log.Println("PostgreSQL terhubung")
}

// Koneksi MongoDB
func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Gagal konek MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB tidak merespon:", err)
	}

	MongoDB = client.Database(dbName)
	log.Println("MongoDB terhubung")
}

// Inisialisasi semua database
func InitDatabase() {
	ConnectPostgres()
	ConnectMongo()
}
