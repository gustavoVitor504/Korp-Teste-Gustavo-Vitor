package notafiscal

import "fmt"

type StatusNotaFiscal string

const (
	StatusAberta    StatusNotaFiscal = "aberta"
	StatusFechada   StatusNotaFiscal = "fechada"
	StatusCancelada StatusNotaFiscal = "cancelada"
)

type NotaFiscal struct {
	ID     uint             `gorm:"primaryKey" json:"id"`
	Numero uint             `gorm:"unique;not null" json:"numero"`
	Status StatusNotaFiscal `gorm:"not null" json:"status"`
	Itens  []NotaFiscalItem `gorm:"foreignKey:NotaFiscalID" json:"itens"`
}

func (NotaFiscal) TableName() string {
	return "nota_fiscais"
}

func (n NotaFiscal) NumeroFormatado() string {
	return fmt.Sprintf("%09d", n.Numero)
}

type NotaFiscalItem struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	NotaFiscalID uint   `gorm:"not null" json:"notaFiscalId"`
	ProdutoID    uint   `gorm:"not null" json:"produtoId"`
	Descricao    string `gorm:"not null" json:"descricao"`
	Quantidade   int    `gorm:"not null" json:"quantidade"`
}
