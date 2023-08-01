package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"strings"
	// "sync"
	"os"
	"errors"
	"sort"
)

// Handler Functions
func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleCreateChirp(w http.ResponseWriter, r *http.Request) {
		type returnVal struct {
		Id int `json:"id"`
		Body string `json:"body"`
	}
	sanitizedString, err := handleChirpValidation(w, r)
	if err != nil {
		fmt.Println(err)
		return
	}
	database, err := NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	newDBStructure, err := database.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	newID := getNewChirpId(newDBStructure.Chirps)

	newDBStructure.Chirps[newID] = Chirp{Id: newID, Body: sanitizedString}
	
	updatedDB, err := json.Marshal(newDBStructure)
	if err != nil {
		log.Fatal(err)
	}
	erruh := os.WriteFile(database.Path, updatedDB, 0666)
	if erruh != nil {
		log.Fatal(err)
	}



	respondWithJson(w, http.StatusOK, returnVal{Id: newID, Body: sanitizedString})
}

func getNewChirpId(chirpMap map[int]Chirp) int {
	if len(chirpMap) == 0 {
		return 1
	}

	chirps := []Chirp{}
	for i := 1; i <= len(chirpMap); i++ {
		chirps = append(chirps, chirpMap[i])
	}
	//Sort chrips by Id field IOT increment new Id for new chirp
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id > chirps[j].Id})
	return chirps[0].Id + 1
}

func handleChirpValidation(w http.ResponseWriter, r *http.Request) (string, error){
	type parameters struct {
		Body string `json:"body"`
	}

	// type returnVal struct {
	// 	Id int `json:"id"`
	// 	Body string `json:"body"`
	// }

	decoder := json.NewDecoder(r.Body)
	reqParam := parameters{}
	err := decoder.Decode(&reqParam)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return "", errors.New("Couldn't decode parameters")
	}

	if len(reqParam.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return "", errors.New("Chirp is too long")
	}

cleanString := sanitizeString(reqParam.Body)
return cleanString, nil
// respondWithJson(w, http.StatusOK, returnVal{Id: 1, Body: cleanString})
}

func sanitizeString(s string) string {
	dirtyMap := make(map[string]bool)
	dirtyMap["kerfuffle"] = true
	dirtyMap["sharbert"] = true
	dirtyMap["fornax"] = true

	sanitized := []string{}
	for _, word := range strings.Split(s, " ") {
		_, ok := dirtyMap[strings.ToLower(word)]; if ok {
			sanitized = append(sanitized, "****")
		} else {
			sanitized = append(sanitized, word)
		}
	}
	return strings.Join(sanitized, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	errorStruct := errorResponse{Error: msg}

	respondWithJson(w, code, errorStruct)
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
	<html>

		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>

  </html>
	`, cfg.fileserverHits)))
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	cfg := apiConfig{
		fileserverHits: 0,
	}

	// mux := http.NewServeMux()
	



	mux := chi.NewRouter()
	apir := chi.NewRouter()
	adminRouter := chi.NewRouter()
	mux.Mount("/api", apir)
	mux.Mount("/admin", adminRouter)
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/app", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	apir.Get("/healthz", getHealth)
	// apir.Get("/metrics", cfg.handleMetrics)  SWITCH OUT METRICS ENDPOINT
	adminRouter.Get("/metrics", cfg.handleMetrics)
	apir.Post("/chirps", handleCreateChirp)


	corsMux := middlewareCors(mux)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
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
