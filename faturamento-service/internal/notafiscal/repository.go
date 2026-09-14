package notafiscal

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {  // construtor
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {  // devolve o banco que o repository possui
	return r.db
}

func (r *Repository) Listar() ([]NotaFiscal, error) { 
	var notasFiscais []NotaFiscal

	if err := r.db.
		Preload("Itens").
		Find(&notasFiscais).Error; err != nil { // Find busca todas as notas
		return nil, err
	}

	return notasFiscais, nil    // retorna as notas sem erro
}

func (r *Repository) BuscarPorID(id uint) (*NotaFiscal, error) {
	var nota NotaFiscal

	if err := r.db.
		Preload("Itens").
		First(&nota, id).Error; err != nil {  // carrega as nota pelo ID passado
		return nil, err
	}

	return &nota, nil   // retorna a nota
}
