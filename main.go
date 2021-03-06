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
	"math/rand"
	"unicode/utf8"
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
						message := "エリアが見つかりませんでした...😂\n別の名称やエリア名を入れてみてください😅"
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message)).Do(); err != nil {
							log.Print(err)
						}
						return
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
							if utf8.RuneCountInString(titleElm.Text()) > 40 {
								restaurant.name = string([]rune(titleElm.Text())[:40])
							} else {
								restaurant.name = titleElm.Text()
							}
							url, _ := titleElm.Attr("href")
							restaurant.url = "https://www.ozmall.co.jp" + url

							description := ozWrap.Find(".ozDinIchiObjInf > p").Text()
							if utf8.RuneCountInString(description) > 60 {
								description = string([]rune(description)[:60])
							}
							restaurant.description = description
							images := ozWrap.Find(".ozDinIchiObjImg > a")
							images.Each(func(index int, image *goquery.Selection) {
								if index == 0 {
									imageElm := image.Find("img")
									image_url, _ := imageElm.Attr("src")
									restaurant.image_url = "https://www.ozmall.co.jp" + image_url
								}
							})

							restaurants = append(restaurants, restaurant)
						})

						displayNum := 3
						if len(restaurants) >= displayNum {
							// ランダムに並び変えてから頭3つをとる
							n := len(restaurants)
							// Fisher-Yates shuffle
							for i := n - 1; i >= 0; i-- {
								j := rand.Intn(i + 1)
								restaurants[i], restaurants[j] = restaurants[j], restaurants[i]
							}
							restaurants = restaurants[0:displayNum]

							template := linebot.NewCarouselTemplate(
								linebot.NewCarouselColumn(
									restaurants[0].image_url, restaurants[0].name, restaurants[0].description,
									linebot.NewURIAction("詳しく見る", restaurants[0].url),
								),
								linebot.NewCarouselColumn(
									restaurants[1].image_url, restaurants[1].name, restaurants[1].description,
									linebot.NewURIAction("詳しく見る", restaurants[1].url),
								),
								linebot.NewCarouselColumn(
									restaurants[2].image_url, restaurants[2].name, restaurants[2].description,
									linebot.NewURIAction("詳しく見る", restaurants[2].url),
								),
							)
							if _, err := bot.ReplyMessage(
								event.ReplyToken,
								linebot.NewTemplateMessage(inputText + "のいい感じのお店が見つかりました！😊", template),
							).Do(); err != nil {
								log.Print(err)
							}
						} else if len(restaurants) == 0 {
							// 0件のとき
							message := inputText + "でいい感じのお店は見つかりませんでした😂\nエリアの変更を検討しましょう😅"
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message)).Do(); err != nil {
								log.Print(err)
							}
							return
						} else if len(restaurants) == 1 {
							template := linebot.NewCarouselTemplate(
								linebot.NewCarouselColumn(
									restaurants[0].image_url, restaurants[0].name, restaurants[0].description,
									linebot.NewURIAction("詳しく見る", restaurants[0].url),
								),
							)
							if _, err := bot.ReplyMessage(
								event.ReplyToken,
								linebot.NewTemplateMessage(inputText + "のいい感じのお店が見つかりました！😊", template),
							).Do(); err != nil {
								log.Print(err)
							}
						} else if len(restaurants) == 2 {
							template := linebot.NewCarouselTemplate(
								linebot.NewCarouselColumn(
									restaurants[0].image_url, restaurants[0].name, restaurants[0].description,
									linebot.NewURIAction("詳しく見る", restaurants[0].url),
								),
								linebot.NewCarouselColumn(
									restaurants[1].image_url, restaurants[1].name, restaurants[1].description,
									linebot.NewURIAction("詳しく見る", restaurants[1].url),
								),
							)
							if _, err := bot.ReplyMessage(
								event.ReplyToken,
								linebot.NewTemplateMessage(inputText + "のいい感じのお店が見つかりました！😊", template),
							).Do(); err != nil {
								log.Print(err)
							}
						}
					}
				}
			}
			if event.Type == linebot.EventTypeFollow {
				message := "こんにちは、デートや女子会、婚活などで使えるお店を提案するTOKYO IKANJINOMISE BOTです✨\n東京都のエリア名や駅名を入力すると、そのエリアでいい感じのお店を提案します！"
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})

	if err := http.ListenAndServe(":" + os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
