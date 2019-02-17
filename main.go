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
	"github.com/PuerkitoBio/goquery"
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
					area, _ := GetAreaQuery(inputText)
					fmt.Println(area.area_query)
					if len(area.area_query) == 0 {

					} else {
						fmt.Println("A")
						query := "https://retty.me/restaurant-search/search-result/?budget_meal_type=2&min_budget=5&max_budget=9&credit_card_use=1&counter_seat=1&" + area.area_query
						fmt.Println(query)

						resp, err := http.Get(query)
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println("B")
						fmt.Printf("%#v", resp.Body)
						fmt.Println("C")
					//	defer resp.Body.Close()

						doc, err := goquery.NewDocumentFromReader(resp.Body)
						fmt.Printf("%#v", doc)
						fmt.Println("D")
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println("E")
						doc.Find(".restaurant").Each(func(_ int, srg *goquery.Selection) {
							fmt.Println("F")
							fmt.Printf("%#v", srg)
							fmt.Println(srg.Text())
							srg.Find(".restaurant__images").Each(func(_ int, s *goquery.Selection) {
								fmt.Println("G")
								fmt.Printf("%#v", srg)
								fmt.Println(srg.Text())
							})
						})
					}
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
