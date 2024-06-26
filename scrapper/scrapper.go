package scrapper

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
	"github.com/tashima42/aves-curitiba/database"
)

const defaultScrapperID int64 = 1

type Scrapper struct {
	DB          *sqlx.DB
	AuthCookie  string
	Total       int64
	CurrentPage int64
	PerPage     int64
	WriteToPath string
	HTMLPath    string
}

func (s *Scrapper) Scrape() error {
	if err := s.getData(); err != nil {
		return err
	}
	pages := math.Ceil(float64(s.Total) / float64(s.PerPage))
	pageCounter := 1
	for i := s.CurrentPage + 1; i <= int64(pages); i++ {
		slog.Info("running for page: " + strconv.Itoa(int(i)))
		tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return err
		}
		p, err := s.scrapePage()
		if err != nil {
			log.Fatal("failed to scrape page: " + err.Error())
		}
		if err := s.savePage(tx, p); err != nil {
			return err
		}
		if err := database.SetScrapperCurrentPageByIDTxx(tx, defaultScrapperID, i); err != nil {
			return err
		}
		s.CurrentPage = i
		if err := tx.Commit(); err != nil {
			return err
		}
		var sleepTime time.Duration = 500
		if pageCounter == 100 {
			sleepTime = 60
			pageCounter = 0
		}
		time.Sleep(time.Millisecond * sleepTime)
		pageCounter++
	}
	return nil
}

