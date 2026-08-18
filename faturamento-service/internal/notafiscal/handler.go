package notafiscal

import (
	"encoding/json"
	"errors"
	"go-nfe-backend-api/internal/estoqueclient"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type Handler struct {
	service *NotaFiscalService
}

func NewHandler(service *NotaFiscalService) *Handler {
	return &Handler{service: service}
}

type itemRequest struct {
	ProdutoID  uint `json:"produtoId"`
	Quantidade int  `json:"quantidade"`
}

type criarRequest struct {
	Status StatusNotaFiscal `json:"status"`
	Itens  []itemRequest    `json:"itens"`
}

type response struct {
	ID              uint             `json:"id"`
	Numero          uint             `json:"numero"`
	NumeroFormatado string           `json:"numeroFormatado"`
	Status          StatusNotaFiscal `json:"status"`
	Itens           []NotaFiscalItem `json:"itens"`
}

func toResponse(n NotaFiscal) response {
	return response{
		ID: n.ID, Numero: n.Numero, NumeroFormatado: n.NumeroFormatado(),
		Status: n.Status, Itens: n.Itens,
	}
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req criarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.Itens) == 0 {
		http.Error(w, "selecione ao menos um produto", http.StatusBadRequest)
		return
	}

	itens := make([]ItemNota, 0, len(req.Itens))
	for _, i := range req.Itens {
		itens = append(itens, ItemNota{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	}

	nota, err := h.service.Criar(itens)
	if err != nil {
		writeErro(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toResponse(*nota))
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	notasFiscais, err := h.service.Listar()
	if err != nil {
		http.Error(w, "erro ao listar nota fiscal", http.StatusInternalServerError)
		return
	}
	resp := make([]response, 0, len(notasFiscais))
	for _, n := range notasFiscais {
		resp = append(resp, toResponse(n))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeErro(w http.ResponseWriter, err error) {
	var errSaldo *estoqueclient.ErrSaldoInsuficiente

	switch {
	case errors.Is(err, estoqueclient.ErrIndisponivel):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "estoque_indisponivel",
			"mensagem": "O serviço de estoque está fora do ar no momento. Tente novamente em instantes.",
		})
	case errors.As(err, &errSaldo):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"erro": "saldo_insuficiente", "produtoId": errSaldo.ProdutoID,
			"produto": errSaldo.Produto, "disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado, "mensagem": errSaldo.Error(),
		})
	case errors.Is(err, estoqueclient.ErrProdutoNaoEncontrado),
		errors.Is(err, ErrQuantidadeInvalida),
		errors.Is(err, ErrItensObrigatorios):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		log.Printf("erro inesperado ao criar/imprimir nota fiscal: %v", err)
		http.Error(w, "erro ao criar nota fiscal", http.StatusInternalServerError)
	}
}
func (h *Handler) Imprimir(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	notaID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	nota, err := h.service.Imprimir(uint(notaID))
	if err != nil {
		writeErroImprimir(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toResponse(*nota))
}

func writeErroImprimir(w http.ResponseWriter, err error) {
	var errSaldo *estoqueclient.ErrSaldoInsuficiente

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		http.Error(
			w,
			"nota fiscal não encontrada",
			http.StatusNotFound,
		)

	case errors.Is(err, ErrNotaNaoAberta):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "nota_nao_aberta",
			"mensagem": "A nota fiscal já foi fechada ou está cancelada.",
		})

	case errors.Is(err, estoqueclient.ErrIndisponivel):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "estoque_indisponivel",
			"mensagem": "O serviço de estoque está fora do ar no momento. Tente novamente em instantes.",
		})

	case errors.As(err, &errSaldo):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		json.NewEncoder(w).Encode(map[string]any{
			"erro":       "saldo_insuficiente",
			"produtoId":  errSaldo.ProdutoID,
			"produto":    errSaldo.Produto,
			"disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado,
			"mensagem":   errSaldo.Error(),
		})

	case errors.Is(err, ErrQuantidadeInvalida),
		errors.Is(err, ErrItensObrigatorios):

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

	default:
		log.Printf("erro inesperado ao criar/imprimir nota fiscal: %v", err)
		http.Error(
			w,
			"erro ao imprimir nota fiscal",
			http.StatusInternalServerError,
		)
	}
}
