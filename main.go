package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	postsUrl = "https://jsonplaceholder.typicode.com/posts"
	postCommentsUrl = "https://jsonplaceholder.typicode.com/posts/%d/comments"
)

type Post struct {
	Id     int    `json:"id"`
	UserId int    `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Comment struct {
	Id     int    `json:"id"`
	PostId int    `json:"postId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

type WordCount struct {
	PostId int    `json:"postId"`
	Word   string `json:"word"`
	Count  int    `json:"count"`
}

var (
	db *sql.DB
	err error
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	} else {
		log.Println("Файл .env подключен")
	}
}

func main() {
	db, err = connectToDB()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Успешно подключена БД")
	}

	go startStatisticCheck()

	r := gin.Default()

	r.GET("/post/:id/comments/statistics", getStatisticsHandler)

	log.Println("Успешно запущено на порту :8095")
	log.Fatal(r.Run(":8085"))
}

func connectToDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("DB_STR"))
	if err != nil {
		return nil, err
	}
	
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	
	return db, nil
}

func startStatisticCheck() {
	log.Println("Обновление статистики включено")
	updateStatistics()
	for range time.Tick(5 * time.Minute) {
		updateStatistics()
	}
}

func updateStatistics() {
	var posts []Post

	resp, _ := http.Get(postsUrl)
	body, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &posts)

	for _, post := range posts {
		var comments []Comment
		resWordCountMap := make(map[string]int)

		resp, _ := http.Get(fmt.Sprintf(postCommentsUrl, post.Id))
		body, _ := ioutil.ReadAll(resp.Body)

		json.Unmarshal(body, &comments)

		for _, comment := range comments {
			for _, word := range strings.Fields(comment.Body) {
				resWordCountMap[word]++
			}
		}

		for word, count := range resWordCountMap {
			db.Exec("INSERT INTO comments_statistics (post_id, word, count) VALUES ($1, $2, $3) ON CONFLICT (post_id, word) DO UPDATE SET count = $3", post.Id, word, count)
		}
	}
}

func getStatisticsHandler(c *gin.Context) {
	var wordCounts []WordCount

	postId := c.Param("id")

	rows, _ := db.Query("SELECT * FROM comments_statistics WHERE post_id = $1 ORDER BY count DESC", postId)

	for rows.Next() {
		var wc WordCount
		rows.Scan(&wc.PostId, &wc.Word, &wc.Count)
		wordCounts = append(wordCounts, wc)
	}

	c.JSON(http.StatusOK, wordCounts)
}
