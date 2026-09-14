package estoqueclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client { // construtor usa model para criar novo cliente
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}  // timeout para evitar que a requisição trave se o estoque não responder
}

type ItemBaixa struct {
	ProdutoID  uint `json:"produtoId"`
	Quantidade int  `json:"quantidade"`
}

type ProdutoInfo struct {
	ID        uint   `json:"id"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

var (
	// ErrIndisponivel é retornado quando o Serviço de Estoque não
	// responde (fora do ar, timeout, rede) — não é erro de negócio.
	ErrIndisponivel         = errors.New("serviço de estoque indisponível")
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado no estoque")
) // os dois erros distingue se estoque não respondeu ou se não encontrou o produto

type ErrSaldoInsuficiente struct {
	ProdutoID  uint
	Produto    string
	Disponivel int
	Solicitado int
}

func (e *ErrSaldoInsuficiente) Error() string {
	return fmt.Sprintf("saldo insuficiente para %q", e.Produto)
}

func (c *Client) BaixarSaldo(itens []ItemBaixa) ([]ProdutoInfo, error) {
	body, err := json.Marshal(map[string]any{"itens": itens})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/produtos/baixar-saldo", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// conexão recusada, timeout, DNS etc → serviço fora do ar
		return nil, ErrIndisponivel
	}
	defer resp.Body.Close()

	switch resp.StatusCode { // escolhe o que retorna com base na resposta http
	case http.StatusOK: // se retornar ok executa
		var out struct {
			Itens []ProdutoInfo `json:"itens"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		return out.Itens, nil // retorna produtos com modelo ProdutoInfo

	case http.StatusConflict: // 409 saldo insuficiente
		var errBody struct {
			ProdutoID  uint   `json:"produtoId"`
			Produto    string `json:"produto"`
			Disponivel int    `json:"disponivel"`
			Solicitado int    `json:"solicitado"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, &ErrSaldoInsuficiente{ // retorna o ProdutoInfo null e o erro
			ProdutoID: errBody.ProdutoID, Produto: errBody.Produto,
			Disponivel: errBody.Disponivel, Solicitado: errBody.Solicitado,
		} // transforma o body conforme o modelo errBody

	case http.StatusBadRequest: // tratando 400 produto não encontrado
		return nil, ErrProdutoNaoEncontrado

	default:
		return nil, fmt.Errorf("estoque respondeu status %d", resp.StatusCode)
	} 
}

// ReporSaldo desfaz uma baixa anterior (compensação de saga).
// Erros aqui são só logados pelo chamador — não faz sentido propagar
// falha de compensação como se fosse falha da operação original.
func (c *Client) ReporSaldo(itens []ItemBaixa) error {
	body, _ := json.Marshal(map[string]any{"itens": itens}) // transforma em json

	// cria a requisição com o metodo e o endpoint
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/produtos/repor-saldo", bytes.NewReader(body)) // bytes transforma em um reader que pode ser usado como corpo de requisição
	if err != nil { // se tiver algum erro retorna o erro
		return err
	}
	req.Header.Set("Content-Type", "application/json") // seta o header da requisição

	resp, err := c.http.Do(req)
	if err != nil { // se ao executar a requisição der erro, retorna o erro padrão
		return ErrIndisponivel
	}
	defer resp.Body.Close() // defer só executa quando terminar a requisição
	return nil // melhoria sempre retorna nil, se caso estoque retornar algum erro após faturamento não fica sabendo
}
func (c *Client) Consultar(ids []uint) ([]ProdutoInfo, error) { // função para busca informação sobre algum produto
	body, _ := json.Marshal(map[string]any{"ids": ids}) // monta o json
	// melhoria: tratar erros do marshal
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/produtos/consultar", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req) // executa a requisição
	if err != nil { // qualquer erro devolve erro padrão
		return nil, ErrIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // se o status for diferente de ok retorna o status
		return nil, fmt.Errorf("estoque respondeu status %d ao consultar produtos", resp.StatusCode)
	}

	var produtos []ProdutoInfo
	if err := json.NewDecoder(resp.Body).Decode(&produtos); err != nil {
		return nil, err
	}
	return produtos, nil  // retorna produtos com modelo ProdutoInfo
}
