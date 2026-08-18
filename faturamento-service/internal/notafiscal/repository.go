package notafiscal

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

func (r *Repository) Listar() ([]NotaFiscal, error) {
	var notasFiscais []NotaFiscal

	if err := r.db.
		Preload("Itens").
		Find(&notasFiscais).Error; err != nil {
		return nil, err
	}

	return notasFiscais, nil
}

func (r *Repository) BuscarPorID(id uint) (*NotaFiscal, error) {
	var nota NotaFiscal

	if err := r.db.
		Preload("Itens").
		First(&nota, id).Error; err != nil {
		return nil, err
	}

	return &nota, nil
}
