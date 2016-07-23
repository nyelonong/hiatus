package main

import (
	"log"
	"time"
)

type UserInvoice struct {
	ID      int
	UserID  int    `db:"customer_id"`
	OrderID int    `db:"order_id"`
	Status  int    `db:"order_status"`
	Invoice string `db:"invoice_ref_num"`
}

func GetAllInvoices(userID int) []UserInvoice {
	ui := []UserInvoice{}

	query := `
		SELECT
			customer_id,
			order_id,
			order_status,
			invoice_ref_num
		FROM ws_order
		WHERE customer_id = $1
		ORDER BY order_id desc
		LIMIT 3
	`

	if err := DBOrder.Select(&ui, query, userID); err != nil {
		log.Println(err)
	}

	return ui
}

func GetUserIDByEmail(email string) int {
	var userID int
	query := `
		SELECT
			user_id
		FROM ws_user
		WHERE user_email = $1
	`
	if err := DBUser.Get(&userID, query, email); err != nil {
		log.Println(err)
	}

	return userID
}

func GetPaymentIDByInvoice(inv string) int {
	var paymentID int
	query := `
		SELECT
			payment_id
		FROM ws_order
		WHERE invoice_ref_num = $1
	`
	if err := DBOrder.Get(&paymentID, query, inv); err != nil {
		log.Println(err)
	}

	return paymentID
}

func GetPaymentGatewayByPaymentID(paymentID int) int {
	var pgID int
	query := `
		SELECT
			pg_id
		FROM ws_payment
		WHERE payment_id = $1
	`
	if err := DBPayment.Get(&pgID, query, paymentID); err != nil {
		log.Println(err)
	}

	return pgID
}

func GetVerifyTimeByPaymentID(paymentID int) time.Time {
	var createTime time.Time
	query := `
		SELECT
			fo2_update_time
		FROM ws_payment
		WHERE payment_id = $1
	`
	if err := DBPayment.Get(&createTime, query, paymentID); err != nil {
		log.Println(err)
	}

	return createTime
}
