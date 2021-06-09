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

func respond(w http.ResponseWriter, r Response) {
	encoded, err := json.Marshal(r)
	if err != nil {
		fmt.Printf("WRN Error while encoding response: %v\n", err)
		w.WriteHeader(500)
	} else {
		fmt.Fprint(w, string(encoded))
	}
}

// +------------------+
// | (1) Submit order |
// +------------------+

func addOrder(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		respond(writer, Response{Error: err.Error()})
		return
	}

	var order orderbook.ClientOrder
	err = json.Unmarshal(body, &order)
	if err != nil {
		respond(writer, Response{Error: err.Error()})
		return
	}

	b := request.Context().Value(BookKey).(*orderbook.Book)
	err = b.AddOrder(order)
	if err != nil {
		respond(writer, Response{Error: err.Error()})
		return
	}

	// Return order's current status.
	order, err = b.GetOrder(order.ID)
	if err != nil {
		respond(writer, Response{Error: err.Error()})
		return
	}
	respond(writer, Response{Response: order})
}

// +------------------+
// | (2) Cancel order |
// +------------------+

func cancelOrder(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]

	b := request.Context().Value(BookKey).(*orderbook.Book)
	err := b.CancelOrder(id)
	if err == nil {
		respond(writer, Response{Response: true})
	} else {
		respond(writer, Response{Error: err.Error()})
	}
}

// +---------------+
// | (3) Get order |
// +---------------+

func queryOrder(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]

	b := request.Context().Value(BookKey).(*orderbook.Book)
	order, err := b.GetOrder(id)
	if err != nil {
		respond(writer, Response{Error: err.Error()})
		return
	}
	respond(writer, Response{Response: order})
}

// +-------------------------+
// | (5) Order book snapshot |
// +-------------------------+

type bookResponse struct {
	// Symbol string                  `json:"symbol"`
	Asks []orderbook.ClientLevel `json:"asks"`
	Bids []orderbook.ClientLevel `json:"bids"`
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
	respond(writer, Response{
		Response: bookResponse{
			Asks: snapshot.Asks,
			Bids: snapshot.Bids,
		},
	})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if Interrupt(ctx) {
			cancel()
		}
	}()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/orders/", addOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", queryOrder).Methods("GET")
	router.HandleFunc("/orders/{id}", cancelOrder).Methods("DELETE")
	router.HandleFunc("/book/", book).Methods("GET")

	b := orderbook.NewBook()
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
