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
	"time"
	"math/rand"
)

type Area struct {
	id int
	area_name string
	area_query string
}

type Restaurant struct {
	name string
	url string
	description string
	image_url string
}

type Restaurants []Restaurant

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
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					inputText := sanitizeInput(message.Text)
					area, _ := GetAreaQuery(inputText)
					if len(area.area_query) == 0 {

					} else {
						query := "https://www.ozmall.co.jp/restaurant/tokyo/" + area.area_query
						fmt.Println(query)

						resp, err := http.Get(query)
						if err != nil {
							fmt.Println(err)
						}
						defer resp.Body.Close()

						doc, err := goquery.NewDocumentFromReader(resp.Body)
						if err != nil {
							fmt.Println(err)
						}

						var restaurants Restaurants
						doc.Find(".ozDinIchiWrp").Each(func(_ int, ozWrap *goquery.Selection) {
							restaurant := Restaurant{}

							titleElm := ozWrap.Find(".ozDinIchiTit > h3 > a")
							restaurant.name = titleElm.Text()
							url, _ := titleElm.Attr("href")
							restaurant.url = url

							description, _ := ozWrap.Find(".ozDinIchiTit > .ozDinIchiObj > .ozDinIchiObjInf > p").Attr("href")
							restaurant.description = description
							images := ozWrap.Find(".ozDinIchiTit > .ozDinIchiObj > .ozDinIchiObjImg > a")
							images.Each(func(index int, image *goquery.Selection) {
								if index == 0 {
									imageElm := image.Find("img")
									image_url, _ := imageElm.Attr("src")
									restaurant.image_url = image_url
								}
							})

							restaurants = append(restaurants, restaurant)
						})

						displayNum := 3
						var displayRestaurants Restaurants
						if len(restaurants) > displayNum {
							// ランダムに並び変えてから頭3つをとる
							/*
							restaurantNum := len(restaurants)
							for i := 0; i < displayNum; i++ {

							}
							*/
						} else {
							finalRestaurants := copy(displayRestaurants, restaurants)
						}
						fmt.Printf("%#v", restaurants)
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(inputText)).Do(); err != nil {
						log.Print(err)
					}
					fmt.Printf("%#v", linebot.NewTextMessage(inputText))
					fmt.Println(linebot.NewTextMessage(inputText))
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
