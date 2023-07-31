package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sqldata "l0/internal"
	"l0/model"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

var stmt *sql.Stmt

func prepare(db *sql.DB) {
	var err error
	for {
		stmt, err = db.Prepare("select orders.*, delivery.name, delivery.phone, delivery.zip, delivery.city, delivery.address, delivery.region, delivery.email, payment.transaction, payment.request_id, payment.currency, payment.provider, payment.amount, payment.payment_dt, payment.bank, payment.delivery_cost, payment.goods_total, payment.custom_fee, items.chrt_id, items.track_number, items.price, items.rid, items.name, items.sale, items.size, items.total_price, items.nm_id, items.brand, items.status from orders join delivery on orders.order_uid = delivery.order_uid join payment on orders.order_uid = payment.order_uid join items on orders.order_uid = items.order_uid")
		if err == nil {
			break
		}
		log.Println("Error preparing SQL statement:", err)
		time.Sleep(time.Second * 2) // Пауза перед повторной попыткой
	}
}

func loadOrders(db *sql.DB, orderCache *sync.Map) {

	var rows *sql.Rows
	var err error
	for {
		rows, err = stmt.Query()
		if err == nil {
			break
		}
		log.Println("Error executing query, retrying:", err)
		time.Sleep(time.Second * 2)
	}

	for rows.Next() {
		var orderData model.Order
		var item model.Item
		err := rows.Scan(
			&orderData.OrderUID,
			&orderData.TrackNumber,
			&orderData.Entry,
			&orderData.Locale,
			&orderData.InternalSignature,
			&orderData.CustomerID,
			&orderData.DeliveryService,
			&orderData.Shardkey,
			&orderData.SmID,
			&orderData.DateCreated,
			&orderData.OofShard,
			&orderData.Delivery.Name,
			&orderData.Delivery.Phone,
			&orderData.Delivery.Zip,
			&orderData.Delivery.City,
			&orderData.Delivery.Address,
			&orderData.Delivery.Region,
			&orderData.Delivery.Email,
			&orderData.Payment.Transaction,
			&orderData.Payment.RequestID,
			&orderData.Payment.Currency,
			&orderData.Payment.Provider,
			&orderData.Payment.Amount,
			&orderData.Payment.PaymentDt,
			&orderData.Payment.Bank,
			&orderData.Payment.DeliveryCost,
			&orderData.Payment.GoodsTotal,
			&orderData.Payment.CustomFee,
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)

		if err != nil {
			log.Println("Error executing query:", err)
			return
		}
		orderData.Items = append(orderData.Items, item)
		orderCache.Store(orderData.OrderUID, orderData)

	}

}

func handleMsg(msg *nats.Msg, db *sql.DB, orderCache *sync.Map) {
	var orderData model.Order
	json.Unmarshal(msg.Data, &orderData)

	sqldata.InsertData(db, orderData)
	orderCache.Store(orderData.OrderUID, orderData)
}

func handleSearch(w http.ResponseWriter, r *http.Request, orderCache *sync.Map) {
	params := mux.Vars(r)
	uid := params["uid"]

	result, found := orderCache.Load(uid)
	if !found {
		http.Error(w, "No data found for the given UID.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(result)
}

func main() {
	var err error
	var nc *nats.Conn
	for {
		nc, err = nats.Connect("nats://localhost:4222")
		if err == nil {
			break
		}
		log.Println("Error nats connection:", err)
		time.Sleep(time.Second * 2)
	}

	configStr := "user=admin password=123456 dbname=base sslmode=disable"
	var db *sql.DB
	for {
		db, err = sql.Open("postgres", configStr)
		if err == nil {
			break
		}
		log.Println("Error opening database connection:", err)
		time.Sleep(time.Second * 2)
	}

	prepare(db)

	var orderCache sync.Map
	go func() {
		loadOrders(db, &orderCache)
	}()

	for {
		_, err = nc.Subscribe("subject", func(msg *nats.Msg) {
			handleMsg(msg, db, &orderCache)
		})
		if err == nil {
			break
		}
		log.Println("Error subscribing:", err)
		time.Sleep(time.Second * 2)
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/search/{uid}", func(w http.ResponseWriter, r *http.Request) {
		handleSearch(w, r, &orderCache)
	}).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	fmt.Println("Starting HTTP server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
