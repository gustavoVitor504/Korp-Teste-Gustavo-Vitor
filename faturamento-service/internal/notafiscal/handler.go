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

type Handler struct { // handler com dependência service
	service *NotaFiscalService
}

func NewHandler(service *NotaFiscalService) *Handler { // construtor
	return &Handler{service: service} // recebe service no main
}

type itemRequest struct { // formato que recebe do http
	ProdutoID  uint `json:"produtoId"`
	Quantidade int  `json:"quantidade"`
}

type criarRequest struct { // o body da requisição da nota
	Status StatusNotaFiscal `json:"status"`
	Itens  []itemRequest    `json:"itens"`
}

type response struct { // formato de resposta para frontend , principalmente pelo numero formatado
	ID              uint             `json:"id"`
	Numero          uint             `json:"numero"`
	NumeroFormatado string           `json:"numeroFormatado"`
	Status          StatusNotaFiscal `json:"status"`
	Itens           []NotaFiscalItem `json:"itens"`
}

func toResponse(n NotaFiscal) response { // função pra não ter que repetir a formatação da resposta nas funções principais
	return response{
		ID: n.ID, Numero: n.Numero, NumeroFormatado: n.NumeroFormatado(),
		Status: n.Status, Itens: n.Itens,
	}
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req criarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // converte o json
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.Itens) == 0 { // tratamento de erro se não houver itens
		http.Error(w, "selecione ao menos um produto", http.StatusBadRequest)
		return
	}

	itens := make([]ItemNota, 0, len(req.Itens))  // converte itens em ItemNota
	for _, i := range req.Itens { // percorre cada item
		itens = append(itens, ItemNota{ProdutoID: i.ProdutoID, Quantidade: i.Quantidade})
	} // append para adicionar em itens

	nota, err := h.service.Criar(itens) // chama a função do service e entrega itens já no formato
	if err != nil { // trata qualquer erro
		writeErro(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json") // header para informar tipo do conteudo
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toResponse(*nota)) // utiliza a função toResponse para formatar para frontend
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	notasFiscais, err := h.service.Listar() // chama função listar do service
	if err != nil { // trata qualquer erro da função
		http.Error(w, "erro ao listar nota fiscal", http.StatusInternalServerError)
		return
	}
	resp := make([]response, 0, len(notasFiscais)) // converte as notas recebidas no modelo response
	for _, n := range notasFiscais {
		resp = append(resp, toResponse(n))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeErro(w http.ResponseWriter, err error) {
	var errSaldo *estoqueclient.ErrSaldoInsuficiente // variável capaz de receber esse tipo de erro

	switch {
	case errors.Is(err, estoqueclient.ErrIndisponivel): // se estoque estiver indisponível
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "estoque_indisponivel",
			"mensagem": "O serviço de estoque está fora do ar no momento. Tente novamente em instantes.",
		})
	case errors.As(err, &errSaldo): // se não ter saldo suficiente
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict) // retorna 409 e as info úteis
		json.NewEncoder(w).Encode(map[string]any{
			"erro": "saldo_insuficiente", "produtoId": errSaldo.ProdutoID,
			"produto": errSaldo.Produto, "disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado, "mensagem": errSaldo.Error(),
		})
	case errors.Is(err, estoqueclient.ErrProdutoNaoEncontrado), // entrada inváida
		errors.Is(err, ErrQuantidadeInvalida),
		errors.Is(err, ErrItensObrigatorios):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default: // se não for erro conhecido retorna o inesperado
		log.Printf("erro inesperado ao criar/imprimir nota fiscal: %v", err)
		http.Error(w, "erro ao criar nota fiscal", http.StatusInternalServerError)
	}
}
func (h *Handler) Imprimir(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	notaID, err := strconv.ParseUint(id, 10, 64) // converte string a inteiro parametros 10 casa decimal e 64 bits
	if err != nil { // se não converter retorna o erro
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	nota, err := h.service.Imprimir(uint(notaID)) // chama função imprimir do service
	if err != nil { // se ocorrer algum erro retorna função de erroImprimir
		writeErroImprimir(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toResponse(*nota))
}

func writeErroImprimir(w http.ResponseWriter, err error) { // parece WriteErro porém específica para imprimir
	var errSaldo *estoqueclient.ErrSaldoInsuficiente

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound): // se não encontrar essa nota
		http.Error(
			w,
			"nota fiscal não encontrada",
			http.StatusNotFound,
		)

	case errors.Is(err, ErrNotaNaoAberta): // se o status não for aberta para a nota
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "nota_nao_aberta",
			"mensagem": "A nota fiscal já foi fechada ou está cancelada.",
		})

	case errors.Is(err, estoqueclient.ErrIndisponivel): // se o estoque não está disponível
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		json.NewEncoder(w).Encode(map[string]any{
			"erro":     "estoque_indisponivel",
			"mensagem": "O serviço de estoque está fora do ar no momento. Tente novamente em instantes.",
		})

	case errors.As(err, &errSaldo): // se não tem saldo suficiente
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		json.NewEncoder(w).Encode(map[string]any{ // retorna uma resposta apropriada para front
			"erro":       "saldo_insuficiente",
			"produtoId":  errSaldo.ProdutoID,
			"produto":    errSaldo.Produto,
			"disponivel": errSaldo.Disponivel,
			"solicitado": errSaldo.Solicitado,
			"mensagem":   errSaldo.Error(),
		})

	case errors.Is(err, ErrQuantidadeInvalida), // se a quantidade for invalida por exemplo ir uma string
		errors.Is(err, ErrItensObrigatorios): // se não tiver itens

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

	default: // se nenhum erro for conhecido retorna inesperado
		log.Printf("erro inesperado ao criar/imprimir nota fiscal: %v", err)
		http.Error(
			w,
			"erro ao imprimir nota fiscal",
			http.StatusInternalServerError,
		)
	}
}