func (s *Scrapper) ScrapeAdditionalData() error {
	if err := s.getData(); err != nil {
		return err
	}
	scr, err := database.GetScrapperByID(context.Background(), s.DB, 1)
	if err != nil {
		return err
	}
	skip := scr.CurrentPage
	limit := 1

	for skip <= scr.Total {
		tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return err
		}
		slog.Info("getting more registros, skip: " + strconv.Itoa(int(skip)))
		registros, err := database.GetNoLocalRegistrosTxx(tx, limit)
		if err != nil {
			return err
		}
		for _, re := range registros {
			skip += 1
			if err := database.SetScrapperCurrentPageByIDTxx(tx, defaultScrapperID, skip); err != nil {
				return err
			}
			time.Sleep(time.Second * 1)
			slog.Info("running for wa_id: " + strconv.Itoa(int(re.WaID)))
			if err := s.scrapeAdditionalPageData(re.WaID, s.WriteToPath); err != nil {
				return err
			}
			if err := database.SetScrappedTxx(tx, re.ID); err != nil {
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scrapper) ScrapeAutoresHTML() error {
	entries, err := os.ReadDir(s.WriteToPath)
	if err != nil {
		return err
	}
	tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	for _, entry := range entries {
		perfil := strings.TrimSuffix(entry.Name(), ".html")
		autor, err := database.GetAutorInfoByPerfilTxx(tx, perfil)
		if err != nil {
			return err
		}
		log.Println(perfil)
		data, err := getAutorDataFromHTML(filepath.Join(s.WriteToPath, entry.Name()))
		if err != nil {
			return err
		}
		autor.Cidade = data.Cidade
		autor.DataCadastro = data.DataCadastro
		if err := database.CreateAutorTxx(tx, autor); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal("failed to commit tx: " + err.Error())
	}
	return nil
}

func (s *Scrapper) ScrapeAutores() error {
	entries, err := os.ReadDir(s.WriteToPath)
	if err != nil {
		return err
	}
	alreadyScrapped := map[string]bool{}
	for _, entry := range entries {
		alreadyScrapped[strings.TrimSuffix(entry.Name(), ".html")] = true
	}
	tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	autores, err := database.GetAutoresTxx(tx)
	if err != nil {
		return err
	}
	for _, autor := range autores {
		log.Println("testing autor: " + autor)
		if _, ok := alreadyScrapped[autor]; ok {
			log.Println("autor already exists, skipping")
			continue
		}
		log.Println("scrapping")
		if err := s.scrapeAutorPage(autor); err != nil {
			return err
		}
		time.Sleep(time.Second / 2)
	}
	return nil
}

func (s *Scrapper) ScrapeHTML() error {
	const limit = 50
	for {
		slog.Info("getting more registros")
		registros, err := database.GetFilteredRegistros(context.Background(), s.DB, limit)
		if err != nil {
			return err
		}
		log.Print(registros)
		if len(registros) < 1 {
			break
		}
		tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
		if err != nil {
			log.Fatal("failed to start transaction" + err.Error())
		}
		for _, re := range registros {
			slog.Info("running for wa_id: " + strconv.Itoa(int(re.WaID)))
			additional, err := getAdditionalDataFromHTML(filepath.Join(s.HTMLPath, strconv.Itoa(int(re.WaID))+".html"))
			if err != nil {
				log.Fatal("failed to scrape data: " + err.Error())
			}
			re.DataPublicacao = additional.DataPublicacao
			// re.Assunto = additional.Assunto
			// re.Acao = additional.Acao
			// re.Sexo = additional.Sexo
			// re.Idade = additional.Idade
			// re.Observacoes = additional.ObservacoesAutor
			// re.Camera = additional.Camera
			// re.LocalNome = additional.LocalNome
			// re.LocalTipo = additional.LocalTipo
			re.Scrapped = true
			slog.Info("found for wa_id: " + strconv.Itoa(int(re.WaID)))
			if err := database.AddADataPublicacaoInfoTxx(tx, re); err != nil {
				log.Fatal("failed to add additional info " + err.Error())
			}
		}
		if err := tx.Commit(); err != nil {
			log.Fatal("failed to commit tx: " + err.Error())
		}
	}
	return nil
}

func (s *Scrapper) getData() error {
	sc, err := database.GetScrapperByID(context.Background(), s.DB, defaultScrapperID)
	if err != nil {
		return err
	}
	s.Total = sc.Total
	s.CurrentPage = sc.CurrentPage
	s.PerPage = sc.PerPage
	return nil
}

func (s *Scrapper) savePage(tx *sqlx.Tx, p *WikiAvesPage) error {
	slog.Info(fmt.Sprintf("saving page '%d' data", s.CurrentPage))
	for i := 1; i <= int(s.PerPage); i++ {
		item, ok := p.Registros.Itens[strconv.Itoa(i)]
		if !ok {
			slog.Warn("couldn't find item skipping: " + strconv.Itoa(i))
			continue
		}
		var especieID int64
		especieFound := true
		especie, err := database.GetEspecieByWaIDTxx(tx, item.Sp.ID)
		if err != nil {
			slog.Error(err.Error())
			if !strings.Contains(err.Error(), "no rows in result set") {
				return err
			}
			especieFound = false
		}
		if !especieFound {
			slog.Warn("especie not found, creating : " + item.Sp.Nome)
			especieCriadaID, err := database.CreateEspecieTxx(tx, &database.Especie{
				WaID:   item.Sp.ID,
				Nome:   item.Sp.Nome,
				Nvt:    item.Sp.Nvt,
				WikiID: item.Sp.Idwiki,
			})
			if err != nil {
				return err
			}
			especieID = especieCriadaID
		} else {
			especieID = especie.ID
		}
		slog.Info("parsing registro time")
		data, err := time.Parse("02/01/2006", item.Data)
		if err != nil {
			return err
		}
		if err := database.CreateRegistroTxx(tx, &database.Registro{
			WaID:        int64(item.ID),
			Tipo:        item.Tipo,
			UsuarioID:   item.IDUsuario,
			EspecieID:   especieID,
			Autor:       item.Autor,
			Por:         item.Por,
			Perfil:      item.Perfil,
			Data:        data,
			Questionada: item.IsQuestionada,
			Local:       item.Local,
			MunicipioID: int64(item.IDMunicipio),
			Comentarios: int64(item.Coms),
			Likes:       int64(item.Likes),
			Views:       int64(item.Vis),
			Grande:      item.Grande,
			Enviado:     item.Enviado,
			Link:        item.Link,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scrapper) scrapePage() (*WikiAvesPage, error) {
	url := "https://www.wikiaves.com.br/getRegistrosJSON.php?tm=f&t=c&c=4106902&o=dp&desc=0&o=dp&p=" + strconv.Itoa(int(s.CurrentPage))
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", s.AuthCookie)
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("DNT", "1")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://www.wikiaves.com.br/midias.php?tm=f&t=c&c=4106902")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Chromium\";v=\"121\", \"Not A(Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"macOS\"")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	wikiAvesPage := &WikiAvesPage{}

	err = decoder.Decode(wikiAvesPage)
	if err != nil {
		if !strings.Contains(err.Error(), "cannot unmarshal number into Go struct field .registros.itens.sp.id of type string") {
			return nil, err
		} else {
			fixed := strings.Replace(string(bodyBytes), `"id":0`, `"id":"0"`, 1)
			if err = json.Unmarshal([]byte(fixed), wikiAvesPage); err != nil {
				return nil, err
			}
		}
	}
	return wikiAvesPage, nil
}

func (s *Scrapper) scrapeAdditionalPageData(id int64, writeToPath string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.wikiaves.com/_midia_detalhes.php?m="+strconv.Itoa(int(id)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", s.AuthCookie)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.wikiaves.com/midias.php?t=u&u=50609")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Not(A:Brand";v="24", "Chromium";v="122"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(writeToPath, strconv.Itoa(int(id))+".html"), body, 0777)
}

func (s *Scrapper) scrapeAutorPage(id string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.wikiaves.com.br/perfil_"+id, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", s.AuthCookie)
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.wikiaves.com.br/index.php")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not(A:Brand";v="24", "Chromium";v="122"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.WriteToPath, id+".html"), body, 0777)
}

