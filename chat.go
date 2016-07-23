package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/paked/messenger"
)

const (
	UD            string = "UNDER DEVEfgdfhdg"
	KALIMAT_AWAL1 string = `
	Hi %v!
Terima kasih telah menghubungiku. Perkenalkan aku Tia, aku adalah kecerdasan buatan Tokopedia yang siap untuk membantu kamu :D
	`
	KALIMAT_AWAL2 string = `
	Dalam percakapan ini, kamu boleh memasukkan angka “9” untuk kembali ke menu sebelumnya atau “0” untuk kembali ke menu awal, kapan saja :D
	`

	KALIMAT_AWAL3 string = `
	Nah, untuk memulai layanan ini, boleh aku tahu status kamu di Tokopedia? :)
	`

	KALIMAT_KEDUA                 string = "Informasi apa yang ingin kamu ketahui?"
	KALIMAT_MASUKAN_EMAIL         string = "Masukkan email yang kamu yang kamu gunakan di Tokopedia"
	KALIMAT_PILIH_INVOICE         string = "Hmm, aku menemukan invoice-invoice ini. Mana yang ingin kamu tanyakan?"
	KALIMAT_AKHIR_INVOCIE         string = "Kamu juga bisa memasukkan invoice yang lainnya."
	KALIMAT_TX_TRANSFER_BATAL     string = "tidak perlu khawatir, karena setiap pesanan yang dibatalkan oleh penjual, dana akan kembali ke Saldo Anda, yang bisa digunakan untuk belanja lagi, atau untuk ditarik."
	KALIMAT_PERTANYAAN_CC         string = "Apakah transaksi kamu dibatalkan penjual?"
	KALIMAT_MENU_LAIN             string = "Mohon maaf, Tia belum dapat memproses opsi ini."
	KALIMAT_CC_OLD                string = "Untuk kendala ini, mohon Anda melampirkan bukti tagihan bulanan Anda. Silahkan gunakan fitur attachment di jendela chat berikut ini."
	KALIMAT_CC_NEW                string = "Mohon tunggu hingga maksimal 14 hari kerja untuk mengetahui perkembangan transaksi kamu "
	KALIMAT_UNSUPPORT_INSTALLMENT string = "Untuk kendala ini, tolong lampirkan bukti tagihan bulanan kamu. Silahkan gunakan fitur attachment di jendela chat ini ya."
	KALIMAT_TAGIHAN_ATTACHMENT    string = "Terima kasih %v! Untuk selanjutnya, tim Customer Service kami akan menindaklanjuti laporan ini dalam waktu 1x24 jam."
	KALIMAT_AWAL_CICILAN          string = "Aku mendeteksi bahwa metode pembayaran yang kamu gunakan adalah cicilan. Di antara beberapa opsi berikut ini, manakah yang menjadi kendala kamu?\n"
	KALIMAT_AWAL_CC               string = "Aku mendeteksi bahwa metode pembayaran yang kamu gunakan adalah kartu kredit.\n"
	KALIMAT_AWAL_TRANSFER         string = "Aku mendeteksi bahwa metode pembayaran yang kamu gunakan adalah transfer bank. Di antara beberapa opsi berikut ini, manakah yang menjadi kendala kamu?"
	KALIMAT_FINISH                string = "Ada hal lain yang bisa aku bantu?"
	KALIMAT_AKHIR                 string = "Terima kasih sudah menggunakan Tia. Aku harap kamu mendapatkan pengalaman bertransaksi yang nyaman di Tokopedia. Sampai jumpa lagi~ :D"
	KALIMAT_UNABLE                string = "Mohon maaf, Tia belum dapat memproses opsi ini."
	KALIMAT_NOT_FOUND             string = "Maaf, Tia tidak dapat menemukan email tersebut di Tokopedia"
)

var userChatLoop = make(map[int64]int, 0)
var userPGChoice = make(map[int64]int, 0)
var userTxChoice = make(map[int64]int, 0)

var userFromCrawl = make(map[int64]int, 0)
var userCrawlLoop = make(map[int64]int, 0)
var userEmail = make(map[int64]string, 0)

