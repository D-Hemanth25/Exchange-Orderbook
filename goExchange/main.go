package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"log"

	"github.com/gorilla/mux"
)

type Balances map[string]int

type User struct {
	Id      string
	Balance Balances
}

type Order struct {
	UserId   string
	Price    float64
	Quantity int64
}

const TICKER = "PLTR"

var users = []User{
	{
		Id: "1",
		Balance: Balances{
			"PLTR": 10,
			"USD":  50000,
		},
	},
	{
		Id: "2",
		Balance: Balances{
			"PLTR": 20,
			"USD":  70000,
		},
	},
}

// list of open orders
var bids []Order
var asks []Order

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/order/{userId}/{side}/{quantity}/{price}", handleOrder).Methods("POST")
	router.HandleFunc("/depth", handleDepth).Methods("GET")
	router.HandleFunc("/balance/{userId}", handleBalance).Methods("GET")

	fmt.Println(users[1])
	http.ListenAndServe(":9090", router)
}

func handleOrder(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userId = vars["userId"]
	var side = vars["side"]

	quantity, err := strconv.ParseInt(vars["quantity"], 10, 32)
	if err != nil {
		http.Error(rw, "invalid quantity", http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(vars["price"], 64)
	if err != nil {
		http.Error(rw, "Invalid price", http.StatusBadRequest)
		return
	}

	var remainingQuantity = fillOrders(side, price, int64(quantity), userId)

	if remainingQuantity == 0 {
		fmt.Fprintf(rw, "filled quantity: %d", quantity)
		log.Println("quantity", quantity)
		return
	}

	order := Order{
		UserId:   userId,
		Price:    price,
		Quantity: remainingQuantity,
	}

	if side == "bid" {
		bids = append(bids, order)
		sort.Slice(bids, func(i, j int) bool {
			return bids[i].Price < bids[j].Price
		})
	} else {
		asks = append(asks, order)
		sort.Slice(asks, func(i, j int) bool {
			return asks[i].Price > asks[j].Price
		})
	}
	fmt.Fprintf(rw, "filled quantity: %d", quantity-remainingQuantity)
	log.Println("remaining: ", quantity)
}

func handleDepth(rw http.ResponseWriter, r *http.Request) {
	depth := struct {
		Bids []Order
		Asks []Order
	}{
		Bids: bids,
		Asks: asks,
	}
	fmt.Fprintf(rw, "current state of the order book: %v", depth)
	log.Println("depth: ", depth)
}

func handleBalance(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userId = vars["userId"]
	user, found := findUser(userId)

	if found {
		pltrQty := user.Balance["PLTR"]
		usdtQty := user.Balance["USD"]
		fmt.Fprintf(rw, "User ID: %s\nPLTR Quantity: %d\nUSD Quantity: %d", userId, pltrQty, usdtQty)

		log.Printf("User ID: %s\nPLTR Quantity: %d\nUSD Quantity: %d", userId, pltrQty, usdtQty)
	} else {
		http.Error(rw, "User not found", http.StatusNotFound)
	}
}

func findUser(userId string) (*User, bool) {
	for _, user := range users {
		if user.Id == userId {
			return &user, true
		}
	}
	return nil, false
}

func fillOrders(side string, price float64, quantity int64, userId string) int64 {
	var remainingQuantity = quantity
	if side == "bid" {
		for index := len(asks) - 1; index >= 0; index++ {
			if asks[index].Price > price {
				continue
			}
			if int64(asks[index].Quantity) > remainingQuantity {
				asks[index].Quantity -= int64(remainingQuantity)
				flipBalance(asks[index].UserId, userId, remainingQuantity, price)
				return 0
			} else {
				remainingQuantity -= asks[index].Quantity
				flipBalance(asks[index].UserId, userId, asks[index].Quantity, price)
				asks = asks[:len(asks)-1]
			}
		}
	} else {
		for index := len(bids) - 1; index >= 0; index++ {
			if bids[index].Price < price {
				continue
			}
			if bids[index].Quantity > remainingQuantity {
				bids[index].Quantity -= remainingQuantity
				flipBalance(userId, bids[index].UserId, remainingQuantity, price)
				return 0
			} else {
				remainingQuantity -= bids[index].Quantity
				flipBalance(userId, bids[index].UserId, bids[index].Quantity, price)
				bids = bids[:len(bids)-1]
			}
		}
	}
	return remainingQuantity
}

func flipBalance(userId1 string, userId2 string, quantity int64, price float64) {
	user1, _ := findUser(userId1)
	user2, _ := findUser(userId2)

	user1.Balance[TICKER] -= int(quantity)
	user2.Balance[TICKER] += int(quantity)

	user1.Balance["USD"] += int(price * float64(quantity))
	user2.Balance["USD"] -= int(price * float64(quantity))
}