func getAdditionalDataFromHTML(htmlPath string) (*HTMLData, error) {
	body, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	data := HTMLData{}
	doc.Find(".wa-lista-detalhes > div:not(.float-right, .row, #divDetalhesBotao, #divDetalhes)").Each(func(i int, s *goquery.Selection) {
		log.Print(s.Text())
		if strings.Contains(s.Text(), "Publicada") {
			dataPublicacao, err := time.Parse("02/01/2006", strings.TrimPrefix(s.Text(), "Publicada em:\xc2\xa0"))
			if err != nil {
				log.Fatal(err)
			}
			data.DataPublicacao = dataPublicacao
		}
		// if strings.Contains(s.Text(), "Câmera") {
		// 	data.Camera = strings.TrimPrefix(s.Text(), "Câmera:")
		// } else if strings.Contains(s.Text(), "Idade") {
		// 	data.Idade = strings.TrimPrefix(s.Text(), "Idade:")
		// } else if strings.Contains(s.Text(), "Sexo") {
		// 	data.Sexo = strings.TrimPrefix(s.Text(), "Sexo:")
		// } else if strings.Contains(s.Text(), "Assunto") {
		// 	data.Assunto = strings.TrimPrefix(s.Text(), "Assunto(s):")
		// } else if strings.Contains(s.Text(), "Ação") {
		// 	data.Acao = strings.TrimPrefix(s.Text(), "Ação principal:")
		// } else if strings.Contains(s.Text(), "Observações") {
		// 	data.ObservacoesAutor = strings.TrimPrefix(s.Text(), "Observações do autor:")
		// }
	})
	// local := doc.Find(".tipoLocalLOV").Text()
	// if local != "" {
	// 	data.LocalNome = local
	// }
	// localTipo := doc.Find(".tipoLocal").Text()
	// if localTipo != "" {
	// 	data.LocalTipo = localTipo
	// }
	return &data, nil
}

func getAutorDataFromHTML(htmlPath string) (*AutorHTMLData, error) {
	body, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	data := AutorHTMLData{}
	data.Cidade = doc.Find("body > div > div.m-grid__item.m-grid__item--fluid.m-grid.m-grid--ver-desktop.m-grid--desktop.m-body > div > div > div > div > div:nth-child(1) > div > div > div > div > div:nth-child(3) > a.m-card-profile__email.m-link").Text()
	dataCadastro := doc.Find("h8").Text()
	log.Println(dataCadastro)
	dataCadastro = strings.TrimSuffix(strings.TrimPrefix(dataCadastro, "\x0a\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09Data cadastro:  \xc2\xa0"), "\x09\x09\x09\x09\x09\x09\x09\x09\x09")
	dataCadastro = strings.TrimSuffix(strings.TrimPrefix(dataCadastro, "\x0a\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09Data cadastro:  \xc2\xa0"), "\x09")
	log.Println(dataCadastro)
	dataCadastroParsed, err := time.Parse("02/01/2006", dataCadastro)
	if err != nil {
		log.Fatal(err)
	}
	data.DataCadastro = dataCadastroParsed

	return &data, nil
}