func ChatHandler(client *messenger.Messenger, m messenger.Message, r *messenger.Response) {
	userID := m.Sender.ID

	if userID == 1875635259330684 {
		return
	}

	defer SetLoop(userID)
	fmt.Printf("%v (Sent, %v)\n", m.Text, m.Time.Format(time.UnixDate))

	p, err := client.ProfileByID(userID)
	if err != nil {
		log.Println("Something went wrong!", err)
	}

	switch m.Text {
	case "9":
		userChatLoop[userID] -= 1
	case "0":
		delete(userChatLoop, userID)
	}

	rcpt := messenger.Recipient{
		ID: m.Sender.ID,
	}

	loop, ok := userChatLoop[userID]
	if loop < 1 || !ok {
		SendKalimatAwalChat(client, r, rcpt, p)
		return
	}

	switch loop {
	case 1:
		switch m.QuickReply.Payload {
		case "1":
			if err := r.TextWithReplies(KALIMAT_KEDUA, GetListQuestionQR(m.QuickReply.Payload)); err != nil {
				log.Println(err)
			}
		default:
			SendKalimatUnable(client, rcpt)
			SendKalimatAwalChat(client, r, rcpt, p)
			userChatLoop[userID] -= 1
		}
	case 2:
		switch m.QuickReply.Payload {
		case "5":
			SendKalimatEmail(r)
		default:
			SendKalimatUnable(client, rcpt)
			if err := r.TextWithReplies(KALIMAT_KEDUA, GetListQuestionQR("1")); err != nil {
				log.Println(err)
			}
			userChatLoop[userID] -= 1
		}
	case 3:
		ui := GetAllInvoices(GetUserIDByEmail(m.Text))
		if len(ui) < 1 {
			SendKalimatNotFound(client, rcpt)
			SendKalimatEmail(r)
			userChatLoop[userID] -= 1
		} else {
			userEmail[userID] = m.Text
			SendAllUserInvoice(r, ui)
		}
	case 4:
		ShowListInvoice(client, rcpt, r, m, userID)
	case 5:
		ShowMenuPaymentGateway(client, rcpt, r, p, m, userID)
	case 6:
		if len(m.Attachments) > 0 {
			SendKalimatAttachTagihan(client, rcpt, p)
		}
		SendKalimatFinish(r)
	case 7:
		switch m.QuickReply.Payload {
		case "0":
			SendKalimatAkhir(r)
		default:
			SendKalimatAwalChat(client, r, rcpt, p)
		}
		delete(userChatLoop, userID)
	}

	return
}

func ChatFromCrawlHandler(client *messenger.Messenger, m messenger.Message, r *messenger.Response) {
	userID := m.Sender.ID
	defer SetCrawlLoop(userID)
	fmt.Printf("%v (Sent, %v)\n", m.Text, m.Time.Format(time.UnixDate))

	p, err := client.ProfileByID(userID)
	if err != nil {
		log.Println("Something went wrong!", err)
	}

	switch m.Text {
	case "9":
		userCrawlLoop[userID] -= 1
	case "0":
		delete(userCrawlLoop, userID)
	}

	rcpt := messenger.Recipient{
		ID: m.Sender.ID,
	}

	loop, ok := userCrawlLoop[userID]
	if loop < 1 || !ok {
		uID := GetUserIDByEmail(m.Text)
		if uID < 1 {
			SendKalimatEmail(r)
			userCrawlLoop[userID] -= 1
		} else {
			SendKalimatPilihInvoice(r, userID)
		}
		return
	}

	switch loop {
	case 1:
		ShowListInvoice(client, rcpt, r, m, userID)
	case 2:
		ShowMenuPaymentGateway(client, rcpt, r, p, m, userID)
	case 3:
		if len(m.Attachments) > 0 {
			SendKalimatAttachTagihan(client, rcpt, p)
		}
		SendKalimatFinish(r)
	case 4:
		switch m.QuickReply.Payload {
		case "0":
			SendKalimatAkhir(r)
		default:
			SendKalimatAwalChat(client, r, rcpt, p)
		}
		delete(userCrawlLoop, userID)
		delete(userFromCrawl, userID)
	}

	return
}

func GetListQuestionQR(id string) []messenger.QuickReply {
	qr := []messenger.QuickReply{}
	i := 1
	for _, q := range GetRelation(id) {
		qr = append(qr, messenger.QuickReply{
			ContentType: "text",
			Title:       GetQuestion(q),
			Payload:     strconv.Itoa(q),
		})
		i++
	}
	return qr
}

func GetListQuestion(id string) string {
	qr := []string{}
	i := 1
	for _, q := range GetRelation(id) {
		qr = append(qr, fmt.Sprintf("%d. %s", i, GetQuestion(q)))
		i++
	}
	return strings.Join(qr, "\n")
}

