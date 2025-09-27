package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

type Book struct {
	ID           string   `json:"id"`            // айди
	Title        string   `json:"title"`         // Название
	Authors      []string `json:"authors"`       // Авторы
	Description  string   `json:"description"`   // Описание
	Rating       float64  `json:"averageRating"` // Средний рейтинг
	RatingsCount int      `json:"ratingsCount"`  // Количесвто оценок
	Categories   []string `json:"categories"`    // Категория

	ImageLinks struct {
		ImageLinksSmall string `json:"smallThumbnail"` // Ссылка на уменьшенную обложку
	} `json:"imageLinks"`
}

// Информация о книге с АПИ
type BookResponse struct {
	Items []struct {
		VolumeInfo Book   `json:"volumeInfo"`
		ID         string `json:"id"`
	} `json:"items"`
}

func main() {
	fmt.Println("Кнжный рекомендатель")
	fmt.Println("====================")

	// Реализуем бесконечную работу
	for {
		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Поиск книг по жанрам")
		fmt.Println("2. Выход")

		fmt.Print("Ваш выбор: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			searchByGenre()
		case "2":
			fmt.Println("До свидания!")
			os.Exit(0)
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

// Функция поиска рекомендации
func searchByGenre() {
	var genre string
	fmt.Print("Введите нужный вам жанр: ")
	fmt.Scanln(&genre)

	books, err := fetchBooksByGenre(genre)
	if err != nil {
		fmt.Printf("Ошибка получения книг: %v\n", err)
		return
	}
	if len(books) == 0 {
		fmt.Println("Книги данного жанра не найдены")
	}

	sort.Slice(books, func(i, j int) bool {
		return books[i].RatingsCount > books[j].RatingsCount
	})

	books = removeDuplicateBooks(books)
	printBooks(genre, books)

}

func fetchBooksByGenre(genre string) ([]Book, error) {
	baseUrl := "https://www.googleapis.com/books/v1/volumes"

	params := url.Values{}
	params.Add("q", "subject:"+genre)
	params.Add("printType", "BOOKS")
	params.Add("key", "AIzaSyBDYonGaktsEUQyPMr935NOuHTFjH1bDek")

	fullUrl := baseUrl + "?" + params.Encode()
	fmt.Println("Отправляется запрос по адресу:", fullUrl)

	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("статус ошибки: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var response BookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	var books []Book
	for _, item := range response.Items {
		book := item.VolumeInfo
		book.ID = item.ID
		books = append(books, book)
	}

	return books, nil
}

func printBooks(genre string, books []Book) {
	books = removeDuplicateBooks(books)

	fmt.Printf("\nРекомендуемые книги в жанре '%s':\n", genre)
	fmt.Println("====================================")

	mx := 5
	if len(books) < mx {
		mx = len(books)
	}

	for i := 0; i < mx; i++ {
		book := books[i]
		fmt.Printf("\n%d. %s\n", i+1, book.Title)
		fmt.Println("   Автор(ы):", resAu(book.Authors))

		if book.RatingsCount > 0 { // Не показываем если нет рейтинга
			fmt.Printf("   Рейтинг: %.1f (%d отзывов)\n", book.Rating, book.RatingsCount)
		}

		if book.ImageLinks.ImageLinksSmall != "" {
			fmt.Println("   Обложка:", book.ImageLinks.ImageLinksSmall)
		}

		desc := shortenDescription(book.Description, 200)
		if desc != "" {
			fmt.Println("   Описание:", desc)
		}
		fmt.Println("------------------------------------")
	}
}

func removeDuplicateBooks(books []Book) []Book {
	unique := make(map[string]bool)
	result := []Book{}

	for _, book := range books {
		if !unique[book.ID] {
			unique[book.ID] = true
			result = append(result, book)
		}
	}
	return result
}

func resAu(authors []string) string {
	if len(authors) == 0 {
		return "Неизвестный автор"
	}

	res := ""
	for i, auth := range authors {
		if i > 0 {
			res += ", "
		}
		res += auth
	}
	return res
}

func shortenDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen] + "..."
}
