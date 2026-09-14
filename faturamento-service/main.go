package main

import (
	"go-nfe-backend-api/config"
	"go-nfe-backend-api/internal/estoqueclient"
	"go-nfe-backend-api/internal/notafiscal"
	"log"
	"net/http"
	"os"
)

func main() {
	db := config.ConnectDatabase() // aponta para o banco

	if err := db.AutoMigrate(&notafiscal.NotaFiscal{}, &notafiscal.NotaFiscalItem{}); err != nil {
		log.Fatal("erro ao migrar banco de dados:", err) // fatal para parar server se não conseguir migrar
	}  // função de migrate para garantir que meu banco tenha as tabelas conforme meus models

	estoqueURL := os.Getenv("ESTOQUE_SERVICE_URL")
	if estoqueURL == "" {
		estoqueURL = "http://localhost:8081/api"
	}
	client := estoqueclient.New(estoqueURL) // cliente responsável por comunicar com estoque

	repository := notafiscal.NewRepository(db) // injeção de dependência
	service := notafiscal.NewService(repository, client) // service recebe duas dependências para persistir dados no faturamento e no estoque usando o client
	handler := notafiscal.NewHandler(service)

	mux := http.NewServeMux()  // mux da biblioteca padrão do GO para associar metodo com URL

	mux.HandleFunc("GET /api/notas-fiscais", handler.Listar) // mapeamentos
	mux.HandleFunc("POST /api/notas-fiscais", handler.Criar)
	mux.HandleFunc("POST /api/notas-fiscais/{id}/imprimir", handler.Imprimir)

	log.Println("Serviço de Faturamento em http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", enableCORS(mux)))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return // retorna para não enviar a options para endpoint
		}
		next.ServeHTTP(w, r)
	})
}
