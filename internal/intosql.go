package sqldata

import (
	"database/sql"
	"l0/model"
	"log"

	_ "github.com/lib/pq"
)

func CreateTables(db *sql.DB) {

	execQuery(db, `CREATE TABLE IF NOT EXISTS orders (
        order_uid text,
        track_number text,
        entry text,
        locale text,
        internal_signature text,
        customer_id text,
        delivery_service text,
        shardkey text,
        sm_id integer,
        date_created timestamp,
        oof_shard text
    )`)

	execQuery(db, `CREATE TABLE IF NOT EXISTS deliveries (
        order_uid text,
        name text,
        phone text,
        zip text,
        city text,
        address text,
        region text,
        email text
    )`)

	execQuery(db, `CREATE TABLE IF NOT EXISTS payments (
        order_uid text,
        transaction text,
        request_id text,
        currency text,
        provider text,
        amount float,
        payment_dt integer,
        bank text,
        delivery_cost float,
        goods_total float,
        custom_fee float
    )`)

	execQuery(db, `CREATE TABLE IF NOT EXISTS items (
        order_uid text,
        chrt_id integer,
        track_number text,
        price float,
        rid text,
        name text,
        sale float,
        size text,
        total_price float,
        nm_id integer,
        brand text,
        status integer
    )`)
}

func InsertData(db *sql.DB, data model.Order) {
	execQuery(db, `INSERT INTO orders (
        order_uid,
        track_number,
        entry,
        locale,
        internal_signature,
        customer_id,
        delivery_service,
        shardkey,
        sm_id,
        date_created,
        oof_shard
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		data.OrderUID,
		data.TrackNumber,
		data.Entry,
		data.Locale,
		data.InternalSignature,
		data.CustomerID,
		data.DeliveryService,
		data.Shardkey,
		data.SmID,
		data.DateCreated,
		data.OofShard,
	)

	execQuery(db, `INSERT INTO delivery (
        order_uid,
        name,
        phone,
        zip,
        city,
        address,
        region,
        email
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		data.OrderUID,
		data.Delivery.Name,
		data.Delivery.Phone,
		data.Delivery.Zip,
		data.Delivery.City,
		data.Delivery.Address,
		data.Delivery.Region,
		data.Delivery.Email,
	)

	execQuery(db, `INSERT INTO payment (
        order_uid,
        transaction,
        request_id,
        currency,
        provider,
        amount,
        payment_dt,
        bank,
        delivery_cost,
        goods_total,
        custom_fee
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		data.OrderUID,
		data.Payment.Transaction,
		data.Payment.RequestID,
		data.Payment.Currency,
		data.Payment.Provider,
		data.Payment.Amount,
		data.Payment.PaymentDt,
		data.Payment.Bank,
		data.Payment.DeliveryCost,
		data.Payment.GoodsTotal,
		data.Payment.CustomFee,
	)

	for _, item := range data.Items {
		execQuery(db, `INSERT INTO items (
            order_uid,
            chrt_id,
            track_number,
            price,
            rid,
            name,
            sale,
            size,
            total_price,
            nm_id,
            brand,
            status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			data.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
	}

}
func execQuery(db *sql.DB, query string, args ...interface{}) {
	_, err := db.Exec(query, args...)
	if err != nil {
		log.Fatal(err)
	}
}
