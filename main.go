package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

type DB struct {
	*sql.DB
}

func dbConnect() (*DB, error) {
	var psqlInfo = "host=localhost port=8080 user=elithairuser password=elithairpassword dbname=elithair sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	return &DB{db}, err
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

	db, err := dbConnect()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	} else {
		log.Println("connected to db")
	}

	queryAuthor := "SELECT * FROM author WHERE author.id=1;"
	//queryBook := "SELECT title, year from book left join attribution on  book_id = book.id WHERE author_id=$1;"
	rows := db.QueryRow(queryAuthor, authorId)
	log.Println("db response received")
	a := Author{}
	err = rows.Scan(&a.AuthorId, &a.Name, &a.Biography)
	if err != nil {
		respondWithError(w, 404, "author not found")
		db.Close()
		return
	}

	//first := Book{
	//	Title: "1984",
	//	Year:  1949,
	//}
	//second := Book{
	//	Title: "Animal Farm",
	//	Year:  1945,
	//}
	//orwell := Author{
	//	AuthorId:  1,
	//	Name:      "George Orwell",
	//	Biography: "British writer known for 1984 and Animal Farm.",
	//	Books:     []Book{first, second},
	//}

	db.Close()
	respondWithJSON(w, 200, a)
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