func SetLoop(userID int64) {
	if _, ok := userChatLoop[userID]; ok {
		userChatLoop[userID] += 1
	} else {
		userChatLoop[userID] = 1
	}
}

func SetCrawlLoop(userID int64) {
	if _, ok := userCrawlLoop[userID]; ok {
		userCrawlLoop[userID] += 1
	} else {
		userCrawlLoop[userID] = 1
	}
}

func SendKalimatAwalChat(client *messenger.Messenger, r *messenger.Response, rcpt messenger.Recipient, p messenger.Profile) {
	if err := client.Send(rcpt, fmt.Sprintf(KALIMAT_AWAL1, p.FirstName), nil); err != nil {
		log.Println(err)
	}
	if err := client.Send(rcpt, KALIMAT_AWAL2, nil); err != nil {
		log.Println(err)
	}
	if err := r.TextWithReplies(KALIMAT_AWAL3, GetListQuestionQR("0")); err != nil {
		log.Println(err)
	}
}

func SendAllUserInvoice(r *messenger.Response, ui []UserInvoice) {
	ivs := []string{}
	for _, iv := range ui {
		ivs = append(ivs, iv.Invoice)
	}
	if err := r.Text(KALIMAT_PILIH_INVOICE + "\n" + strings.Join(ivs, "\n") + "\n" + KALIMAT_AKHIR_INVOCIE); err != nil {
		log.Println(err)
	}
}

func SendKalimatUnable(client *messenger.Messenger, rcpt messenger.Recipient) {
	if err := client.Send(rcpt, KALIMAT_UNABLE, nil); err != nil {
		log.Println(err)
	}
}

func SendKalimatEmail(r *messenger.Response) {
	if err := r.Text(KALIMAT_MASUKAN_EMAIL); err != nil {
		log.Println(err)
	}
}

func SendKalimatNotFound(client *messenger.Messenger, rcpt messenger.Recipient) {
	if err := client.Send(rcpt, KALIMAT_NOT_FOUND, nil); err != nil {
		log.Println(err)
	}
}

func SendKalimatAwalTransfer(r *messenger.Response) {
	if err := r.Text(KALIMAT_AWAL_TRANSFER + "\n" + GetListQuestion("43")); err != nil {
		log.Println(err)
	}
}

func SendKalimatAwalCC(r *messenger.Response) {
	if err := r.Text(KALIMAT_AWAL_CC + KALIMAT_PERTANYAAN_CC); err != nil {
		log.Println(err)
	}
}

func SendKalimatAwalCicilan(client *messenger.Messenger, rcpt messenger.Recipient, r *messenger.Response) {
	if err := client.Send(rcpt, KALIMAT_AWAL_CICILAN, nil); err != nil {
		log.Println(err)
	}
	if err := r.Text(GetListQuestion("41")); err != nil {
		log.Println(err)
	}
}

func SendKalimatTxTransferBatal(client *messenger.Messenger, rcpt messenger.Recipient, p messenger.Profile) {
	if err := client.Send(rcpt, fmt.Sprintf("%v %s", p.FirstName, KALIMAT_TX_TRANSFER_BATAL), nil); err != nil {
		log.Println(err)
	}
}

func SendKalimatMenuLain(client *messenger.Messenger, rcpt messenger.Recipient) {
	if err := client.Send(rcpt, KALIMAT_MENU_LAIN, nil); err != nil {
		log.Println(err)
	}
}

func SendCheckCCTime(client *messenger.Messenger, rcpt messenger.Recipient, userID int64) bool {
	verifTime := GetVerifyTimeByPaymentID(userTxChoice[userID])
	dur := time.Now().Sub(verifTime)
	if (dur.Hours() / 24) > 14 {
		if err := client.Send(rcpt, KALIMAT_CC_OLD, nil); err != nil {
			log.Println(err)
		}

		return false
	} else {
		durDay := int(dur.Hours() / 24)
		if err := client.Send(rcpt, fmt.Sprintf("%s (%s)", KALIMAT_CC_NEW, verifTime.AddDate(0, 0, 14-durDay).Format("02 January 2006")), nil); err != nil {
			log.Println(err)
		}
	}

	return true
}

func SendKalimatUnsupportCicilan(client *messenger.Messenger, rcpt messenger.Recipient) {
	if err := client.Send(rcpt, KALIMAT_UNSUPPORT_INSTALLMENT, nil); err != nil {
		log.Println(err)
	}
}

