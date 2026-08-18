package produto

type Produto struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Codigo    string `gorm:"unique;not null" json:"codigo"`
	Descricao string `gorm:"unique;not null" json:"descricao"`
	Saldo     int    `json:"saldo"`
}
