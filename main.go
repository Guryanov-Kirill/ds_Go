package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Movie struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Year any `json:"year"`
	Director string `json:"director"`
}

func fetchMovie(ctx context.Context, client *http.Client, id int, timeout time.Duration, mu *sync.Mutex) {
	reqCtx, cansel := context.WithTimeout(ctx, timeout)
	defer cansel()

	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)
	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		printError(mu, id, fmt.Errorf("не удалось создать запрос: %w", err))
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		printError(mu, id, fmt.Errorf("ошибка сети/таймаут: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		printError(mu, id, fmt.Errorf("сервер вернул статус %d", resp.StatusCode))
		return
	}
	var movie Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		printError(mu, id, fmt.Errorf("ошибка парсинга битого JSON: %v", err))
		return
	}
	mu.Lock()
	fmt.Printf("%d - %s - %v - %s\n", movie.ID, movie.Title, movie.Year, movie.Director)
	mu.Unlock()
}

func printError(mu *sync.Mutex, id int, err error) {
	mu.Lock()
	fmt.Fprintf(os.Stderr, "[Фильм ID %d] Ошибка: %v\n", id, err)
	mu.Unlock()
}

func main() {
	fromFlag := flag.Int("from", -1, "ID первого фильма")
	toFlag := flag.Int("to", -1, "ID последнего фильма")
	workersFlag := flag.Int("workers", 10, "Количество параллельных воркеров")
	timeoutFlag := flag.Int("timeout", 5, "Таймаут одного HTTP-запроса в секундах")

	flag.Parse()

	if *fromFlag == -1 || *toFlag == -1 {
		fmt.Fprintln(os.Stderr, "нужно заполнить флаги --from и --to")
		flag.Usage()
		os.Exit(1)
	}

	if *fromFlag > *toFlag {
		fmt.Fprintf(os.Stderr, "значение --from (%d) не может быть больше --to (%d).\n", *fromFlag, *toFlag)
		os.Exit(1)
	}

	if *workersFlag <= 0 {
		fmt.Fprintf(os.Stderr, "кол-во воркеров должно быть больше 0 (указано: %d).\n", *workersFlag)
		os.Exit(1)
	}

	if *timeoutFlag <= 0 {
		fmt.Fprintf(os.Stderr, "время должно быть больше 0 секунд")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	totalJobs := *toFlag - *fromFlag + 1
	if *workersFlag > totalJobs {
		*workersFlag = totalJobs
	}

	jobs := make(chan int, totalJobs)
	var wg sync.WaitGroup
	var mu sync.Mutex

	client := &http.Client{}

	for w := 1; w <= *workersFlag; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				if ctx.Err() != nil {
					return
				}
				fetchMovie(ctx, client, id, time.Duration(*timeoutFlag)*time.Second, &mu)
			}
		}()
	}

	for id := *fromFlag; id <= *toFlag; id++ {
		jobs <- id
	}
	close(jobs)
	
	wg.Wait()

	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "\nПрограмма была прервана пользователем (Ctrl+C). Оставшиеся запросы отменены.")
		os.Exit(1)
	}
}