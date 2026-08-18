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

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
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
)

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

	switch resp.StatusCode {
	case http.StatusOK:
		var out struct {
			Itens []ProdutoInfo `json:"itens"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		return out.Itens, nil

	case http.StatusConflict:
		var errBody struct {
			ProdutoID  uint   `json:"produtoId"`
			Produto    string `json:"produto"`
			Disponivel int    `json:"disponivel"`
			Solicitado int    `json:"solicitado"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, &ErrSaldoInsuficiente{
			ProdutoID: errBody.ProdutoID, Produto: errBody.Produto,
			Disponivel: errBody.Disponivel, Solicitado: errBody.Solicitado,
		}

	case http.StatusBadRequest:
		return nil, ErrProdutoNaoEncontrado

	default:
		return nil, fmt.Errorf("estoque respondeu status %d", resp.StatusCode)
	}
}

// ReporSaldo desfaz uma baixa anterior (compensação de saga).
// Erros aqui são só logados pelo chamador — não faz sentido propagar
// falha de compensação como se fosse falha da operação original.
func (c *Client) ReporSaldo(itens []ItemBaixa) error {
	body, _ := json.Marshal(map[string]any{"itens": itens})

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/produtos/repor-saldo", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return ErrIndisponivel
	}
	defer resp.Body.Close()
	return nil
}
func (c *Client) Consultar(ids []uint) ([]ProdutoInfo, error) {
	body, _ := json.Marshal(map[string]any{"ids": ids})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/produtos/consultar", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ErrIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("estoque respondeu status %d ao consultar produtos", resp.StatusCode)
	}

	var produtos []ProdutoInfo
	if err := json.NewDecoder(resp.Body).Decode(&produtos); err != nil {
		return nil, err
	}
	return produtos, nil
}
