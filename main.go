package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"github.com/go-chi/chi/v5"
)

// Handler Functions
func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParam := parameters{}
	err := decoder.Decode(&reqParam)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if len(reqParam.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	respondWithJson(w, http.StatusOK, returnVal{Valid: true})

	
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
		
	// }


	// error := json.Unmarshal(body, &reqParam)
	// if error != nil {
	// 	fmt.Println("error")
	// }
	// if len(reqParam.Body) > 140 {
	// 	fmt.Println("error")
	// }
	// data, error := json.Marshal(reqParam)
	// if error != nil {
	
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(200)
	// w.Write(data)
	// fmt.Println("Good Request/Response!")

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
	apir.Post("/validate_chirp", handleChirpValidation)


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