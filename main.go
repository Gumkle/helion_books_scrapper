package main

import (
	"fmt"
	"github.com/Gumkle/consoler/consoler"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
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
	response_sink chan *http.Response
	link_sink chan string
	books_sink chan Book
	http_client *http.Client
}

func main() {
	utilities := Utilities{
		consoler.NewLogger(),
		make(chan *http.Response, 5),
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
	tasklog.SetDone()
	utilities.response_sink <- response
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
		responseDetails, err := getResponseDetails(response)
		if err != nil {
			utilities.logger.PrintError("Nie udało się otworzyć treści strony")
		}
		utilities.logger.PrintInfo(fmt.Sprintf("Rozpoczynanie parsowania danych dla %s", responseDetails.address))
		tasklog := utilities.logger.NewTask(fmt.Sprintf("Parsowanie danych z %s", responseDetails.address))
		linksSuccessful := make(chan bool)
		booksSuccessful := make(chan bool)
		var wg sync.WaitGroup
		wg.Add(2)
		go searchForLinks(response, utilities, linksSuccessful, &wg)
		go searchForBookDetails(response, utilities, booksSuccessful, &wg)
		wg.Wait()
		if <-linksSuccessful && <-booksSuccessful {
			err = response.Body.Close()
			if err != nil {
				utilities.logger.PrintError("Nie udało się zamknąć połączenia ze stroną")
			}
			tasklog.SetDone()
		} else {
			tasklog.SetFailed()
		}
	}
}

func searchForLinks(response *http.Response, utilities Utilities, done chan bool, s *sync.WaitGroup) {
	defer s.Done()
	z := html.NewTokenizer(response.Body)
	for {
		z.Next()
		fmt.Println(z.Err())
	}
	//if err != nil {
	//	done <- false
	//	utilities.logger.PrintError(fmt.Sprintf("%s", err.Error()))
	//	return
	//}
	//if exists {
	//	utilities.logger.PrintInfo(fmt.Sprintf("Znaleziony link: %s", href))
	//	utilities.link_sink <- href
	//}
	done <- true
}

func searchForBookDetails(details *http.Response, utilities Utilities, done chan bool, s *sync.WaitGroup) {
	defer s.Done()
	done<-true
}

func saveData(utilities Utilities) {

}
