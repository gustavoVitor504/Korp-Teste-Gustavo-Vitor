package produto

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Criar(produto *Produto) error {
	if produto.Codigo == "" {
		return ErrCodigoObrigatorio
	}
	if produto.Descricao == "" {
		return ErrDescricaoObrigatória
	}
	if produto.Saldo < 0 {
		return ErrSaldoInvalido
	}
	return s.repository.Criar(produto)
}

func (s *Service) Listar() ([]Produto, error) {
	return s.repository.Listar()
}

type ItemQuantidade struct {
	ProdutoID  uint
	Quantidade int
}

func (s *Service) BaixarSaldo(itens []ItemQuantidade) ([]Produto, error) {
	db := s.repository.DB()
	var atualizados []Produto

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range itens {
			if item.Quantidade <= 0 {
				return ErrQuantidadeInvalida
			}

			var p Produto
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&p, item.ProdutoID).Error; err != nil {
				return ErrProdutoNaoEncontrado
			}

			if p.Saldo < item.Quantidade {
				return &ErrSaldoInsuficiente{
					ProdutoID: p.ID, Produto: p.Descricao,
					Disponivel: p.Saldo, Solicitado: item.Quantidade,
				}
			}

			p.Saldo -= item.Quantidade
			if err := tx.Save(&p).Error; err != nil {
				return err
			}
			atualizados = append(atualizados, p)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return atualizados, nil
}

// ReporSaldo é a compensação da saga: desfaz uma baixa quando o faturamento falha
func (s *Service) ReporSaldo(itens []ItemQuantidade) error {
	db := s.repository.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range itens {
			var p Produto
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&p, item.ProdutoID).Error; err != nil {
				continue
			}
			p.Saldo += item.Quantidade
			if err := tx.Save(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (s *Service) ConsultarPorIds(ids []uint) ([]Produto, error) {
	return s.repository.BuscarPorIds(ids)
}
