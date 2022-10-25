package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ydm/orderbook"
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

//nolint:forbidigo
func logf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func respond(writer http.ResponseWriter, response Response) {
	encoded, err := json.Marshal(response)
	if err != nil {
		logf("WRN: Error while encoding response: %v\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		fmt.Fprint(writer, string(encoded))
	}
}

// +------------------+
// | (1) Submit order |
// +------------------+

func addOrder(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respond(writer, Response{Response: nil, Error: err.Error()})

		return
	}

	var order orderbook.ClientOrder
	if err := json.Unmarshal(body, &order); err != nil {
		respond(writer, Response{Response: nil, Error: err.Error()})

		return
	}

	book, ok := request.Context().Value(BookKey).(*orderbook.Book)
	if !ok {
		panic("")
	}

	if err := book.AddOrder(order); err != nil {
		respond(writer, Response{Response: nil, Error: err.Error()})

		return
	}

	// Return order's current status.
	order, err = book.GetOrder(order.ID)
	if err != nil {
		respond(writer, Response{Response: nil, Error: err.Error()})

		return
	}

	respond(writer, Response{Response: order, Error: ""})
}

// +------------------+
// | (2) Cancel order |
// +------------------+

func cancelOrder(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	orderID := vars["id"]

	b, ok := request.Context().Value(BookKey).(*orderbook.Book)
	if !ok {
		panic("")
	}

	if err := b.CancelOrder(orderID); err == nil {
		respond(writer, Response{Response: true, Error: ""})
	} else {
		respond(writer, Response{Response: false, Error: err.Error()})
	}
}

// +---------------+
// | (3) Get order |
// +---------------+

func queryOrder(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	orderID := vars["id"]

	b, ok := request.Context().Value(BookKey).(*orderbook.Book)
	if !ok {
		panic("")
	}

	if order, err := b.GetOrder(orderID); err != nil {
		respond(writer, Response{Response: nil, Error: err.Error()})
	} else {
		respond(writer, Response{Response: order, Error: ""})
	}
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
	depths, depthsOK := request.URL.Query()["depth"]
	if !depthsOK {
		depths = []string{"20"}
	}

	depth, err := strconv.Atoi(depths[len(depths)-1])
	if err != nil {
		depth = 20
	}

	book, bookOK := request.Context().Value(BookKey).(*orderbook.Book)
	if !bookOK {
		panic("")
	}

	snapshot := book.GetSnapshot(depth)

	respond(writer, Response{
		Response: bookResponse{
			Asks: snapshot.Asks,
			Bids: snapshot.Bids,
		},
		Error: "",
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

	book := orderbook.NewBook()
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), BookKey, book)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	var server http.Server
	server.Addr = ":7701"
	server.Handler = handler(router)

	go func() {
		logf("INF: Starting server at port 7701...\n")

		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()

	logf("INF: Shutting down...\n")

	if err := server.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}