func SendKalimatAttachTagihan(client *messenger.Messenger, rcpt messenger.Recipient, p messenger.Profile) {
	if err := client.Send(rcpt, fmt.Sprintf(KALIMAT_TAGIHAN_ATTACHMENT, p.FirstName), nil); err != nil {
		log.Println(err)
	}
}

func SendKalimatFinish(r *messenger.Response) {
	qr := []messenger.QuickReply{
		messenger.QuickReply{
			ContentType: "text",
			Title:       "Ya",
			Payload:     "1",
		},
		messenger.QuickReply{
			ContentType: "text",
			Title:       "Tidak",
			Payload:     "0",
		},
	}

	if err := r.TextWithReplies(KALIMAT_FINISH, qr); err != nil {
		log.Println(err)
	}
}

func SendKalimatPilihInvoice(r *messenger.Response, userID int64) {
	if err := r.Text(KALIMAT_PILIH_INVOICE + "\n" + strings.Join(userInvoices[userID], "\n") + "\n" + KALIMAT_AKHIR_INVOCIE); err != nil {
		log.Println(err)
	}
}

func ShowListInvoice(client *messenger.Messenger, rcpt messenger.Recipient, r *messenger.Response, m messenger.Message, userID int64) {
	paymentID := GetPaymentIDByInvoice(m.Text)
	if paymentID < 1 {
		ui := GetAllInvoices(GetUserIDByEmail(userEmail[userID]))
		SendAllUserInvoice(r, ui)
		userChatLoop[userID] -= 1
		userCrawlLoop[userID] -= 1
	} else {
		pgID := GetPaymentGatewayByPaymentID(paymentID)
		log.Println(pgID)
		userTxChoice[userID] = paymentID
		userPGChoice[userID] = pgID
		switch pgID {
		case 1:
			SendKalimatAwalTransfer(r)
		case 8:
			SendKalimatAwalCC(r)
		case 12:
			SendKalimatAwalCicilan(client, rcpt, r)
		}
	}
}

func ShowMenuPaymentGateway(client *messenger.Messenger, rcpt messenger.Recipient, r *messenger.Response, p messenger.Profile, m messenger.Message, userID int64) {
	switch userPGChoice[userID] {
	case 1:
		switch m.Text {
		case "1":
			SendKalimatTxTransferBatal(client, rcpt, p)
			SendKalimatFinish(r)
			userChatLoop[userID] += 1
			userCrawlLoop[userID] += 1
		default:
			ShowGandalf(r)
			delete(userFromCrawl, userID)
			delete(userChatLoop, userID)
			delete(userCrawlLoop, userID)
		}
	case 8:
		switch strings.ToLower(m.Text) {
		case "no", "tidak":
			SendKalimatMenuLain(client, rcpt)
			SendKalimatFinish(r)
			userChatLoop[userID] += 1
			userCrawlLoop[userID] += 1
		case "ya", "iya", "yes":
			if SendCheckCCTime(client, rcpt, userID) {
				SendKalimatFinish(r)
				userChatLoop[userID] += 1
				userCrawlLoop[userID] += 1
			}
		default:
			SendKalimatUnable(client, rcpt)
			SendKalimatAwalCC(r)
			userChatLoop[userID] -= 1
			userCrawlLoop[userID] -= 1
		}
	case 12:
		switch m.Text {
		case "5":
			SendCheckCCTime(client, rcpt, userID)
			SendKalimatFinish(r)
			userChatLoop[userID] += 1
			userCrawlLoop[userID] += 1
		case "1", "2", "3", "4":
			SendKalimatUnsupportCicilan(client, rcpt)
		default:
			SendKalimatMenuLain(client, rcpt)
			SendKalimatFinish(r)
			userChatLoop[userID] += 1
			userCrawlLoop[userID] += 1
		}
	}
}

func SendKalimatAkhir(r *messenger.Response) {
	if err := r.Text(KALIMAT_AKHIR); err != nil {
		log.Println(err)
	}
}

func ShowGandalf(r *messenger.Response) {
	sme := &[]messenger.StructuredMessageElement{
		messenger.StructuredMessageElement{
			Title:    "Go to GANDALF",
			ImageURL: "http://i.imgur.com/SM96caR.gif",
			Subtitle: "Pelajari lebih lanjut tentang Tokopedia",
			Buttons: []messenger.StructuredMessageButton{
				messenger.StructuredMessageButton{
					Type:  "web_url",
					URL:   "https://www.tokopedia.com/contact-us.pl",
					Title: "FLY!",
				},
			},
		},
	}

	if err := r.GenericTemplate("", sme); err != nil {
		log.Println(err)
	}
}
