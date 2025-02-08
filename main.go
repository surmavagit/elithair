package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
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
	address := "localhost:8080"
	var err error
	if len(os.Args) == 3 && (os.Args[1] == "--address" || os.Args[1] == "-a") {
		address = os.Args[2]
	} else if len(os.Args) > 1 {
		fmt.Println(`
            -a localhost:8080   --address localhost:8080    specify address (default is localhost:8080)
            -h                  --help                      print this help (all invalid commands print this help too)
            `)
		return
	}

	err = godotenv.Load()
	if err != nil {
		log.Println(err.Error())
		return
	}

	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Get("/authors/{author_id}", apiHandler)
	r.Mount("/api", apiRouter)

	server := http.Server{
		Addr:    address,
		Handler: middlewareCors(r),
	}
	fmt.Println("listening on: ", server.Addr)
	server.ListenAndServe()
}

type DB struct {
	*sql.DB
}

func dbConnect() (*DB, error) {
	dbhost, err := loadEnv("dbhost")
	if err != nil {
		return nil, err
	}
	dbport, err := loadEnv("dbport")
	if err != nil {
		return nil, err
	}
	dbuser, err := loadEnv("dbuser")
	if err != nil {
		return nil, err
	}
	dbpassword, err := loadEnv("dbpassword")
	if err != nil {
		return nil, err
	}
	dbname, err := loadEnv("dbname")
	if err != nil {
		return nil, err
	}

	var psqlInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", dbhost, dbport, dbuser, dbpassword, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	return &DB{db}, err
}

func loadEnv(name string) (string, error) {
	variable, found := os.LookupEnv(name)
	if !found || variable == "" {
		return "", fmt.Errorf("%s env variable is not defined", name)
	}
	return variable, nil
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	authorId, err := strconv.Atoi(chi.URLParam(r, "author_id"))
	if err != nil {
		respondWithError(w, 400, "invalid author id")
		return
	}

	db, err := dbConnect()
	if err != nil {
		log.Println("can't connect to the db: ", err.Error())
		respondWithError(w, 500, "something went wrong")
		return
	}
	defer db.Close()

	queryAuthor := "SELECT name, biography FROM author WHERE id=$1;"
	authorRow := db.QueryRow(queryAuthor, authorId)
	author := Author{AuthorId: authorId}
	err = authorRow.Scan(&author.Name, &author.Biography)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, "author not found")
		} else {
			log.Println("authorRow.Scan gives unusual error: ", err.Error())
			respondWithError(w, 500, "something went wrong")
		}
		return
	}

	queryBook := "SELECT title, year from book left join attribution on book_id=book.id WHERE author_id=$1;"
	bookRows, err := db.Query(queryBook, authorId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithJSON(w, 200, author)
		} else {
			log.Println("db.Query gives unusual error: ", err.Error())
			respondWithError(w, 500, "something went wrong")
		}
		return
	}
	defer bookRows.Close()

	for bookRows.Next() {
		b := Book{}
		err := bookRows.Scan(&b.Title, &b.Year)
		if err != nil {
			log.Println("bookRows.Scan gives unusual error: ", err.Error())
			respondWithError(w, 500, "something went wrong")
		}
		author.Books = append(author.Books, b)
	}
	respondWithJSON(w, 200, author)
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
