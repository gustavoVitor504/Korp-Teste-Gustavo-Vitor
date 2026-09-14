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
	if produto.Codigo == "" { // tratamento de erros nullable
		return ErrCodigoObrigatorio
	}
	if produto.Descricao == "" {
		return ErrDescricaoObrigatória
	}
	if produto.Saldo < 0 {
		return ErrSaldoInvalido
	}
	return s.repository.Criar(produto) // executa repository para registrar no banco
}

func (s *Service) Listar() ([]Produto, error) {
	return s.repository.Listar()
}        // executa Listar no banco de dados para buscar todos produtos

type ItemQuantidade struct {
	ProdutoID  uint
	Quantidade int
}

func (s *Service) BaixarSaldo(itens []ItemQuantidade) ([]Produto, error) {
	db := s.repository.DB()
	var atualizados []Produto    // instancia um array vazio de Produto

	err := db.Transaction(func(tx *gorm.DB) error { // transaction usado para fazer apenas uma transação para toda função
		for _, item := range itens { // ou todos são executados ou nenhum é alterado
			if item.Quantidade <= 0 {
				return ErrQuantidadeInvalida
			} // primeiro tratamento para não faturar com quantidade zerada

			var p Produto // variável para colocar o produto que o banco encontrou
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}). // usado tx inves de db pois faz parte da transação
				First(&p, item.ProdutoID).Error; err != nil { // locka para transações concorrentes
				return ErrProdutoNaoEncontrado
			}

			if p.Saldo < item.Quantidade {
				return &ErrSaldoInsuficiente{ // verifica saldo suficiente
					ProdutoID: p.ID, Produto: p.Descricao,
					Disponivel: p.Saldo, Solicitado: item.Quantidade,
				}
			}

			p.Saldo -= item.Quantidade  // atualiza o saldo após o faturamento
			if err := tx.Save(&p).Error; err != nil {
				return err
			}
			atualizados = append(atualizados, p) // adiciona na lista de atualizados o produto completo
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return atualizados, nil   // retorna o atualizado
}

// ReporSaldo é a compensação da saga: desfaz uma baixa quando o faturamento falha
func (s *Service) ReporSaldo(itens []ItemQuantidade) error {
	db := s.repository.DB() // faz conexão com o banco
	return db.Transaction(func(tx *gorm.DB) error { // da mesma forma da baixa, o retorno também é por transação
		for _, item := range itens {
			var p Produto
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}). // lock para bloquear requisições concorrentes
				First(&p, item.ProdutoID).Error; err != nil {
				continue  // nesse caso se der erro ignora o produto atual
			}
			p.Saldo += item.Quantidade  // devolve a quantidade
			if err := tx.Save(&p).Error; err != nil {  // salvando o conteudo
				return err
			}
		}
		return nil
	})
}
func (s *Service) ConsultarPorIds(ids []uint) ([]Produto, error) {
	return s.repository.BuscarPorIds(ids)
}
