package main

import (
	"html/template"
	"net/http"
	"strconv"
)

type Status = struct {
	Card1 string
	Card2 string
	Table []string
	Token string
	Bal   string
}

func playersHandle(w http.ResponseWriter, r *http.Request) {
	if !isGameStart {
		http.Error(w, "Game is not start", http.StatusTooEarly)
		return
	}
	val := r.URL.Query()
	a, ok := val["id"]
	if !ok {
		http.Error(w, "Id not found", http.StatusBadRequest)
		return
	}
	token := a[0]
	// lk.RLock()
	hand, ok := mapa[token]
	// lk.RUnlock()
	if !ok {
		http.Error(w, "Invalid id", http.StatusForbidden)
		return
	}
	var cur Status
	c := Convert(hand)
	cur.Card1 = c[0]
	cur.Card2 = c[1]
	// lk.RLock()
	cur.Table = Convert(table)
	cur.Bal = strconv.Itoa(curBal[token])
	// lk.RUnlock()
	cur.Token = token
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(w, cur); err != nil {
		panic(err)
	}
}
