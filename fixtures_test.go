package cursor

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
)

var (
	db runner.Connection
	tx *runner.Tx
)

func init() {
	err := SetupFixtures()
	if err != nil {
		panic(err)
	}
}

// Conn returns the connection to exec queries on. If this is called within
// the context of a transaction you will get a transaction connection.
func Conn() runner.Connection {
	if tx != nil {
		return tx
	}

	return db
}

// Connect establishes a connection with the configured DATABASE_URL. You can
// keep a copy of the returned connection or fetch it using database.Conn()
func Connect() (runner.Connection, error) {
	dsn := os.Getenv("DATABASE_URL")
	u, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}

	if u.Path == "" {
		panic("DATABASE_URL does not set the database name")
	}

	driver, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// ensures the database can be pinged with an exponential backoff (15 min)
	runner.MustPing(driver)

	// set this to enable interpolation
	dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	db = runner.NewDB(driver, "postgres")
	return db, nil
}

// Reset will connect to the Postgres instance using the provided DATABASE_URL
// and drop and recreate the database.
func Reset() error {
	dsn := os.Getenv("DATABASE_URL")
	u, err := url.Parse(dsn)
	if err != nil {
		return err
	}

	if u.Path == "" {
		panic("DATABASE_URL does not set the database name")
	}

	// Keep a copy of the database name so we can drop/create it
	name := strings.Replace(u.Path, "/", "", -1)

	// Need to unset the path
	u.Path = "/"

	db, err := sql.Open("postgres", u.String())
	if err != nil {
		return err
	}

	dropCmd := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name)
	_, err = db.Exec(dropCmd)
	if err != nil {
		return err
	}

	createCmd := fmt.Sprintf("CREATE DATABASE %s;", name)
	_, err = db.Exec(createCmd)
	if err != nil {
		panic(err)
	}

	return nil
}

func SetupFixtures() error {
	err := Reset()
	if err != nil {
		return err
	}

	db, err = Connect()
	if err != nil {
		return err
	}

	err = loadFixtures()
	if err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	done := make(chan int, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for {
			switch <-signalChan {
			case syscall.SIGHUP:
				fmt.Println("Received SIGHUP")
				done <- 0
			case syscall.SIGINT:
				fmt.Println("Received SIGINT")
				done <- 1
			case syscall.SIGTERM:
				fmt.Println("Received SIGTERM")
				done <- 0
			case syscall.SIGQUIT:
				fmt.Println("Received SIGQUIT")
				done <- 0
			default:
				fmt.Println("Received unknown signal")
				done <- 1
			}
		}
	}()

	go func() {
		code := <-done
		os.Exit(code)
	}()

	return nil
}

func loadFixtures() error {
	seedPath := filepath.Join("testdata/seed.sql")
	seedData, err := ioutil.ReadFile(seedPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(seedData))
	if err != nil {
		return err
	}

	return nil
}
