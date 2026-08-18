package notafiscal

import "errors"

var (
	ErrItensObrigatorios  = errors.New("a nota fiscal precisa de ao menos um item")
	ErrQuantidadeInvalida = errors.New("quantidade deve ser maior que zero")
	ErrNotaNaoAberta      = errors.New("a nota fiscal não está aberta")
)
