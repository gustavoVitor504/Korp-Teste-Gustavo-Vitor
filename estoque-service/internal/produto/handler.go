package produto

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

type consultarRequest struct {
	Ids []uint `json:"ids"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var produto Produto
	if err := json.NewDecoder(r.Body).Decode(&produto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := h.service.Criar(&produto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(produto)
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	produtos, err := h.service.Listar()
	if err != nil {
		http.Error(w, "erro ao listar produtos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}

type itemBaixaRequest struct {
	ProdutoID  uint `json:"produtoId"`
	Quantidade int  `json:"quantidade"`
}

type baixarSaldoRequest struct {
	Itens []itemBaixaRequest `json:"itens"`
}

type produtoInfoResponse struct {
	ID        uint   `json:"id"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

// BaixarSaldo é o endpoint interno chamado pelo Serviço de Faturamento.
func (h *Handler) BaixarSaldo(w http.ResponseWriter, r *http.Request) {
	var req baixarSaldoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.Itens) == 0 {
		http.Error(w, "nenhum item informado", http.StatusBadRequest)
		return
	}

	itens := make([]ItemQuantidade, 0, len(req.Itens))
	for _, i := range req.Itens {
		itens = append(itens, ItemQuantidade{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	}

	atualizados, err := h.service.BaixarSaldo(itens)
	if err != nil {
		writeErro(w, err)
		return
	}

	resp := make([]produtoInfoResponse, 0, len(atualizados))
	for _, p := range atualizados {
		resp = append(resp, produtoInfoResponse{ID: p.ID, Descricao: p.Descricao, Saldo: p.Saldo})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"itens": resp})
}

// ReporSaldo é o endpoint interno de compensação (saga rollback).
func (h *Handler) ReporSaldo(w http.ResponseWriter, r *http.Request) {
	var req baixarSaldoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	itens := make([]ItemQuantidade, 0, len(req.Itens))
	for _, i := range req.Itens {
		itens = append(itens, ItemQuantidade{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	}

	if err := h.service.ReporSaldo(itens); err != nil {
		http.Error(w, "erro ao repor saldo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeErro(w http.ResponseWriter, err error) {
	var errSaldo *ErrSaldoInsuficiente

	switch {
	case errors.As(err, &errSaldo):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"erro":       "saldo_insuficiente",
			"produtoId":  errSaldo.ProdutoID,
			"produto":    errSaldo.Produto,
			"disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado,
		})
	case errors.Is(err, ErrProdutoNaoEncontrado), errors.Is(err, ErrQuantidadeInvalida):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		log.Printf("erro inesperado ao processar baixa de estoque: %v", err)
		http.Error(w, "erro ao processar baixa de estoque", http.StatusInternalServerError)
	}
}
func (h *Handler) Consultar(w http.ResponseWriter, r *http.Request) {
	var req consultarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	produtos, err := h.service.ConsultarPorIds(req.Ids)
	if err != nil {
		http.Error(w, "erro ao consultar produtos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}
