package main

import (
	"math/rand/v2"
	"slices"
	"sort"
)

var isGameStart bool
var curBal map[Token]int
var games map[int]*Game
var WhereIsUser map[ID]int

var srv Server

func NewGame(Owner ID) *Game {
	return &Game{
		UserIDs: make([]ID, 0),
		Table:   make([]Card, 0),
		Pots:    make([][]ID, 0),
		Bet:     make([]int, 0),
		Stack:   make([]int, 0),
		Hand:    make(map[ID][]Card),
		Owner:   Owner,
		Deck:    NewDeck(),
	}
}

func combination(a []Card) []int {
	b := make([]int, 0)
	for _, i := range a {
		b = append(b, i.Rank)
	}
	sort.Slice(b, func(i, j int) bool {
		return b[i] > b[j]
	})
	is_street, is_flush := true, true
	for i := range a {
		if i != 0 && a[i].Suit != a[i-1].Suit {
			is_flush = false
		}
	}
	for i := range b {
		if i != 0 && b[i-1]-1 != b[i] {
			is_street = false
		}
	}
	if b[1] == 3 && b[2] == 2 && b[3] == 1 && b[4] == 0 && b[0] == 12 {
		b[0] = 3
		is_street = true
	}
	if is_street && is_flush {
		return []int{8, b[0]}
	}
	if b[1] == b[2] && b[2] == b[3] && (b[0] == b[1] || b[3] == b[4]) {
		if b[0] == b[1] {
			return []int{7, b[0], b[4]}
		}
		return []int{7, b[4], b[0]}
	}
	if b[0] == b[1] && b[3] == b[4] && (b[2] == b[3] || b[1] == b[2]) {
		if b[2] == b[3] {
			return []int{6, b[2], b[0]}
		}
		return []int{6, b[2], b[3]}
	}
	if is_flush {
		return []int{5, b[0], b[1], b[2], b[3], b[4]}
	}
	if is_street {
		return []int{4, b[0]}
	}
	if b[0] == b[1] && b[1] == b[2] {
		return []int{3, b[0], b[3], b[4]}
	}
	if b[3] == b[1] && b[1] == b[2] {
		return []int{3, b[1], b[0], b[4]}
	}
	if b[3] == b[4] && b[3] == b[2] {
		return []int{3, b[2], b[0], b[1]}
	}
	if b[0] == b[1] && b[2] == b[3] {
		return []int{2, b[0], b[2], b[4]}
	}
	if b[0] == b[1] && b[3] == b[4] {
		return []int{2, b[0], b[3], b[2]}
	}
	if b[1] == b[2] && b[3] == b[4] {
		return []int{2, b[1], b[3], b[0]}
	}
	if b[0] == b[1] {
		return []int{1, b[0], b[2], b[3], b[4]}
	}
	if b[1] == b[2] {
		return []int{1, b[1], b[0], b[3], b[4]}
	}
	if b[2] == b[3] {
		return []int{1, b[2], b[0], b[1], b[4]}
	}
	if b[3] == b[4] {
		return []int{1, b[3], b[0], b[1], b[2]}
	}
	return []int{0, b[0], b[1], b[2], b[3], b[4]}
}

func cmp(a, b []Card) int {
	if len(a) == 0 {
		return -1
	}
	return slices.Compare(combination(a), combination(b))
}

func NewDeck() []Card {
	deck := make([]Card, 52)
	for i := range 4 {
		for j := range 13 {
			deck[i*13+j] = Card{i, j}
		}
	}
	rand.Shuffle(len(deck), func(a, b int) {
		deck[a], deck[b] = deck[b], deck[a]
	})
	return deck
}

func Convert(a [][]int) []string {
	mast := []string{"a", "b", "c", "d"}
	nom := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	ans := make([]string, 0)
	for _, i := range a {
		ans = append(ans, nom[i[1]]+mast[i[0]])
	}
	return ans
}

// func EndGame() {
// 	for len(table) < 5 {
// 		OpenCard()
// 	}
// 	lk.RLock()
// 	bests := make([][][]int, 0)
// 	for _, token := range tokens {
// 		cur := table
// 		cur = append(cur, mapa[token]...)
// 		best := make([][]int, 0)
// 		for i := range 7 {
// 			for j := range 7 {
// 				if i >= j {
// 					continue
// 				}
// 				vec := make([][]int, 0)
// 				for k := range 7 {
// 					if k != i && k != j {
// 						vec = append(vec, cur[k])
// 					}
// 				}
// 				if cmp(best, vec) == -1 {
// 					best = vec
// 				}
// 			}
// 		}
// 		bests = append(bests, best)
// 	}
// 	lk.RUnlock()
// 	winners := make([]int, 0)
// 	for i, x := range bests {
// 		if folds[tokens[i]] {
// 			continue
// 		}
// 		if len(winners) == 0 {
// 			winners = append(winners, i)
// 			continue
// 		}
// 		c := cmp(x, bests[winners[0]])
// 		switch c {
// 		case 0:
// 			winners = append(winners, i)
// 		case 1:
// 			winners = []int{i}
// 		}
// 	}
// 	lk.RLock()
// 	fmt.Println("Winners:")
// 	for _, i := range winners {
// 		fmt.Println(name[tokens[i]])
// 	}
// 	lk.RUnlock()
// }

func main() {
	setupServer()
	err := StartServer()
	if err != nil {
		panic(err)
	}
}
