package scrapper

import "time"

type WikiAvesItem struct {
	ID        int    `json:"id"`
	Tipo      string `json:"tipo"`
	IDUsuario string `json:"id_usuario"`
	Sp        struct {
		ID     string `json:"id"`
		Nome   string `json:"nome"`
		Nvt    string `json:"nvt"`
		Idwiki string `json:"idwiki"`
	} `json:"sp"`
	Autor         string `json:"autor"`
	Por           string `json:"por"`
	Perfil        string `json:"perfil"`
	Data          string `json:"data"`
	IsQuestionada bool   `json:"is_questionada"`
	Local         string `json:"local"`
	IDMunicipio   int    `json:"idMunicipio"`
	Coms          int    `json:"coms"`
	Likes         int    `json:"likes"`
	Vis           int    `json:"vis"`
	Grande        string `json:"grande"`
	Enviado       string `json:"enviado"`
	Link          string `json:"link"`
}

type WikiAvesPage struct {
	Registros struct {
		Titulo string                  `json:"titulo,omitempty"`
		Link   string                  `json:"link,omitempty"`
		Total  string                  `json:"total,omitempty"`
		Itens  map[string]WikiAvesItem `json:"itens,omitempty"`
	} `json:"registros,omitempty"`
}

type AdditionalData struct {
	Nome      string
	Especie   string
	Data      string
	Publicada string
	Local     string
	Autor     string
}

type HTMLData struct {
	Assunto          string
	Acao             string
	Sexo             string
	Idade            string
	ObservacoesAutor string
	Camera           string
	DataPublicacao   time.Time
	LocalNome        string
	LocalTipo        string
}
