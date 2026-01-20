package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
	"os"
)

type User struct {
	ID       int
	Username string
	Rating   int
	Rank     int
}

var (
	users []*User
	mu    sync.Mutex
)

func seedUsers(count int) {
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= count; i++ {
		users = append(users, &User{
			ID:       i,
			Username: fmt.Sprintf("user_%d", i),
			Rating:   rand.Intn(4901) + 100,
		})
	}
}

func calculateRanks() {
	sort.Slice(users, func(i, j int) bool {
		return users[i].Rating > users[j].Rating
	})

	rank := 1
	prev := -1
	for i := 0; i < len(users); i++ {
		if users[i].Rating != prev {
			rank = i + 1
		}
		users[i].Rank = rank
		prev = users[i].Rating
	}
}

func startLiveUpdates() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			mu.Lock()

			for i := 0; i < 50; i++ {
				idx := rand.Intn(len(users))
				change := rand.Intn(200) - 100

				users[idx].Rating += change
				if users[idx].Rating < 100 {
					users[idx].Rating = 100
				}
				if users[idx].Rating > 5000 {
					users[idx].Rating = 5000
				}
			}

			calculateRanks()
			mu.Unlock()
		}
	}()
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i := 0; i < 10; i++ {
		u := users[i]
		fmt.Fprintf(w, "Rank: %d | %s | Rating: %d\n", u.Rank, u.Username, u.Rating)
	}
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}

	query := strings.ToLower(r.URL.Query().Get("query"))

	mu.Lock()
	defer mu.Unlock()

	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Username), query) {
			fmt.Fprintf(
				w,
				"Rank: %d | %s | Rating: %d\n",
				u.Rank,
				u.Username,
				u.Rating,
			)
		}
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}


func main() {
	seedUsers(10000)
	calculateRanks()
	startLiveUpdates()

	http.HandleFunc("/leaderboard", leaderboardHandler)
	http.HandleFunc("/search", searchHandler)

	port := "9090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
