package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"orderbook"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type BookKeyType int

const BookKey = BookKeyType(1601486424)

type Response struct {
	Response interface{} `json:"response"`
	Error    string      `json:"error"`
}

// Interrupt returns when either (1) interrupt signal is received by
// the OS or (2) the given context is done.
func Interrupt(ctx context.Context) bool {
	appSignal := make(chan os.Signal, 1)
	signal.Notify(appSignal, os.Interrupt)
	select {
	case <-appSignal:
		return true
	case <-ctx.Done():
		return false
	}
}

func GoInterrupt(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		if Interrupt(ctx) {
			cancel()
		}
	}()
}

func handler(writer http.ResponseWriter, request *http.Request) {
	request.Context()
	fmt.Printf("%v\n", request)
}

func addOrder(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	fmt.Printf("body:\n%v", body)
	if err != nil {
		// TODO
		panic(err)
	}
	// fmt.Fprint(writer, "TODO")
}

func queryOrder(writer http.ResponseWriter, request *http.Request) {

}

func cancelOrder(writer http.ResponseWriter, request *http.Request) {

}

type bookResponse struct {
	Symbol string                  `json:"symbol"`
	Asks   []orderbook.ClientLevel `json:"asks"`
	Bids   []orderbook.ClientLevel `json:"bids"`
}

func book(writer http.ResponseWriter, request *http.Request) {
	depths, ok := request.URL.Query()["depth"]
	if !ok {
		depths = []string{"20"}
	}
	depth, err := strconv.Atoi(depths[len(depths)-1])
	if err != nil {
		depth = 20
	}

	b := request.Context().Value(BookKey).(*orderbook.Book)
	snapshot := b.GetSnapshot(depth)

	r := Response{
		Response: bookResponse{
			Symbol: "generic",
			Asks:   snapshot.Asks,
			Bids:   snapshot.Bids,
		},
		Error: "",
	}
	encoded, err := json.Marshal(r)
	if err != nil {
		fmt.Printf("WRN book: Error while encoding response: %v\n", err)
		writer.WriteHeader(500)
	} else {
		fmt.Fprint(writer, string(encoded))
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	GoInterrupt(ctx, cancel)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/orders/", addOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", queryOrder).Methods("GET")
	router.HandleFunc("/orders/{id}", cancelOrder).Methods("DELETE")
	router.HandleFunc("/book/", book).Methods("GET")

	b := orderbook.NewBook()
	for price := 11; price <= 30; price++ {
		for i := 0; i < price; i++ {
			order := orderbook.ClientOrder{
				Side:             orderbook.SideBuy,
				OriginalQuantity: decimal.NewFromInt(int64(2 * price)),
				ExecutedQuantity: decimal.Zero,
				Price:            decimal.NewFromInt(int64(price)),
				ID:               fmt.Sprintf("%d_%d", price, i),
				Type:             orderbook.TypeLimit,
			}
			if price >= 21 {
				order.Side = orderbook.SideSell
			}
			b.AddOrder(order)
		}
	}
	f := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), BookKey, b)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	server := &http.Server{Addr: ":7701", Handler: f(router)}
	go func() {
		fmt.Printf("Starting server at port 7701...\n")
		server.ListenAndServe()
	}()

	<-ctx.Done()
	fmt.Printf("Shutting down...\n")
	server.Shutdown(context.Background())
}
