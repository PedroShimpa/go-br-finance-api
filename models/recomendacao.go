package models

type Recomendacao struct {
	ID        int    `db:"id" json:"id"`
	Titulo    string `db:"titulo" json:"titulo"`
	Descricao string `db:"descricao" json:"descricao"`
	CreatedAt string `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt string `db:"updated_at" json:"updated_at,omitempty"`
}
