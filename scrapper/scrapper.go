package scrapper

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
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
		var sleepTime time.Duration = 1
		if pageCounter == 100 {
			sleepTime = 60
			pageCounter = 0
		}
		time.Sleep(time.Second * sleepTime)
		pageCounter++
	}
	return nil
}

func (s *Scrapper) ScrapeAdditionalData() error {
	_, err := s.scrapeAdditionalPageData()
	return err
	// if err := s.getData(); err != nil {
	// 	return err
	// }
	// pages := math.Ceil(float64(s.Total) / float64(s.PerPage))
	// pageCounter := 1
	// for i := s.CurrentPage + 1; i <= int64(pages); i++ {
	// 	slog.Info("running for page: " + strconv.Itoa(int(i)))
	// 	tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	p, err := s.scrapePage()
	// 	if err != nil {
	// 		log.Fatal("failed to scrape page: " + err.Error())
	// 	}
	// 	if err := s.savePage(tx, p); err != nil {
	// 		return err
	// 	}
	// 	if err := database.SetScrapperCurrentPageByIDTxx(tx, defaultScrapperID, i); err != nil {
	// 		return err
	// 	}
	// 	s.CurrentPage = i
	// 	if err := tx.Commit(); err != nil {
	// 		return err
	// 	}
	// 	var sleepTime time.Duration = 1
	// 	if pageCounter == 100 {
	// 		sleepTime = 60
	// 		pageCounter = 0
	// 	}
	// 	time.Sleep(time.Second * sleepTime)
	// 	pageCounter++
	// }
	// return nil
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

func (s *Scrapper) scrapeAdditionalPageData() (*WikiAvesAdditionalData, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.wikiaves.com/_midia_detalhes.php?m=5920941", nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()
	return getAdditionalDataFromHTML(resp.Body)
}

func getAdditionalDataFromHTML(body io.ReadCloser) (*WikiAvesAdditionalData, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	fmt.Println(doc.Find(".tipoLocalLOV").Text())
	fmt.Println(doc.Find(".tipoLocal").Text())
	return nil, nil
}

func (s *Scrapper) CSVAdditionalData(fileLocation string) error {
	f, err := os.ReadFile(fileLocation)
	if err != nil {
		return err
	}
	additionalDatas := []AdditionalData{}
	r := csv.NewReader(bytes.NewBuffer(f))
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if record[4] == "Curitiba" {
			continue
		}
		ad := AdditionalData{
			Nome:      strings.TrimSpace(record[0]),
			Especie:   strings.TrimSpace(record[1]),
			Data:      record[2],
			Publicada: record[3],
			Local:     record[4],
			Autor:     strings.TrimSpace(record[5]),
		}
		additionalDatas = append(additionalDatas, ad)
	}

	additionalMap := map[string]string{}
	for _, a := range additionalDatas {
		key := fmt.Sprintf("%s-%s-%s", a.Nome, a.Data, strings.ToLower(strings.ReplaceAll(a.Autor, " ", "")))
		slog.Info(key)
		additionalMap[key] = a.Local
	}

	tx, err := s.DB.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	registros, err := database.GetFilteredRegistrosTxx(tx)
	if err != nil {
		return err
	}
	for _, r := range registros {
		registro := *r
		key := fmt.Sprintf("%s-%s-%s", registro.Especie, strings.TrimSuffix(registro.Data, "T00:00:00Z"), strings.ToLower(strings.ReplaceAll(registro.Autor, " ", "")))
		// slog.Info(key)
		local, ok := additionalMap[key]
		if !ok {
			continue
		}
		slog.Info("FOUND!: " + key)
		r.LocalNome = local
		database.UpdateLocalTxx(tx, r)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
