package main

import (
	"fmt"
	"github.com/Gumkle/consoler/consoler"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Book struct  {
	title string
	price float32
	site string
}

type ResponseDetails struct {
	address string
	content string
	reader io.Reader
}

type Utilities struct {
	logger        *consoler.Logger
	response_sink chan ResponseDetails
	link_sink chan string
	books_sink chan Book
	http_client *http.Client
}

func main() {
	utilities := Utilities{
		consoler.NewLogger(),
		make(chan ResponseDetails, 5),
		make(chan string),
		make(chan Book),
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
	go initials(utilities)
	go distributeConnections(utilities)
	go parseContent(utilities)
	go saveData(utilities)
	var input string
	fmt.Scanln(&input)
}

func initials(utilities Utilities) {
	utilities.link_sink <- "https://helion.pl/kategorie/ksiazki"
}

func distributeConnections(utilities Utilities) {
	for {
		link := <-utilities.link_sink
		utilities.logger.PrintInfo(fmt.Sprintf("Scrappowanie: %s", link))
		go connectToSite(link, utilities)
	}
}

func connectToSite(site string, utilities Utilities) {
	tasklog := utilities.logger.NewTask(fmt.Sprintf("Łączenie z %s", site))
	request, err := http.NewRequest("GET", site,nil)
	if err != nil {
		tasklog.SetFailed()
		utilities.logger.PrintError("Niepoprawna konfiguracja requesta!")
		return
	}
	request.Header.Set("Accept-Charset", "utf-8")

	response, err := utilities.http_client.Do(request)
	if err != nil {
		tasklog.SetFailed()
		utilities.logger.PrintError(fmt.Sprintf("Połączenie nieudane: %s \n %s", site, err))
		return
	}
	responseDetails, err := getResponseDetails(response)
	if err != nil {
		utilities.logger.PrintError("Nie udało się otworzyć treści strony")
	}
	err = response.Body.Close()
	if err != nil {
		utilities.logger.PrintError("Nie udało się zamknąć połączenia ze stroną")
	}
	tasklog.SetDone()
	utilities.response_sink <- responseDetails
	return
}

func getResponseDetails(response *http.Response) (ResponseDetails, error) {
	url := response.Request.URL.Scheme + "://" + response.Request.URL.Host + response.Request.URL.Path
	dataInBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ResponseDetails{"", "", nil}, err
	}
	return ResponseDetails{url, string(dataInBytes), response.Body}, nil
}

func parseContent(utilities Utilities) {
	for {
		response := <-utilities.response_sink
		utilities.logger.PrintInfo(fmt.Sprintf("Rozpoczynanie parsowania danych dla %s", response.address))
		tasklog := utilities.logger.NewTask(fmt.Sprintf("Parsowanie danych z %s", response.address))
		linksDone := make(chan bool)
		booksDone := make(chan bool)
		go searchForLinks(response, utilities, linksDone)
		go searchForBookDetails(response, utilities, booksDone)
		var linksSuccessful, booksSuccessful bool
		select {
			case status := <- linksDone:
				if !status {
					tasklog.SetFailed()
					break
				}
				linksSuccessful = status
			case status := <- booksDone:
				if !status {
					tasklog.SetFailed()
					break
				}
				booksSuccessful = status
		}
		if linksSuccessful && booksSuccessful {
			tasklog.SetDone()
		}
	}
}

func searchForLinks(response ResponseDetails, utilities Utilities, done chan bool) {
	document, err := goquery.NewDocumentFromReader(response.reader)
	if err != nil {
		done <- false
		utilities.logger.PrintError("Parsowanie dokumentu goquery nie udało się")
		return
	}
	document.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists {
			utilities.logger.PrintInfo(fmt.Sprintf("Znaleziony link: %s", href))
			utilities.link_sink <- href
		}
	})
	done <- true
}

func searchForBookDetails(details ResponseDetails, utilities Utilities, done chan bool) {

}

func saveData(utilities Utilities) {

}
