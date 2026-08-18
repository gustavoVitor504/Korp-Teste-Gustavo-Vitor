package produto

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Criar(produto *Produto) error {
	return r.db.Create(produto).Error
}

func (r *Repository) Listar() ([]Produto, error) {
	var produtos []Produto
	if err := r.db.Find(&produtos).Error; err != nil {
		return nil, err
	}
	return produtos, nil
}

func (r *Repository) BuscarPorIds(ids []uint) ([]Produto, error) {
	var produtos []Produto
	if err := r.db.Find(&produtos, ids).Error; err != nil {
		return nil, err
	}
	return produtos, nil
}
