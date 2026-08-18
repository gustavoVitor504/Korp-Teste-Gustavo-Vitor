package notafiscal

import (
	"go-nfe-backend-api/internal/estoqueclient"

	"gorm.io/gorm"
)

type NotaFiscalService struct {
	repository    *Repository
	estoqueClient *estoqueclient.Client
}

func NewService(repository *Repository, estoqueClient *estoqueclient.Client) *NotaFiscalService {
	return &NotaFiscalService{
		repository:    repository,
		estoqueClient: estoqueClient,
	}
}

type ItemNota struct {
	ProdutoID  uint
	Quantidade int
}

func (n *NotaFiscalService) Criar(itensReq []ItemNota) (*NotaFiscal, error) {
	if len(itensReq) == 0 {
		return nil, ErrItensObrigatorios
	}
	for _, i := range itensReq {
		if i.Quantidade <= 0 {
			return nil, ErrQuantidadeInvalida
		}
	}

	ids := make([]uint, 0, len(itensReq))
	for _, i := range itensReq {
		ids = append(ids, i.ProdutoID)
	}

	produtosInfo, err := n.estoqueClient.Consultar(ids)
	if err != nil {
		return nil, err
	}

	descricaoPorID := make(map[uint]string, len(produtosInfo))
	for _, p := range produtosInfo {
		descricaoPorID[p.ID] = p.Descricao
	}

	return n.criarLocal(StatusAberta, itensReq, descricaoPorID)
}

func (n *NotaFiscalService) Imprimir(id uint) (*NotaFiscal, error) {
	nota, err := n.repository.BuscarPorID(id)
	if err != nil {
		return nil, err
	}

	if nota.Status != StatusAberta {
		return nil, ErrNotaNaoAberta
	}

	if len(nota.Itens) == 0 {
		return nil, ErrItensObrigatorios
	}

	itensBaixa := make([]estoqueclient.ItemBaixa, 0, len(nota.Itens))

	for _, item := range nota.Itens {
		itensBaixa = append(itensBaixa, estoqueclient.ItemBaixa{
			ProdutoID:  item.ProdutoID,
			Quantidade: item.Quantidade,
		})
	}

	produtosInfo, err := n.estoqueClient.BaixarSaldo(itensBaixa)
	if err != nil {
		return nil, err
	}

	descricaoPorID := make(map[uint]string, len(produtosInfo))

	for _, p := range produtosInfo {
		descricaoPorID[p.ID] = p.Descricao
	}

	var notaFechada NotaFiscal

	db := n.repository.DB()

	err = db.Transaction(func(tx *gorm.DB) error {
		var notaAtual NotaFiscal

		if err := tx.First(&notaAtual, id).Error; err != nil {
			return err
		}

		if notaAtual.Status != StatusAberta {
			return ErrNotaNaoAberta
		}

		if err := tx.Model(&NotaFiscal{}).
			Where("id = ?", id).
			Update("status", StatusFechada).Error; err != nil {
			return err
		}

		if err := tx.Preload("Itens").First(&notaAtual, id).Error; err != nil {
			return err
		}

		notaFechada = notaAtual
		return nil
	})

	if err != nil {

		_ = n.estoqueClient.ReporSaldo(itensBaixa)

		return nil, err
	}

	return &notaFechada, nil
}

func (n *NotaFiscalService) criarLocal(
	status StatusNotaFiscal,
	itensReq []ItemNota,
	descricoes map[uint]string,
) (*NotaFiscal, error) {

	db := n.repository.DB()

	var notaCriada NotaFiscal

	err := db.Transaction(func(tx *gorm.DB) error {

		var ultimaNota NotaFiscal

		numero := uint(1)

		if err := tx.
			Order("numero desc").
			First(&ultimaNota).Error; err == nil {

			numero = ultimaNota.Numero + 1
		}

		nota := NotaFiscal{
			Numero: numero,
			Status: status,
		}

		if err := tx.Create(&nota).Error; err != nil {
			return err
		}

		for _, item := range itensReq {

			descricao := ""

			if descricoes != nil {
				descricao = descricoes[item.ProdutoID]
			}

			notaItem := NotaFiscalItem{
				NotaFiscalID: nota.ID,
				ProdutoID:    item.ProdutoID,
				Descricao:    descricao,
				Quantidade:   item.Quantidade,
			}

			if err := tx.Create(&notaItem).Error; err != nil {
				return err
			}
		}

		if err := tx.
			Preload("Itens").
			First(&nota, nota.ID).Error; err != nil {

			return err
		}

		notaCriada = nota

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &notaCriada, nil
}

func (n *NotaFiscalService) Listar() ([]NotaFiscal, error) {
	return n.repository.Listar()
}
