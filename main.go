package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// --- ПРОСТОЕ ПРИЛОЖЕНИЕ: только JSON ---

type Order struct {
	OrderUID        string    `json:"order_uid"`
	TrackNumber     string    `json:"track_number"`
	Entry           string    `json:"entry"`
	Locale          string    `json:"locale"`
	InternalSig     *string   `json:"internal_signature,omitempty"`
	CustomerID      string    `json:"customer_id"`
	DeliveryService string    `json:"delivery_service"`
	Shardkey        string    `json:"shardkey"`
	SmID            int       `json:"sm_id"`
	DateCreated     time.Time `json:"date_created"`
	OofShard        string    `json:"oof_shard"`

	Delivery *Delivery `json:"delivery,omitempty"`
	Payment  *Payment  `json:"payment,omitempty"`
	Items    []Item    `json:"items,omitempty"`
}

type Delivery struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Zip     *string `json:"zip,omitempty"`
	City    string  `json:"city"`
	Address string  `json:"address"`
	Region  *string `json:"region,omitempty"`
	Email   *string `json:"email,omitempty"`
}

type Payment struct {
	Transaction  string  `json:"transaction"`
	RequestID    *string `json:"request_id,omitempty"`
	Currency     string  `json:"currency"`
	Provider     string  `json:"provider"`
	Amount       int     `json:"amount"`
	PaymentDt    int64   `json:"payment_dt"`
	Bank         *string `json:"bank,omitempty"`
	DeliveryCost int     `json:"delivery_cost"`
	GoodsTotal   int     `json:"goods_total"`
	CustomFee    int     `json:"custom_fee"`
}

// Item - это предмет покупки
type Item struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type apiError struct {
	Error string `json:"error"`
}

func main() {
	dsn := defaultDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("cannot connect to DB: %v", err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// GET /order?id=<order_uid>
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "missing id (order_uid)"})
			return
		}

		ctx := r.Context()
		ord := Order{}
		var internalSig sql.NullString

		// order + delivery
		qOrder := `
			SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email
			FROM orders o
			LEFT JOIN deliveries d ON d.order_uid = o.order_uid
			WHERE o.order_uid = $1`

		var del Delivery
		var zip, region, email sql.NullString
		row := db.QueryRowContext(ctx, qOrder, id)
		if err := row.Scan(
			&ord.OrderUID, &ord.TrackNumber, &ord.Entry, &ord.Locale, &internalSig,
			&ord.CustomerID, &ord.DeliveryService, &ord.Shardkey, &ord.SmID, &ord.DateCreated, &ord.OofShard,
			&del.Name, &del.Phone, &zip, &del.City, &del.Address, &region, &email,
		); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, apiError{Error: "order not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		if internalSig.Valid {
			ord.InternalSig = &internalSig.String
		}
		// attach delivery if present
		if del.Name != "" || del.Phone != "" {
			if zip.Valid {
				del.Zip = &zip.String
			}
			if region.Valid {
				del.Region = &region.String
			}
			if email.Valid {
				del.Email = &email.String
			}
			ord.Delivery = &del
		}

		// payment (optional)
		qPay := `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
				FROM payments WHERE order_uid = $1`
		var p Payment
		var reqID, bank sql.NullString
		if err := db.QueryRowContext(ctx, qPay, id).Scan(
			&p.Transaction, &reqID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt, &bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
		); err == nil {
			if reqID.Valid {
				p.RequestID = &reqID.String
			}
			if bank.Valid {
				p.Bank = &bank.String
			}
			ord.Payment = &p
		} else if err != sql.ErrNoRows {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		// items (0..n)
		qItems := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
				FROM order_items WHERE order_uid = $1 ORDER BY chrt_id`
		rows, err := db.QueryContext(ctx, qItems, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var it Item
			if err := rows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err != nil {
				writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
				return
			}
			ord.Items = append(ord.Items, it)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, ord)
	})

	addr := getEnv("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func defaultDSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" { // готовая строка подключения
		return v
	}
	host := getEnv("PGHOST", "127.0.0.1")
	port := getEnv("PGPORT", "6432")
	user := getEnv("PGUSER", "yemtsova_anna")
	pass := getEnv("PGPASSWORD", "admin!!!")
	db := getEnv("PGDATABASE", "rwb_data")
	ssl := getEnv("PGSSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, db, ssl)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
