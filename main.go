package main

import (
	"net/http"
	"os"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"fmt"
)

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
	fmt.Println("this is bot")
	fmt.Println(bot)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		fmt.Println(events)
		fmt.Println("this is event")
		fmt.Println(events)
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
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
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
