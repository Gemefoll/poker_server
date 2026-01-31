package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// "github.com/georgysavva/scany/v2/pgxscan"
	"github.com/golang-jwt/jwt/v5"
)

func GetToken(Name string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Name":         Name,
		"WillExpireAt": time.Now().Add(time.Hour * 24).Format(time.UnixDate),
	})
	var str string
	str, err := token.SignedString(srv.secret)
	if err != nil {
		panic(err)
	}
	return str
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Name string
		Pass string
	}
	var res Data
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(res.Name) > 20 || len(res.Pass) > 20 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(res.Pass) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Name FROM players WHERE Name = $1", res.Name)
	var username string
	if err := a.Scan(&username); err == nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	Salt := rand.Text()
	_, err := srv.dbpool.Exec(context.Background(), "INSERT INTO players (Name, Pass, Role, Salt, Balance) VALUES ($1, $2, $3, $4, $5)", res.Name, fmt.Sprintf("%x", sha256.Sum256([]byte(res.Pass+Salt))), "User", Salt, srv.conf.Start_balance)
	if err != nil {
		fmt.Println(err)
	}
	type Resp struct {
		AccessToken  string
		RefreshToken string
	}
	enc := json.NewEncoder(w)
	enc.Encode(Resp{
		AccessToken:  GetToken(res.Name),
		RefreshToken: "",
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Name string
		Pass string
	}
	var res Data
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(res.Name) > 20 {
		w.WriteHeader(http.StatusRequestHeaderFieldsTooLarge)
		return
	}
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Pass, Salt FROM players WHERE name = $1", res.Name)
	var sum, salt string
	err := a.Scan(&sum, &salt)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if sum != fmt.Sprintf("%x", sha256.Sum256([]byte(res.Pass+salt))) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	enc := json.NewEncoder(w)
	type Resp struct {
		AccessToken  string
		RefreshToken string
	}
	enc.Encode(Resp{
		AccessToken:  GetToken(res.Name),
		RefreshToken: "",
	})
}

func CheckAuth(r *http.Request) (string, error) {
	var tokenString string
	var Name string
	if res, ok := r.Header["Authorization"]; ok {
		if _, err := fmt.Sscanf(res[0], "Bearer %s", &tokenString); err != nil {
			return "", errors.New("UAU")
		} else {
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				return srv.secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil {
				return "", errors.New("UAU")
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				Name = claims["Name"].(string)
				WillExpireAtStr, _ := (claims["WillExpireAt"].(string))
				WillExpireAt, _ := time.Parse(time.UnixDate, WillExpireAtStr)
				if time.Now().After(WillExpireAt) {
					return "", errors.New("UAU")
				}
			} else {
				return "", errors.New("UAU")
			}
		}
	} else {
		return "", errors.New("UAU")
	}
	return Name, nil
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	username, err := CheckAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, ok := WhereIsUser[username]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}
	srv.gamesPtr++
	res := srv.gamesPtr
	fmt.Fprint(w, res-1)
	games[res-1] = NewGame(username)
	WhereIsUser[username] = res - 1
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	username, err := CheckAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Balance FROM players WHERE name = $1", username)
	var bal int
	if err := a.Scan(&bal); err != nil {
		panic(err)
	}
	type Data struct {
		Username string
		Balance  int
	}
	enc := json.NewEncoder(w)
	dat := Data{
		Username: username,
		Balance:  bal,
	}
	if err := enc.Encode(dat); err != nil {
		panic(err)
	}
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	username, err := CheckAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, err := srv.dbpool.Exec(context.Background(), "DELETE FROM players WHERE name = $1", username); err != nil {
		panic(err)
	}
}

func UserHeader(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetUser(w, r)
	case http.MethodDelete:
		DeleteUser(w, r)
	}
}

func JoinGame(w http.ResponseWriter, r *http.Request) {
	username, err := CheckAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, ok := WhereIsUser[username]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if v, ok := r.URL.Query()["id"]; !ok || len(v) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := r.URL.Query()["id"][0]
	ind, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	gm, ok := games[ind]
	if !ok || gm.IsStart {
		w.WriteHeader(http.StatusBadRequest)
	}
	gm.UsersId = append(gm.UsersId, username)
	gm.Bet = append(gm.Bet, 0)
	a := srv.dbpool.QueryRow(context.Background(), "SELECT Balance FROM players WHERE name = $1", username)
	var bal int
	if err := a.Scan(&bal); err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	gm.MaxBet = append(gm.MaxBet, bal)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}