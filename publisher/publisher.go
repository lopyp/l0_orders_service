package main

import (
	"encoding/json"
	"fmt"
	"l0/model"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

// Импорт структуры данных из другого файла

func main() {
	// Подключение к серверу NATS Streaming
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Создание экземпляра структуры данных
	order := model.Order{
		OrderUID:    fmt.Sprintf("%d", rand.Int63()),
		TrackNumber: "1234567890",
		Entry:       "entry",

		Delivery: model.Delivery{
			Name:    "Name",
			Phone:   "555-555-5555",
			Zip:     "12345",
			City:    "ds",
			Address: "address",
			Region:  "reg",
			Email:   "name@email.com",
		},
		Payment: model.Payment{
			Transaction:  "1234567890",
			RequestID:    "123",
			Currency:     "USD",
			Provider:     "VISA",
			Amount:       100,
			PaymentDt:    int(time.Now().Unix()),
			Bank:         "Bank",
			DeliveryCost: 5,
			GoodsTotal:   105,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      123,
				TrackNumber: "1234567890",
				Price:       100,
				Rid:         "ABC123",
				Name:        "A",
				Sale:        0,
				Size:        "XL",
				TotalPrice:  100,
				NmID:        456,
				Brand:       "B",
				Status:      1,
			},
		},

		Locale:            "en_US",
		InternalSignature: "",
		CustomerID:        "1234567890",
		DeliveryService:   "USPS",
		Shardkey:          "1234567890",
		SmID:              123,
		DateCreated:       time.Now().Format("2006-01-02 15:04:05"),
		OofShard:          "OOF123",
	}

	// Кодирование структуры данных в JSON
	jsonData, err := json.Marshal(order)
	if err != nil {
		panic(err)
	}

	// Отправка сообщения в NATS Streaming
	err = nc.Publish("subject", jsonData)
	if err != nil {
		panic(err)
	}

	//Отправка нескольких сообщений в NATS Streaming
	//	for i := 0; i < 10; i++ {
	// 	// Изменение значения поля OrderUID
	// 	order.OrderUID = fmt.Sprintf("%d", i)

	// 	// Кодирование структуры данных в JSON
	// 	jsonData, err := json.Marshal(order)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// Отправка сообщения в NATS Streaming
	// 	err = nc.Publish("subject", jsonData)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
}
