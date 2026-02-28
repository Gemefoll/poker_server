package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func GetToken(Id ID) TokenPair {
	AccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Id,
		"Access",
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	})
	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Id,
		"Refresh",
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	})
	StrAccessToken, _ := AccessToken.SignedString(srv.secret)
	StrRefreshToken, _ := RefreshToken.SignedString(srv.secret)
	return TokenPair{
		AccessToken:  StrAccessToken,
		RefreshToken: StrRefreshToken,
	}
}

func SignUp(c fiber.Ctx) error {
	type Data struct {
		Name string
		Pass string
	}
	res := new(Data)
	if err := c.Bind().JSON(res); err != nil {
		return fiber.ErrBadRequest
	}
	if len(res.Name) > 20 || len(res.Pass) > 20 {
		return fiber.ErrBadRequest
	}
	if len(res.Pass) < 6 || len(res.Name) == 0 {
		return fiber.ErrBadRequest
	}
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Name FROM players WHERE Name = $1", res.Name)
	var username string
	if err := a.Scan(&username); err == nil {
		return fiber.ErrConflict
	}
	Salt := rand.Text()
	b := srv.dbpool.QueryRow(context.Background(), "INSERT INTO players (Name, Pass, Role, Salt, Balance) VALUES ($1, $2, $3, $4, $5) RETURNING id", res.Name, fmt.Sprintf("%x", sha256.Sum256([]byte(res.Pass+Salt))), "User", Salt, srv.conf.Start_balance)
	var Id ID
	if err := b.Scan(&Id); err != nil {
		panic(err)
	}
	return c.JSON(GetToken(Id))
}

func SignIn(c fiber.Ctx) error {
	type Data struct {
		Name string
		Pass string
	}
	res := new(Data)
	if err := c.Bind().JSON(res); err != nil {
		return fiber.ErrBadRequest
	}
	if len(res.Name) > 20 {
		return fiber.ErrBadRequest
	}
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Id, Pass, Salt FROM players WHERE name = $1", res.Name)
	var Id ID
	var sum, salt string
	err := a.Scan(&Id, &sum, &salt)
	if err != nil {
		return fiber.ErrConflict
	}
	if sum != fmt.Sprintf("%x", sha256.Sum256([]byte(res.Pass+salt))) {
		return fiber.ErrForbidden
	}
	return c.JSON(GetToken(Id))
}

func CheckAuth(r *http.Request, ReqTp string) (ID, error) {
	var tokenString string
	if res, ok := r.Header["Authorization"]; ok {
		if _, err := fmt.Sscanf(res[0], "Bearer %s", &tokenString); err != nil {
			return 0, errors.New("UAU")
		} else {
			token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
				return srv.secret, nil
			}, jwt.WithLeeway(5*time.Second))
			if err != nil {
				return 0, errors.New("UAU")
			} else if claims, ok := token.Claims.(*TokenClaims); ok && claims.Type == ReqTp {
				return claims.ID, nil
			} else {
				return 0, errors.New("UAU")
			}
		}
	} else {
		return 0, errors.New("UAU")
	}
}

func Refresh(c fiber.Ctx) error {
	return c.JSON(GetToken(fiber.Locals[ID](c, "Id")))
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	Id, err := CheckAuth(r, "Access")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, ok := WhereIsUser[Id]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}
	srv.gamesPtr++
	res := srv.gamesPtr
	WhereIsUser[Id] = res - 1
	games[res-1] = NewGame(Id)
	enc := json.NewEncoder(w)
	type Response struct {
		Id ID
	}
	enc.Encode(Response{
		Id: res - 1,
	})
	// TODO
	// JoinGame(w, r)
}

func GetUserMe(c fiber.Ctx) error {
	Id := fiber.Locals[ID](c, "Id")
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Name, Role, Balance FROM players WHERE Id = $1", Id)
	var Balance int
	var Username, Role string
	if err := a.Scan(&Username, &Role, &Balance); err != nil {
		panic(err)
	}
	type Data struct {
		Username string
		Balance  int
	}
	return c.JSON(User{
		Name:    Username,
		Balance: Balance,
		Role:    Role,
		Id:      Id,
	})
}

func DeleteUserMe(c fiber.Ctx) {
	Id := fiber.Locals[ID](c, "Id")
	if _, err := srv.dbpool.Exec(context.Background(), "DELETE FROM players WHERE Id = $1", Id); err != nil {
		panic(err)
	}
}

// TODO
func JoinGame(w http.ResponseWriter, r *http.Request) {
	Id, err := CheckAuth(r, "Access")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, ok := WhereIsUser[Id]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}
	type Req struct {
		Id    ID
		Stack int
	}
	var res Req
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&res); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var gid ID
	if res.Id == -1 {
		srv.gamesPtr++
		res := srv.gamesPtr
		WhereIsUser[Id] = res - 1
		games[res-1] = NewGame(Id)
		gid = res - 1
	} else {
		gm, ok := games[res.Id]
		if !ok || gm.IsStart {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gid = res.Id
	}
	gm := games[gid]
	gm.UserIDs = append(gm.UserIDs, Id)
	gm.Bet = append(gm.Bet, 0)
	gm.Delta = append(gm.Delta, 0)
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Balance FROM players WHERE Id = $1", Id)
	var bal int
	if err := a.Scan(&bal); err != nil || bal < res.Stack {
		w.WriteHeader(http.StatusConflict)
		return
	}
	gm.Stack = append(gm.Stack, res.Stack)
	enc := json.NewEncoder(w)
	type Resp struct {
		Id ID
	}
	enc.Encode(Resp{gid})
}

func Ping(c fiber.Ctx) error {
	return c.SendString("ok")
}
