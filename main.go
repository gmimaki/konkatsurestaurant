package main

import (
	"net/http"
	"os"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"strings"
)

type Area struct {
	id int
	area_name string
	area_query string
}

var Db *sql.DB
func init() {
	var err error
	Db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	fmt.Println("接続")
}

func GetAreaQuery(area_name string) (area Area, err error) {
	area = Area{}
	err = Db.QueryRow("SELECT area_query FROM areas WHERE area_name = $1", area_name).Scan(&area.area_query)
	return
}

func sanitizeInput(input string) string {
	input = strings.Replace(input, "駅", "", -1)
	input = strings.Replace(input, "\n", "", -1)
	input = strings.Replace(input, "\r", "", -1)
	input = strings.Replace(input, "\r\n", "", -1)
	input = strings.Replace(input, "\r\n", "", -1)
	input = strings.Replace(input, "・", "", -1)
	input = strings.Replace(input, "/", "", -1)
	input = strings.Replace(input, "*", "", -1)
	input = strings.Replace(input, "'", "", -1)
	input = strings.Replace(input, ";", "", -1)
	input = strings.Replace(input, "<", "", -1)
	input = strings.Replace(input, ">", "", -1)
	input = strings.Replace(input, "=", "", -1)
	input = strings.TrimSpace(input)
	return input
}

func main() {
	client := &http.Client{}
	bot, err := linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_ACCESS_TOKEN"),
		linebot.WithHTTPClient(client),
	)
	if (err != nil) {
		log.Fatal(err)
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			fmt.Println(err)
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		fmt.Println("SUCCESSS")
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					inputText := sanitizeInput(message.Text)
					area_query, _ := GetAreaQuery(inputText)
					fmt.Println(area_query[0])
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(inputText)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
			if event.Type == linebot.EventTypeFollow {
				text := "こんにちは、婚活で使えるお店を提案するBotです\n東京都のエリアを入力すると、そのエリアで婚活に使えそうなお店を提案します"
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})

	if err := http.ListenAndServe(":" + os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
