package main

import (
	"log"
	"net/http"

	"go-nfe-backend-api/config"
	"go-nfe-backend-api/internal/produto"
)

func main() {
	db := config.ConnectDatabase() // aponta pro banco estoque_db
	db.AutoMigrate(&produto.Produto{})

	repository := produto.NewRepository(db) // injeção de dependência
	service := produto.NewService(repository)
	handler := produto.NewHandler(service)

	mux := http.NewServeMux()  // mapeamento de endpoints
	mux.HandleFunc("GET /api/produtos", handler.Listar)
	mux.HandleFunc("POST /api/produtos", handler.Criar)
	mux.HandleFunc("POST /api/produtos/baixar-saldo", handler.BaixarSaldo)
	mux.HandleFunc("POST /api/produtos/repor-saldo", handler.ReporSaldo)
	mux.HandleFunc("POST /api/produtos/consultar", handler.Consultar)

	log.Println("Serviço de Estoque em http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", enableCORS(mux)))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {  // tipo de requisição options não passa para o handler
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
