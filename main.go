package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MenuItem struct untuk merepresentasikan setiap item di menu
type MenuItem struct {
	Nama   string
	Harga  float64
	Jumlah int
}

// ProsesPemesanan interface mendefinisikan metode untuk memproses pesanan
type ProsesPemesanan interface {
	ProsesPemesanan() float64
	BuktiPesanan()
}

// ProsesPemesanan menghitung total harga untuk setiap item di menu
func (item *MenuItem) ProsesPemesanan() float64 {
	return item.Harga * float64(item.Jumlah)
}

// BuktiPesanan mencetak bukti pesanan untuk setiap item di menu
func (item *MenuItem) BuktiPesanan() {
	fmt.Printf("Item: %s, Jumlah: %d, Total: Rp.%.2f\n", item.Nama, item.Jumlah, item.ProsesPemesanan())
}

// validasiInput memvalidasi input pengguna menggunakan ekspresi reguler
func validasiInput(input string, pola string) bool {
	matched, _ := regexp.MatchString(pola, input)
	return matched
}

// mintaJumlah meminta pengguna untuk memasukkan jumlah untuk setiap item menu
func mintaJumlah(item *MenuItem) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Pulih dari kesalahan:", r)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Masukkan jumlah untuk %s: ", item.Nama)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if !validasiInput(input, `^\d+$`) {
		panic("Format jumlah yang dimasukkan tidak valid.")
	}

	jumlah, err := strconv.Atoi(input)
	if err != nil || jumlah <= 0 {
		panic("Nilai jumlah yang dimasukkan tidak valid.")
	}
	item.Jumlah = jumlah
	return true
}

// pilihItem meminta pengguna untuk memilih item dari menu
func pilihItem(menu []MenuItem) *MenuItem {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Pulih dari kesalahan:", r)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Pilih item dari menu dengan memasukkan nomor:")
	for i, item := range menu {
		fmt.Printf("%d. %s - Rp.%.2f\n", i+1, item.Nama, item.Harga)
	}
	fmt.Print("Masukkan nomor pilihan: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if !validasiInput(input, `^\d+$`) {
		panic("Format pilihan yang dimasukkan tidak valid.")
	}

	pilihan, err := strconv.Atoi(input)
	if err != nil || pilihan < 1 || pilihan > len(menu) {
		panic("Pilihan item yang tidak valid.")
	}

	return &menu[pilihan-1]
}

// pulihDariPanic menangani setiap panic selama eksekusi program
func pulihDariPanic() {
	if r := recover(); r != nil {
		fmt.Println("Terjadi kesalahan:", r)
	}
	fmt.Println("Program selesai.")
}

func main() {
	defer pulihDariPanic()

	// Definisikan menu
	menu := []MenuItem{
		{Nama: "Nasi Goreng", Harga: 20.000},
		{Nama: "Mie Kuah", Harga: 25.000},
		{Nama: "Mie Goreng", Harga: 25.000},
	}

	var pesanan []MenuItem
	var wg sync.WaitGroup
	ch := make(chan string)

	// Proses pesanan pengguna
	for {
		itemDipilih := pilihItem(menu)

		if mintaJumlah(itemDipilih) {
			wg.Add(1)
			go func(item *MenuItem) {
				defer wg.Done()
				pesanan = append(pesanan, *item)
				ch <- fmt.Sprintf("Ditambahkan %d %s ke pesanan.", item.Jumlah, item.Nama)
			}(itemDipilih)
		}

		// Tanya pengguna apakah ingin memesan item lain setelah memasukkan jumlah
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Ingin memesan item lain? (y/n): ")
		lain, _ := reader.ReadString('\n')
		lain = strings.TrimSpace(strings.ToLower(lain))
		if lain != "y" {
			break
		}
	}

	// Tutup channel setelah semua goroutine selesai
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Tangani timeout
	timeout := time.After(5 * time.Second)

	// Cetak pesan dari goroutine menggunakan select untuk menangani beberapa channel
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				ch = nil
			} else {
				fmt.Println(msg)
			}
		case <-timeout:
			fmt.Println("Pemrosesan pesanan terlalu lama.")
			return
		}

		if ch == nil {
			break
		}
	}

	// Hitung dan cetak total pesanan
	var total float64
	for _, order := range pesanan {
		total += order.ProsesPemesanan()
		order.BuktiPesanan()
	}
	fmt.Printf("Total Harga: Rp%.2f\n", total)

	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Pesanan: %v\nTotal Harga: Rp%.2f", pesanan, total)))
	fmt.Println("Pesanan Terencoded:", encoded)

}
