package main

import (
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"github.com/stripe/stripe-go/v71"
	"github.com/xngln/photo-server/graph"
	"github.com/xngln/photo-server/graph/generated"
)

const defaultPort = "8080"

type SuccessPageData struct {
	FullsizeURL string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	var mb int64 = 1 << 20

	router := chi.NewRouter()

	stripe.Key = os.Getenv("STRIPE_SECRET")

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8080",
			"http://localhost:3000",
			"http://photo.davidxliu.com",
			"https://photo.davidxliu.com",
		},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxMemory:     32 * mb,
		MaxUploadSize: 50 * mb,
	})
	srv.Use(extension.Introspection{})

	router.Handle("/", playground.Handler("Photo Store", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		fullsizeURL, _ := hex.DecodeString(r.URL.Query().Get("downloadurl"))
		tmpl := template.Must(template.ParseFiles("views/success.html"))
		data := SuccessPageData{
			FullsizeURL: string(fullsizeURL),
		}
		tmpl.Execute(w, data)
	})
	router.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("views/cancel.html"))
		tmpl.Execute(w, nil)
	})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
