package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type Book struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

type Author struct {
	AuthorId  int    `json:"author_id"`
	Name      string `json:"name"`
	Biography string `json:"biography"`
	Books     []Book `json:"books"`
}

func main() {
	godotenv.Load()

	r := chi.NewRouter()
	r.HandleFunc("/", rootHandler)
	apiRouter := chi.NewRouter()
	apiRouter.Get("/authors/{author_id}", apiHandler)
	r.Mount("/api", apiRouter)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: middlewareCors(r),
	}
	fmt.Println("listening on: ", server.Addr)
	server.ListenAndServe()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 404, "this is an error")
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("got a request")
	authorId, err := strconv.Atoi(chi.URLParam(r, "author_id"))
	if err != nil {
		respondWithError(w, 400, "invalid author id")
		return
	}
	first := Book{
		Title: "1984",
		Year:  1949,
	}
	second := Book{
		Title: "Animal Farm",
		Year:  1945,
	}
	orwell := Author{
		AuthorId:  1,
		Name:      "George Orwell",
		Biography: "British writer known for 1984 and Animal Farm.",
		Books:     []Book{first, second},
	}

	if authorId == 1 {
		respondWithJSON(w, 200, orwell)
	} else {
		respondWithError(w, 404, "author not found")
	}
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with server error: %d - %s", code, msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{msg})
}
