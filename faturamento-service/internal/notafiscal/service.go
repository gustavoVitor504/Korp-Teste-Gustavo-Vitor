package notafiscal

import (
	"go-nfe-backend-api/internal/estoqueclient"

	"gorm.io/gorm"
)

type NotaFiscalService struct { // recebe duas dependências no model
	repository    *Repository
	estoqueClient *estoqueclient.Client
}

func NewService(repository *Repository, estoqueClient *estoqueclient.Client) *NotaFiscalService {
	return &NotaFiscalService{ // construtor com as duas dependências
		repository:    repository,
		estoqueClient: estoqueClient,
	}
}

type ItemNota struct {
	ProdutoID  uint
	Quantidade int
}

func (n *NotaFiscalService) Criar(itensReq []ItemNota) (*NotaFiscal, error) {
	if len(itensReq) == 0 { // se não colocar itens na nota
		return nil, ErrItensObrigatorios
	}
	for _, i := range itensReq { // se a quantidade for menor que 1
		if i.Quantidade <= 0 {
			return nil, ErrQuantidadeInvalida
		}
	}
    
	ids := make([]uint, 0, len(itensReq)) // transforma os itensReq em ids
	for _, i := range itensReq { // percorre todos e adiciona em ids
		ids = append(ids, i.ProdutoID)
	}

	produtosInfo, err := n.estoqueClient.Consultar(ids) // recebe os produtosInfo dos ids requisitados
	if err != nil {
		return nil, err
	}

	descricaoPorID := make(map[uint]string, len(produtosInfo)) // mapeia os ids para pegar as desc
	for _, p := range produtosInfo { // para cada p.ID recebe uma p.Descricao
		descricaoPorID[p.ID] = p.Descricao
	}

	return n.criarLocal(StatusAberta, itensReq, descricaoPorID) 
}

func (n *NotaFiscalService) Imprimir(id uint) (*NotaFiscal, error) {
	nota, err := n.repository.BuscarPorID(id) // busca a nota por ID
	if err != nil { 
		return nil, err
	}

	if nota.Status != StatusAberta { // se não estiver aberta a nota
		return nil, ErrNotaNaoAberta
	}

	if len(nota.Itens) == 0 { // se não tiver itens
		return nil, ErrItensObrigatorios
	}
	// transforma itens em ItensBaixa
	itensBaixa := make([]estoqueclient.ItemBaixa, 0, len(nota.Itens))

	for _, item := range nota.Itens { // percorre todos itens e adiciona a itensBaixa
		itensBaixa = append(itensBaixa, estoqueclient.ItemBaixa{
			ProdutoID:  item.ProdutoID,
			Quantidade: item.Quantidade,
		})
	}

	produtosInfo, err := n.estoqueClient.BaixarSaldo(itensBaixa) // recebe o produtosInfo após executar o baixar saldo
	if err != nil {
		return nil, err
	}

	descricaoPorID := make(map[uint]string, len(produtosInfo)) // coloca descrição em cada produto

	for _, p := range produtosInfo {
		descricaoPorID[p.ID] = p.Descricao
	}

	var notaFechada NotaFiscal

	db := n.repository.DB() // recebe o banco

	err = db.Transaction(func(tx *gorm.DB) error { // começa a transação
		var notaAtual NotaFiscal

		if err := tx.First(&notaAtual, id).Error; err != nil { // busca nota novamente
			return err
		}

		if notaAtual.Status != StatusAberta { // verifica status novamente
			return ErrNotaNaoAberta
		}

		if err := tx.Model(&NotaFiscal{}).
			Where("id = ?", id). // atualiza status para fechada e ? é preenchido pelo GORM
			Update("status", StatusFechada).Error; err != nil {
			return err
		}

		if err := tx.Preload("Itens").First(&notaAtual, id).Error; err != nil { // busca nota novamente com preload itens
			return err
		}

		notaFechada = notaAtual  // variavel externa recebe a nota fechada
		return nil
	})

	if err != nil {  // se caso não conseguir completar a transação ele chama a reporSaldo

		_ = n.estoqueClient.ReporSaldo(itensBaixa)

		return nil, err
	}

	return &notaFechada, nil    // retorna nota fechada e sem erro
}

func (n *NotaFiscalService) criarLocal(
	status StatusNotaFiscal, // recebe o status, itens e as descrições
	itensReq []ItemNota,
	descricoes map[uint]string,
) (*NotaFiscal, error) {

	db := n.repository.DB() // recebe o banco para manipular a transação

	var notaCriada NotaFiscal

	err := db.Transaction(func(tx *gorm.DB) error {

		var ultimaNota NotaFiscal

		numero := uint(1)

		if err := tx.
			Order("numero desc").
			First(&ultimaNota).Error; err == nil { // descobre a ultima nota feita

			numero = ultimaNota.Numero + 1 // cria com o prox número
		} // possível melhora mudar a geração do número de nota, pois é possível ter concorrência

		nota := NotaFiscal{
			Numero: numero,
			Status: status,
		}

		if err := tx.Create(&nota).Error; err != nil { // persistência no banco
			return err
		}

		for _, item := range itensReq { // percorre os produtos recebidos

			descricao := ""

			if descricoes != nil {
				descricao = descricoes[item.ProdutoID] // procura a desc por ID
			}

			notaItem := NotaFiscalItem{
				NotaFiscalID: nota.ID,
				ProdutoID:    item.ProdutoID,
				Descricao:    descricao,
				Quantidade:   item.Quantidade,
			}

			if err := tx.Create(&notaItem).Error; err != nil { // persiste o item da nota
				return err
			}
		}

		if err := tx.
			Preload("Itens").
			First(&nota, nota.ID).Error; err != nil { // recarrega a nota com os itens

			return err
		}

		notaCriada = nota

		return nil
	})

	if err != nil {  // trata o erro da transação
		return nil, err
	}

	return &notaCriada, nil // retorna a nota	
}

func (n *NotaFiscalService) Listar() ([]NotaFiscal, error) {  // chama a função de listar no repository
	return n.repository.Listar()
}
