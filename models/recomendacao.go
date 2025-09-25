package models

type Recomendacao struct {
	ID        int    `db:"id" json:"id"`
	Titulo    string `db:"titulo" json:"titulo"`
	Descricao string `db:"descricao" json:"descricao"`
}
