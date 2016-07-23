package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nyelonong/finapimate/utils"
	"github.com/paked/messenger"
)

var (
	Config                     *utils.GConfig
	DBOrder, DBUser, DBPayment *sqlx.DB
)

func main() {
	log.SetFlags(log.Lshortfile)

	var err error
	Config, err = utils.NewConfig("config.ini")
	if err != nil {
		log.Fatalln(err)
	}

	DBOrder, err = sqlx.Connect("postgres", Config.Database.Order)
	if err != nil {
		log.Fatalln(err)
	}

	DBUser, err = sqlx.Connect("postgres", Config.Database.User)
	if err != nil {
		log.Fatalln(err)
	}

	DBPayment, err = sqlx.Connect("postgres", Config.Database.Payment)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      Config.Token.Verify,
		VerifyToken: Config.Token.VerifyToken,
		Token:       Config.Token.PageToken,
	})

	for k, _ := range userInvoices {
		delete(userInvoices, k)
	}

	CrawlHandler()

	// Setup a handler to be triggered when a message is received
	client.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		if l, ok := userFromCrawl[m.Sender.ID]; ok || l > 1 {
			ChatFromCrawlHandler(client, m, r)
		} else {
			ChatHandler(client, m, r)
		}
	})

	fmt.Println("Serving Sacred & Minotaur")
	http.ListenAndServe("localhost:31337", client.Handler())
}
