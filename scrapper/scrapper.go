package scrapper

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

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

	for i := s.CurrentPage; i <= int64(pages); i++ {
		if err := database.SetScrapperCurrentPageByID(context.Background(), s.DB, defaultScrapperID, i); err != nil {
			return err
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

func (s *Scrapper) savePage()

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

	decoder := json.NewDecoder(res.Body)
	wikiAvesPage := &WikiAvesPage{}

	err = decoder.Decode(wikiAvesPage)

	return wikiAvesPage, err
}
