package produto

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB              // objeto do gorm que representa a conexão com o banco
}

func NewRepository(db *gorm.DB) *Repository { // construtor 
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { // devolve o banco que o repository possui
	return r.db
}

func (r *Repository) Criar(produto *Produto) error { // persiste novo produto no banco
	return r.db.Create(produto).Error      // .Error para se caso ocorra um erro devolver o erro
}

func (r *Repository) Listar() ([]Produto, error) { // retornar todos registros
	var produtos []Produto    // inicia a slice vazia
	if err := r.db.Find(&produtos).Error; err != nil { // o Find busca todos registros sem condição
		return nil, err     // retorna nulo no lugar da slice e o erro
	}
	return produtos, nil      // retorna minha slice e o erro nulo
}

func (r *Repository) BuscarPorIds(ids []uint) ([]Produto, error) {
	var produtos []Produto
	if err := r.db.Find(&produtos, ids).Error; err != nil {
		return nil, err  // nessa Find busca todos registros com as condições impostas
	}
	return produtos, nil   
}
