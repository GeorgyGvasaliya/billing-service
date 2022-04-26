package handler

import (
	"billing-service/db"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type handler struct {
	cache map[int]int
	cfg   db.Config
}

type MoneyResponse struct {
	Id    int `json:"user"`
	Money int `json:"money"`
}

type SendResponse struct {
	SenderID   int `json:"sender_id"`
	ReceiverID int `json:"receiver_id"`
	Money      int `json:"money"`
}

type Handler interface {
	Register(router *httprouter.Router)
}

func NewHandler(cache map[int]int, cfg db.Config) Handler {
	return &handler{
		cache: cache,
		cfg:   cfg,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET("/get", h.GetUser)
	router.POST("/add", h.AddMoney)
	router.POST("/withdraw", h.Withdraw)
	router.POST("/send", h.SendMoney)
}

func (h *handler) GetUser(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	data := db.NewPostgresDB(h.cfg)
	userID := r.URL.Query().Get("id")

	var balance int
	queryGet := "select balance from public.wallet where user_id=$1"
	err := data.QueryRow(queryGet, userID).Scan(&balance)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("There is no user with such ID or wrong query"))
		return
	}

	balanceS := strconv.Itoa(balance)
	w.Write([]byte(balanceS))

	err = data.Close()
	if err != nil {
		log.Println("Cannot close connection")
	}
}

func (h *handler) AddMoney(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	w.Header().Add("Content-Type", "application/json")
	data := db.NewPostgresDB(h.cfg)
	body, _ := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	var mr MoneyResponse
	err := json.Unmarshal(body, &mr)
	if err != nil {
		log.Println("Cannot Unmarshal data")
		return
	}

	// store in cache balance
	if _, ok := h.cache[mr.Id]; !ok {
		h.cache[mr.Id] = mr.Money
	} else {
		h.cache[mr.Id] += mr.Money
	}

	// insert if not exists, update if exists
	queryUpsert := "INSERT INTO public.wallet (wallet_id, user_id, balance) VALUES ($1, $1, $2) ON CONFLICT (wallet_id) DO UPDATE SET balance=wallet.balance+$2;"
	data.QueryRow(queryUpsert, mr.Id, mr.Money)

	err = data.Close()
	if err != nil {
		log.Println("Cannot close connection")
	}
}

func (h *handler) Withdraw(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	w.Header().Add("Content-Type", "application/json")
	data := db.NewPostgresDB(h.cfg)
	body, _ := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	var mr MoneyResponse
	err := json.Unmarshal(body, &mr)
	if err != nil {
		log.Println("Cannot Unmarshal data")
		return
	}

	// вообще этот кеш нужен был, чтобы при выводе средств проверять на отрицательный баланс, одним sql запросом не обойдёшься
	balance := h.cache[mr.Id]
	newBalance := balance - mr.Money
	if newBalance < 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("You have no money to purchase"))
		return
	}

	h.cache[mr.Id] = newBalance
	queryUpsert := "INSERT INTO public.wallet (wallet_id, user_id, balance) VALUES ($1, $1, $2) ON CONFLICT (wallet_id) DO UPDATE SET balance=wallet.balance-$2;"
	data.QueryRow(queryUpsert, mr.Id, mr.Money)

	err = data.Close()
	if err != nil {
		log.Println("Cannot close connection")
	}
}

func (h *handler) SendMoney(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	data := db.NewPostgresDB(h.cfg)
	body, _ := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	var sr SendResponse
	err := json.Unmarshal(body, &sr)
	if err != nil {
		log.Println("Cannot Unmarshal data")
		return
	}

	balance := h.cache[sr.SenderID]
	fmt.Println(h.cache)
	fmt.Println(sr.SenderID)
	fmt.Println(balance, sr.Money)
	newBalanceSender := balance - sr.Money
	if newBalanceSender < 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("You have no money to send"))
		return
	}

	h.cache[sr.SenderID] = newBalanceSender
	h.cache[sr.ReceiverID] += sr.Money

	querySender := "update public.wallet set balance=balance-$1 where user_id=$2"
	data.QueryRow(querySender, sr.Money, sr.SenderID)

	queryReceiver := "update public.wallet set balance=balance+$1 where user_id=$2"
	data.QueryRow(queryReceiver, sr.Money, sr.ReceiverID)

	err = data.Close()
	if err != nil {
		log.Println("Cannot close connection")
	}
}

//Метод начисления средств на баланс. Принимает id пользователя и сколько средств зачислить.
//
//Метод списания средств с баланса. Принимает id пользователя и сколько средств списать.
//
//Метод перевода средств от пользователя к пользователю. Принимает id пользователя с которого нужно списать средства, id пользователя которому должны зачислить средства, а также сумму.
//
//Метод получения текущего баланса пользователя. Принимает id пользователя. Баланс всегда в рублях.
