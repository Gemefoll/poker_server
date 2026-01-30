package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupServer() {
	fl, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	res, err := io.ReadAll(fl)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(res, &srv.conf)
	if err != nil {
		panic(err)
	}
	srv.secret = []byte(rand.Text())
	var dsn = "postgres://" + srv.conf.Postgres_user + ":@"+ srv.conf.Postgres_host + ":" + strconv.Itoa(srv.conf.Postgres_port) + "/" + srv.conf.Postgres_db_name
	srv.dbpool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := srv.dbpool.Ping(context.Background()); err != nil {
		panic(err)
	}
	games = make(map[int]*Game)
}

func StartServer() error {
	http.HandleFunc("/", homeHendle)
	http.HandleFunc("/api/ping", Ping)
	http.HandleFunc("/api/cards", Cards)
	http.HandleFunc("/api/cnt", LenTable)
	http.HandleFunc("/api/table", GetTable)
	http.HandleFunc("/api/fold", Fold)
	http.HandleFunc("/game", playersHandle)
	http.HandleFunc("/api/signup", SignUp)
	http.HandleFunc("/api/signin", SignIn)
	http.HandleFunc("/api/game/create", CreateGame)
	http.HandleFunc("/api/game/join", JoinGame)
	http.Handle("/card_img/", http.StripPrefix("/card_img/", http.FileServer(http.Dir("./card_img"))))
	return http.ListenAndServe(":"+strconv.Itoa(srv.conf.Port), nil)
}
