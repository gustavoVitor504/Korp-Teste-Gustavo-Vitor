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
		http.Error(w, "JSON inválido", http.StatusBadRequest) // decoder verifica se é
		return               // possivel transformar esse body do json em um objeto Produto
	}
	if err := h.service.Criar(&produto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // aqui ja foi convertido em obj
		return                  // agora faz o service tentar criar e se der erro retorna
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(produto) // retorna o produto criado em forma de json
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	produtos, err := h.service.Listar()
	if err != nil { // no get só o tratamento do service é necessário pois não recebemos nada
		http.Error(w, "erro ao listar produtos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos) // se não ter erro devolve lista de produtos em json
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
		return // API de faturamento requisita baixar saldo na impressão e verifica formato json
	}
	if len(req.Itens) == 0 {
		http.Error(w, "nenhum item informado", http.StatusBadRequest)
		return // se não tiver item cai no erro
	}
            // novo objeto / tamanho atual / capacidade(o que tem de produtos para cadastrar)
	itens := make([]ItemQuantidade, 0, len(req.Itens)) // vai transformar em um novo formato
	for _, i := range req.Itens {
		itens = append(itens, ItemQuantidade{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	}

	atualizados, err := h.service.BaixarSaldo(itens)
	if err != nil {
		writeErro(w, err) // aqui joga os dados transformados na service para verificar possiveis erros
		return
	}

	resp := make([]produtoInfoResponse, 0, len(atualizados))
	for _, p := range atualizados { // transforma os dados dos produtos alterados para id desc e saldo
		resp = append(resp, produtoInfoResponse{ID: p.ID, Descricao: p.Descricao, Saldo: p.Saldo})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"itens": resp}) // devolve json dos produtos atualizados
}

// ReporSaldo é o endpoint interno de compensação (saga rollback).
func (h *Handler) ReporSaldo(w http.ResponseWriter, r *http.Request) {
	var req baixarSaldoRequest // Api chama se caso der erro após baixar o saldo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return        // verifica formato do json para repor o saldo
	}

	itens := make([]ItemQuantidade, 0, len(req.Itens))
	for _, i := range req.Itens {     // muda formato do json para o mesmo que se usa para faturar
		itens = append(itens, ItemQuantidade{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	}

	if err := h.service.ReporSaldo(itens); err != nil {
		http.Error(w, "erro ao repor saldo", http.StatusInternalServerError)
		return     // verifica erro de requisição do service
	}
	w.WriteHeader(http.StatusNoContent) // retorna o statur para o backend
}

func writeErro(w http.ResponseWriter, err error) {
	var errSaldo *ErrSaldoInsuficiente    // função para padronizar os erros 

	switch {
	case errors.As(err, &errSaldo):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{   // trata dos erros conhecidos
			"erro":       "saldo_insuficiente",
			"produtoId":  errSaldo.ProdutoID,   // type de ErrSaldoInsuficiente no errors.go
			"produto":    errSaldo.Produto,
			"disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado,
		})
	case errors.Is(err, ErrProdutoNaoEncontrado), errors.Is(err, ErrQuantidadeInvalida):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		log.Printf("erro inesperado ao processar baixa de estoque: %v", err)
		http.Error(w, "erro ao processar baixa de estoque", http.StatusInternalServerError)
	} // se não estiver nos erros esperados devolve um default
}
func (h *Handler) Consultar(w http.ResponseWriter, r *http.Request) {
	var req consultarRequest  // consulta diferente do Listar() por devolve apenas os produtos pedidos
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return    // verifica formato do json
	}
	produtos, err := h.service.ConsultarPorIds(req.Ids)
	if err != nil { 
		http.Error(w, "erro ao consultar produtos", http.StatusInternalServerError)
		return   // trata dos erros provindos do service ex: ID inválido 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)   // devolve json com os produtos pedidos
}
// metodo post para consultar serve para centralizar em uma única requisição ao invés
// de mandar vários http get com o id do produto requisitado