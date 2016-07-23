package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	fb "github.com/huandu/facebook"
	"github.com/paked/messenger"
)

const (
	FROM_CRAWL int = 34

	SALAM     string = "Perkenalkan saya Tia, saya adalah kecerdasan buatan Tokopedia untuk membantu Anda."
	MSG_SEND3 string = "Untuk verifikasi, bisakah kamu masukkan email yang kamu gunakan di Tokopedia? :)"
	MSG_SEND1 string = `
	Hi %v! Perkenalkan aku Tia, aku adalah kecerdasan buatan Tokopedia yang siap untuk membantu kamu :D
	`
	MSG_SEND2 string = `
	Firasat Tia mengatakan bahwa kamu mempunyai masalah untuk invoice
%s
Wah, ayo Tia bantu kamu sekarang!
	`
)

var (
	commentFlag  = []string{}
	userFlag     = []string{}
	userInvoices = make(map[int64][]string)
)

type FacebookFeed struct {
	ID          string           `facebook:"id"`
	Message     string           `facebook:"message"`
	FeedFrom    FacebookFeedFrom `facebook:"from"` // use customized field name "from"
	CreatedTime string           `facebook:"created_time"`
}

type FacebookFeedFrom struct {
	Name, ID string
}

func CrawlHandler() error {
	res, err := fb.Get(fmt.Sprintf("/%s/posts", Config.Token.PageID), fb.Params{
		"access_token": Config.Token.AccessToken,
	})

	if err != nil {
		log.Println(err)
		return err
	}

	var items []fb.Result
	if err := res.DecodeField("data", &items); err != nil {
		log.Println(err)
		return err
	}

	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      Config.Token.Verify,
		VerifyToken: Config.Token.VerifyToken,
		Token:       Config.Token.PageToken,
	})

	for _, item := range items {
		id := ""
		if str, ok := item["id"].(string); ok {
			id = str
		}
		if err := GetComments(id, client); err != nil {
			log.Println(err)
			continue
		}
	}

	for k, v := range userInvoices {
		rcpt := messenger.Recipient{
			ID: k,
		}

		p, err := client.ProfileByID(k)
		if err != nil {
			log.Println("Something went wrong!", err)
		}

		inv := strings.Join(v, "\n")
		if err := client.Send(rcpt, fmt.Sprintf(MSG_SEND1, p.FirstName), nil); err != nil {
			log.Println(err)
		}
		if err := client.Send(rcpt, fmt.Sprintf(MSG_SEND2, inv), nil); err != nil {
			log.Println(err)
		}
		if err := client.Send(rcpt, fmt.Sprintf(MSG_SEND3), nil); err != nil {
			log.Println(err)
		}

		userFromCrawl[k] = 1
	}

	return nil
}

func GetComments(postID string, client *messenger.Messenger) error {
	res, err := fb.Get(fmt.Sprintf("/%s/comments", postID), fb.Params{
		"access_token": Config.Token.AccessToken,
	})

	if err != nil {
		log.Println(err)
		return err
	}

	var items []FacebookFeed
	if err := res.DecodeField("data", &items); err != nil {
		log.Println(err)
		return err
	}

	r, err := regexp.Compile(`INV[\/\w]+`)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range items {
		uid, err := strconv.ParseInt(item.FeedFrom.ID, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		id := GetMessengerID(uid)
		invs := r.FindAllString(item.Message, -1)
		for _, inv := range invs {
			if _, ok := userInvoices[id]; ok {
				userInvoices[id] = append(userInvoices[id], inv)
			} else {
				userInvoices[id] = []string{inv}
			}
		}
	}

	return nil
}
