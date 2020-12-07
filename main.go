package main

import (
	"fmt"
	"github.com/Gumkle/helion_price_scrapper/client"
	"github.com/Gumkle/helion_price_scrapper/datatypes"
	"github.com/Gumkle/helion_price_scrapper/logger"
	"github.com/Gumkle/helion_price_scrapper/sinks"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"sync"
)

func main() {

	logger.InitLogger()
	sinks.InitSinks()
	client.InitClient()
	go initials()
	go distributeConnections()
	go parseContent()
	go saveData()
	var input string
	fmt.Scanln(&input)
}

func initials() {
	sinks.Links <- "https://helion.pl/kategorie/ksiazki"
}

func distributeConnections() {
	for {
		link := <-sinks.Links
		logger.Get().PrintInfo(fmt.Sprintf("Scrappowanie: %s", link))
		go connectToSite(link)
	}
}

func connectToSite(site string) {
	tasklog := logger.Get().NewTask(fmt.Sprintf("Łączenie z %s", site))
	request, err := http.NewRequest("GET", site,nil)
	if err != nil {
		tasklog.SetFailed()
		logger.Get().PrintError("Niepoprawna konfiguracja requesta!")
		return
	}
	request.Header.Set("Accept-Charset", "utf-8")

	response, err := client.Get().Do(request)
	if err != nil {
		tasklog.SetFailed()
		logger.Get().PrintError(fmt.Sprintf("Połączenie nieudane: %s \n %s", site, err))
		return
	}
	tasklog.SetDone()
	sinks.Responses <- response
	return
}

func getResponseDetails(response *http.Response) (datatypes.ResponseDetails, error) {
	url := response.Request.URL.Scheme + "://" + response.Request.URL.Host + response.Request.URL.Path
	dataInBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return datatypes.ResponseDetails{"", "", nil}, err
	}
	return datatypes.ResponseDetails{url, string(dataInBytes), response.Body}, nil
}

func parseContent() {
	for {
		response := <-sinks.Responses
		responseDetails, err := getResponseDetails(response)
		if err != nil {
			logger.Get().PrintError("Nie udało się otworzyć treści strony")
		}
		logger.Get().PrintInfo(fmt.Sprintf("Rozpoczynanie parsowania danych dla %s", responseDetails.Address))
		tasklog := logger.Get().NewTask(fmt.Sprintf("Parsowanie danych z %s", responseDetails.Address))
		linksSuccessful := make(chan bool)
		booksSuccessful := make(chan bool)
		var wg sync.WaitGroup
		wg.Add(2)
		go searchForLinks(response, linksSuccessful, &wg)
		go searchForBookDetails(response, booksSuccessful, &wg)
		wg.Wait()
		if <-linksSuccessful && <-booksSuccessful {
			err = response.Body.Close()
			if err != nil {
				logger.Get().PrintError("Nie udało się zamknąć połączenia ze stroną")
			}
			tasklog.SetDone()
		} else {
			tasklog.SetFailed()
		}
	}
}

func searchForLinks(response *http.Response, done chan bool, s *sync.WaitGroup) {
	defer s.Done()
	z := html.NewTokenizer(response.Body)
	for {
		z.Next()
		fmt.Println(z.Err())
	}
	//if err != nil {
	//	done <- false
	//	logger.Get().PrintError(fmt.Sprintf("%s", err.Error()))
	//	return
	//}
	//if exists {
	//	logger.Get().PrintInfo(fmt.Sprintf("Znaleziony link: %s", href))
	//	sinks.Links <- href
	//}
	done <- true
}

func searchForBookDetails(details *http.Response, done chan bool, s *sync.WaitGroup) {
	defer s.Done()
	done<-true
}

func saveData() {

}
