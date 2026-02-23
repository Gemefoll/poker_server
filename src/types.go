package main

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Token = string
type ID = int

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
	Id      ID
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

type TokenClaims struct {
	ID   int
	Type string
	jwt.RegisteredClaims
}

type Game struct {
	UserIDs []ID
	Owner   ID
	Table   []Card
	Pots    [][]ID
	Bet     []int
	Stack   []int
	Delta   []int
	Deck    []Card
	Hand    map[ID][]Card
	Iter    int
	Round   int
	Turn    int
	IsStart bool
}
