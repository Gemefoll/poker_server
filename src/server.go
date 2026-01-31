package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const tables = `
CREATE TABLE IF NOT EXISTS players (
	ID SERIAL PRIMARY KEY,
	Name varchar(20) UNIQUE NOT NULL, 
	Role varchar(5) NOT NULL, 
	Pass varchar(64) NOT NULL, 
	Salt varchar(26) NOT NULL, 
	Balance int NOT NULL
);
`

func ReadCofig() *Config {
	var res Config
	if str, ok := os.LookupEnv("PORT"); ok {
		if port, err := strconv.Atoi(str); err == nil {
			res.Port = port
		} else {
			panic(err)
		}
	} else {
		res.Port = 8080
	}
	if str, ok := os.LookupEnv("START_BALANCE"); ok {
		if port, err := strconv.Atoi(str); err == nil {
			res.Start_balance = port
		} else {
			panic(err)
		}
	} else {
		res.Start_balance = 15000
	}
	if str, ok := os.LookupEnv("DB_HOST"); ok {
		res.Postgres_host = str
	} else {
		panic(fmt.Errorf("env DB_HOST not exist"))
	}
	if str, ok := os.LookupEnv("DB_PASSWORD"); ok {
		res.Postgres_password = str
	} else {
		panic(fmt.Errorf("env DB_PASSWORD not exist"))
	}
	if str, ok := os.LookupEnv("DB_USER"); ok {
		res.Postgres_user = str
	} else {
		panic(fmt.Errorf("env DB_USER not exist"))
	}
	if str, ok := os.LookupEnv("DB_NAME"); ok {
		res.Postgres_db_name = str
	} else {
		panic(fmt.Errorf("env DB_NAME not exist"))
	}
	if str, ok := os.LookupEnv("DB_PORT"); ok {
		if port, err := strconv.Atoi(str); err == nil {
			res.Postgres_port = port
		} else {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("env DB_PORT not exist"))
	}
	return &res
}

func setupServer() {
	srv.conf = ReadCofig()
	srv.secret = []byte(rand.Text())
	var dsn = "postgres://" + srv.conf.Postgres_user + ":" + srv.conf.Postgres_password + "@" + srv.conf.Postgres_host + ":" + strconv.Itoa(srv.conf.Postgres_port) + "/" + srv.conf.Postgres_db_name
	var err error
	srv.dbpool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := srv.dbpool.Exec(context.Background(), tables); err != nil {
		panic(err)
	}
	fmt.Println("Done!")
	games = make(map[int]*Game)
}

func StartServer() error {
	http.HandleFunc("/api/ping", Ping)
	http.HandleFunc("/api/user/me", UserHeaderMe)
	http.HandleFunc("/api/user/create", SignUp)
	http.HandleFunc("/api/user/auth", SignIn)
	http.HandleFunc("/api/game/create", CreateGame)
	http.HandleFunc("/api/game/join", JoinGame)
	fmt.Println(srv.conf.Port)
	return http.ListenAndServe(":"+strconv.Itoa(srv.conf.Port), nil)
}
