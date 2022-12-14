package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)

	router.Get("/getAllPrograms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dgClient := newClient()
		txn := dgClient.NewTxn()

		resp, err := txn.Query(context.Background(), getNodeData)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(resp.Json)
	})

	type Data struct {
		Result          float64 `json:"result,omitempty"`
		Number          string  `json:"number,omitempty"`
		Assign          float64 `json:"assign,omitempty"`
		Variable        string  `json:"variable,omitempty"`
		Num1            string  `json:"num1,omitempty"`
		Num2            string  `json:"num2,omitempty"`
		Option          string  `json:"option,omitempty"`
		Conditionresult string  `json:"conditionResult,omitempty"`
	}
	type Connections1 struct {
		Node  string `json:"node"`
		Input string `json:"input"`
	}
	type Input1 struct {
		Connections []Connections1 `json:"connections"`
	}
	type Input2 struct {
		Connections []Connections1 `json:"connections"`
	}
	type Inputs struct {
		Input1 Input1 `json:"input_1"`
		Input2 Input2 `json:"input_2"`
	}
	type Connections2 struct {
		Node   string `json:"node"`
		Output string `json:"output"`
	}
	type Output1 struct {
		Connections []Connections2 `json:"connections"`
	}
	type Outputs struct {
		Output1 Output1 `json:"output_1"`
	}
	type Nodesdata struct {
		Nid      int     `json:"id"`
		Name     string  `json:"name"`
		Data     Data    `json:"data"`
		Class    string  `json:"class"`
		HTML     string  `json:"html"`
		Typenode string  `json:"typenode"`
		Inputs   Inputs  `json:"inputs,omitempty"`
		Outputs  Outputs `json:"outputs,omitempty"`
		PosX     float64 `json:"pos_x"`
		PosY     float64 `json:"pos_y"`
	}

	type AutoGenerated struct {
		Nodesdata   []Nodesdata `json:"nodesData,omitempty"`
		Programname string      `json:"programName"`
	}

	router.Post("/setAllPrograms", func(w http.ResponseWriter, r *http.Request) {
		var p AutoGenerated

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		fmt.Println("set json: ", p)

		pb, err := json.Marshal(p)
		if err != nil {
			log.Fatal(err)
		}

		mu := &api.Mutation{
			CommitNow: true,
			SetJson:   pb,
		}
		log.Println("mutation object:", mu)

		dgClient := newClient()
		txn := dgClient.NewTxn()
		ctx := context.Background()
		_, err = txn.Mutate(ctx, mu)
		fmt.Println("json added to dgraph")
	})

	router.Post("/deleteProgram", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		triples := "<" + id + ">" + "<nodesData> * ."
		fmt.Println("ID:", id)

		mu := &api.Mutation{
			DelNquads: []byte(triples),
		}

		req := &api.Request{
			Mutations: []*api.Mutation{mu},
			CommitNow: true,
		}

		dgClient := newClient()
		ctx := context.Background()
		txn := dgClient.NewTxn()
		_, err := txn.Do(ctx, req)

		if err != nil {
			log.Fatal(err)
		}
	})

	ServerGo := &http.Server{
		Addr:           ":5000",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server running in localhost:5000")
	err := ServerGo.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
