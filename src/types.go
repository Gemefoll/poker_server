package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Token = string

type Config struct {
	Port              int
	Start_balance     int
	Postgres_db_name  string
	Postgres_user     string
	Postgres_password string
	Postgres_host     string
	Postgres_port     int
}

type Server struct {
	secret   []byte
	dbpool   *pgxpool.Pool
	conf     *Config
	gamesPtr int
}

type User struct {
	Id      int
	Name    string
	Role    string
	Balance int
}

type Card struct {
	Suit int
	Rank int
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Game struct {
	UsersId []string
	Owner   string
	Table   []Card
	Pots    [][]string
	Bet     []int
	MaxBet  []int
	Deck    []Card
	Hand    map[string][]Card
	Iter    int
	Round   int
	Turn    int
	IsStart bool
}
