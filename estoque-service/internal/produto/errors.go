package produto

import (
	"errors"
	"fmt"
)

var (   // tipagem dos erros
	ErrDescricaoObrigatória = errors.New("descrição do produto é obrigatória")
	ErrCodigoObrigatorio    = errors.New("código do produto é obrigatório")
	ErrSaldoInvalido        = errors.New("saldo não pode ser negativo")
	ErrQuantidadeInvalida   = errors.New("quantidade deve ser maior que zero")
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
)

type ErrSaldoInsuficiente struct {  // model para saldo insuficiente
	ProdutoID  uint
	Produto    string
	Disponivel int
	Solicitado int
}

func (e *ErrSaldoInsuficiente) Error() string { // erro padrão para saldo insuficiente
	return fmt.Sprintf("saldo insuficiente para %q: disponível %d, solicitado %d", e.Produto, e.Disponivel, e.Solicitado)
}
