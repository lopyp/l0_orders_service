package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sqldata "l0/internal"
	"l0/model"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

var orderCache = make(map[string]model.Order)

// loadOrders loads data from the database and populates the orderCache
func loadOrders(db *sql.DB) {
	rows, err := db.Query("select orders.*, delivery.name, delivery.phone, delivery.zip, delivery.city, delivery.address, delivery.region, delivery.email, payment.transaction, payment.request_id, payment.currency, payment.provider, payment.amount, payment.payment_dt, payment.bank, payment.delivery_cost, payment.goods_total, payment.custom_fee, items.chrt_id, items.track_number, items.price, items.rid, items.name, items.sale, items.size, items.total_price, items.nm_id, items.brand, items.status from orders join delivery on orders.order_uid = delivery.order_uid join payment on orders.order_uid = payment.order_uid join items on orders.order_uid = items.order_uid") // Your SQL query here
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

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
			log.Fatal(err)
		}
		orderData.Items = append(orderData.Items, item)

		orderCache[orderData.OrderUID] = orderData

		log.Printf("Loaded order with ID %s into cache", orderData.OrderUID)
	}

	log.Printf("Loaded %d orders into cache", len(orderCache))
}

// handleMsg processes incoming NATS messages
func handleMsg(msg *nats.Msg, db *sql.DB) {
	var orderData model.Order
	err := json.Unmarshal(msg.Data, &orderData)
	if err != nil {
		return
	}
	sqldata.InsertData(db, orderData)
	orderCache[orderData.OrderUID] = orderData
}

// handleSearch handles the search request for a specific UID
func handleSearch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	uid := params["uid"]

	result, found := orderCache[uid]
	if !found {
		http.Error(w, "No data found for the given UID.", http.StatusNotFound)
		return
	}

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Convert the result to JSON and write it to the response
	json.NewEncoder(w).Encode(result)
}

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	connStr := "user=admin password=123456 dbname=base sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Load orders from the database into the cache
	loadOrders(db)

	// Subscribe to NATS messages
	_, err = nc.Subscribe("subject", func(msg *nats.Msg) {
		handleMsg(msg, db)
	})
	if err != nil {
		log.Fatal("Error subscribing:", err)
	}

	r := mux.NewRouter()

	// Route for handling the search request (REST API endpoint)
	r.HandleFunc("/api/search/{uid}", handleSearch).Methods("GET")

	// Serve the index.html page
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Start the HTTP server with the router
	fmt.Println("Starting HTTP server on port 8088...")
	log.Fatal(http.ListenAndServe(":8088", r))
}
