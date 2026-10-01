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
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type MovieResult struct {
	ID    int
	Movie Movie
	Err   error
}

func loadMovie(ctx context.Context, client *http.Client, id int) (Movie, error) {
	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Movie{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return Movie{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Movie{}, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	var movie Movie

	err = json.NewDecoder(resp.Body).Decode(&movie)
	if err != nil {
		return Movie{}, fmt.Errorf("ошибка JSON: %w", err)
	}

	return movie, nil
}

func worker(
	ctx context.Context,
	client *http.Client,
	jobs <-chan int,
	results chan<- MovieResult,
) {
 	for {
  		select {
  		case <-ctx.Done():
   			return

  		case id, ok := <-jobs:
   			if !ok {
    			return
   			}

			movie, err := loadMovie(ctx, client, id)

			select {
			case results <- MovieResult{
				ID:    id,
				Movie: movie,
				Err:   err,
			}:
			case <-ctx.Done():
				return
   			}
  		}
 	}
}

func main() {
	from := flag.Int("from", 0, "ID первого фильма")
	to := flag.Int("to", 0, "ID последнего фильма")
	workers := flag.Int("workers", 10, "Количество воркеров")
	timeout := flag.Duration("timeout", 5*time.Second, "Таймаут HTTP-запроса")

	flag.Parse()

	if *from <= 0 {
		fmt.Fprintln(os.Stderr, "ошибка: --from должен быть больше 0")
		os.Exit(1)
	}

	if *to <= 0 {
		fmt.Fprintln(os.Stderr, "ошибка: --to должен быть больше 0")
		os.Exit(1)
	}

	if *from > *to {
		fmt.Fprintln(os.Stderr, "ошибка: --from не может быть больше --to")
		os.Exit(1)
	}

	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, "ошибка: --workers должен быть больше 0")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

 	client := &http.Client{
 	 	Timeout: *timeout,
 	}

 	jobs := make(chan int)
 	results := make(chan MovieResult)

 	var wg sync.WaitGroup

 	// Запускаем workers
 	for i := 0; i < *workers; i++ {
  		wg.Add(1)
 		go func() {
   			defer wg.Done()
   			worker(ctx, client, jobs, results)
 		 }()
 	}

 	// Отправляем ID фильмов в jobs
 	go func() {
  		defer close(jobs)

 		for id := *from; id <= *to; id++ {
   			select {
  			case jobs <- id:
   			case <-ctx.Done():
    			return
  			 }
 		 }
 	}()

	 // Когда все workers закончат, закрываем results
 	go func() {
  		wg.Wait()
  		close(results)
	}()

 	// Получаем результаты от workers
 	for result := range results {
 		if result.Err != nil {
   			fmt.Printf("Ошибка фильма %d: %v\n", result.ID, result.Err)
   			continue
 		}

  		fmt.Printf(
			"%d — %s — %d — %s\n",
			result.Movie.ID,
			result.Movie.Title,
			result.Movie.Year,
			result.Movie.Director,
  		)
 	}
}



