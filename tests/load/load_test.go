package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

const link = "http://localhost:8080"

type params struct {
	total     int
	success   int
	failed    int
	latencies []time.Duration
	mu        sync.Mutex
}

func (s *params) add(latency time.Duration, ok bool) {
	s.mu.Lock()
	s.total++
	if ok {
		s.success++
	} else {
		s.failed++
	}
	s.latencies = append(s.latencies, latency)
	s.mu.Unlock()
}

func (s *params) print() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.total == 0 {
		fmt.Println("Нет запросов")
		return
	}
	var sum time.Duration
	var max time.Duration
	min := time.Hour
	for _, lat := range s.latencies {
		sum += lat
		if lat > max {
			max = lat
		}
		if lat < min {
			min = lat
		}
	}

	avg := sum / time.Duration(s.total)
	successRate := float64(s.success) / float64(s.total) * 100
	fmt.Printf("Всего: %d, Успешно: %d (%.1f%%), Ошибок: %d\n", s.total, s.success, successRate, s.failed)
	fmt.Printf("Задержка: мин=%v, средняя=%v, макс=%v\n", min, avg, max)
}

// функция для запроса
func doRequest(method, url string, body interface{}) (int, time.Duration, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, 0, err
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, 0, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return 0, latency, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, latency, nil
}

// Тест создания команд
func TestLoadTeamAdd(t *testing.T) {
	params := &params{}
	var wg sync.WaitGroup

	// 3 горутины, каждая делает по 15 запросов = 10 RPS
	workers := 3
	requests := 15

	fmt.Println("Тест POST /team/add")

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < requests; j++ {
				teamName := fmt.Sprintf("team_%d_%d", id, j)
				members := []map[string]interface{}{
					{"user_id": fmt.Sprintf("u_%d_%d_1", id, j), "username": "User1", "is_active": true},
					{"user_id": fmt.Sprintf("u_%d_%d_2", id, j), "username": "User2", "is_active": true},
				}

				body := map[string]interface{}{
					"team_name": teamName,
					"members":   members,
				}

				code, latency, err := doRequest("POST", link+"/team/add", body)
				ok := err == nil && (code == 200 || code == 409)
				params.add(latency, ok)

				time.Sleep(200 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	params.print()
}

// Тест получения команд
func TestLoadTeamGet(t *testing.T) {
	for i := 0; i < 5; i++ {
		teamName := fmt.Sprintf("test_team_%d", i)
		members := []map[string]interface{}{
			{"user_id": fmt.Sprintf("u_%d_1", i), "username": "User1", "is_active": true},
			{"user_id": fmt.Sprintf("u_%d_2", i), "username": "User2", "is_active": true},
		}
		body := map[string]interface{}{
			"team_name": teamName,
			"members":   members,
		}
		doRequest("POST", link+"/team/add", body)
		time.Sleep(50 * time.Millisecond)
	}

	params := &params{}
	var wg sync.WaitGroup
	workers := 3
	requests := 15
	fmt.Println("Тест GET /team/get")

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requests; j++ {
				teamName := fmt.Sprintf("test_team_%d", j%5)
				code, latency, err := doRequest("GET", link+"/team/get?team_name="+teamName, nil)
				ok := err == nil && (code == 200 || code == 404)
				params.add(latency, ok)

				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	params.print()
}

// Тест создания PR
func TestLoadCreatePR(t *testing.T) {
	for i := 0; i < 3; i++ {
		teamName := fmt.Sprintf("pr_team_%d", i)
		members := []map[string]interface{}{
			{"user_id": fmt.Sprintf("pr_u_%d_1", i), "username": "User1", "is_active": true},
			{"user_id": fmt.Sprintf("pr_u_%d_2", i), "username": "User2", "is_active": true},
		}
		body := map[string]interface{}{
			"team_name": teamName,
			"members":   members,
		}
		doRequest("POST", link+"/team/add", body)
		time.Sleep(50 * time.Millisecond)
	}

	params := &params{}
	var wg sync.WaitGroup
	workers := 3
	requests := 15
	fmt.Println("Тест POST /pullRequest/create")

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < requests; j++ {
				prID := fmt.Sprintf("pr_%d_%d", id, j)
				body := map[string]interface{}{
					"pull_request_id":   prID,
					"pull_request_name": fmt.Sprintf("PR %d", j),
					"author_id":         fmt.Sprintf("pr_u_%d_1", j%3),
				}
				code, latency, err := doRequest("POST", link+"/pullRequest/create", body)
				ok := err == nil && (code == 200 || code == 404 || code == 409)
				params.add(latency, ok)
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	params.print()
}
