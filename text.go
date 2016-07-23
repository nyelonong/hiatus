package main

import (
	"math/rand"
)

func GetQuestion(uid int) string {
	m := make(map[int]string)

	m[1] = "Pembeli"
	m[2] = "Penjual"
	m[3] = "Pengguna Baru"
	m[4] = "Lainnya"

	m[5] = "Transaksi"
	m[6] = "Dana"
	m[7] = "Promo"
	m[8] = "Fitur"

	m[9] = "Transaksi dibatalkan oleh penjual"
	m[10] = "Pembayaran belum diverifikasi"
	m[11] = "Lupa konfirmasi pembayaran"
	m[12] = "Salah data konfirmasi pembayaran"
	m[13] = "Lebih atau kurang bayar"

	m[14] = "Cicilan yang ditagihkan tidak sesuai"
	m[15] = "Transaksi berhasil tapi mendapat sms cicilan tidak bisa diproses"
	m[16] = "Cicilan 0" + "%" + " dikenakan bunga"
	m[17] = "Tagihan yang masuk bukan berbentuk cicilan"
	m[18] = "Transaksi dibatalkan oleh penjual"

	t, ok := m[uid]
	if ok {
		return t
	}

	return "Not Found."
}

func GetRelation(uid string) []int {
	m := make(map[string][]int)

	m["0"] = []int{1, 2, 3, 4}
	m["1"] = []int{5, 6, 7, 8}
	m["43"] = []int{9, 10, 11, 12, 13}
	m["41"] = []int{14, 15, 16, 17, 18}

	t, ok := m[uid]
	if ok {
		return t
	}

	return []int{}
}

func GetMessengerID(uid int64) int64 {
	m := make(map[int64]int64)

	m[10203174190768205] = 1224713740886001
	// m[1082946775125891] = 1224713740886001

	t, ok := m[uid]
	if ok {
		return t
	}

	return rand.Int63()
}
