package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Koper :  yang berisi data yang akan ditulis ke blockchain.
type Koper struct {
	Pos       int
	Data      KoperDokumen
	Timestamp string
	Hash      string
	PrevHash  string
}

// KoperDokumen : berisi data dokumen
type KoperDokumen struct {
	DokumentID   string `json:"dokumentId"`
	UserKey      string `json:"user-key"`
	DokumentDate string `json:"dokumen-date"`
	IsGenesis    bool   `json:"is_genesis"`
}

// Dokumen : berisi data untuk sampel Dokumen
type Dokumen struct {
	ID          string `json:"id"`
	Judul       string `json:"judul"`
	Notaris     string `json:"Notaris"`
	PublishDate string `json:"publish_date"`
	NoAkta      string `json:"NoAkta"`
}

func (b *Koper) generateHash() {
	// dapatkan nilai string dari Data
	bytes, _ := json.Marshal(b.Data)
	// menyatukan dataset
	data := string(b.Pos) + b.Timestamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

// BuatKoper : ..
func BuatKoper(KoperLama *Koper, checkoutItem KoperDokumen) *Koper {
	koper := &Koper{}
	koper.Pos = KoperLama.Pos + 1
	koper.Timestamp = time.Now().String()
	koper.Data = checkoutItem
	koper.PrevHash = KoperLama.Hash
	koper.generateHash()

	return koper
}

// Blockchain adalah daftar koper yang ber- koperdokumen
type Blockchain struct {
	kopers []*Koper
}

// BlockChain adalah variabel global yang akan mengembalikan struct Blockchain yang dimutasi
var BlockChain *Blockchain

// MasukkanKoper :  menambahkan Koper ke rantai Blockchain
func (bc *Blockchain) MasukkanKoper(data KoperDokumen) {
	// dapatkan koper sebelumnya
	KoperLama := bc.kopers[len(bc.kopers)-1]
	// membuat koper baru
	block := BuatKoper(KoperLama, data)
	//  memvalidasi integritas kopers
	if validBlock(block, KoperLama) {
		bc.kopers = append(bc.kopers, block)
	}
}

// KoperGenesis : ..
func KoperGenesis() *Koper {
	return BuatKoper(&Koper{}, KoperDokumen{IsGenesis: true})
}

// BlockchainBaru : ..
func BlockchainBaru() *Blockchain {
	return &Blockchain{[]*Koper{KoperGenesis()}}
}

// validBlock : ..
func validBlock(block, KoperLama *Koper) bool {
	// Konfirmasikan hash
	if KoperLama.Hash != block.PrevHash {
		return false
	}
	// konfirmasi hash koper jika valid
	if !block.validateHash(block.Hash) {
		return false
	}
	// Periksa posisi untuk mengkonfirmasi kenaikannya atau penambahannya
	if KoperLama.Pos+1 != block.Pos {
		return false
	}
	return true
}

func (b *Koper) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.kopers, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	// meniliskan atau mengirimkan JSON string
	io.WriteString(w, string(jbytes))
}

// MenulisKoperDokumen : ..
func MenulisKoperDokumen(w http.ResponseWriter, r *http.Request) {
	var koperDokumen KoperDokumen
	if err := json.NewDecoder(r.Body).Decode(&koperDokumen); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("tidak bisa mendata Koper: %v", err)
		w.Write([]byte("tidak dapat mendata Koper"))
		return
	}
	// Buat Koper
	BlockChain.MasukkanKoper(koperDokumen)
	resp, err := json.MarshalIndent(koperDokumen, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("tidak dapat menyusun muatan: %v", err)
		w.Write([]byte("tidak dapat mendata Koper"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// DokumenBaru : ..
func DokumenBaru(w http.ResponseWriter, r *http.Request) {
	var dokumen Dokumen
	if err := json.NewDecoder(r.Body).Decode(&dokumen); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create: %v", err)
		w.Write([]byte("could not create new Dokumen"))
		return
	}
	// Kami akan membuat ID, menggabungkan NoAkta dan mempublikasikan tanggal
	// contoh ini hanya menggunakan md5 . tidak direkomendasikan untuk di publis menggunakan md5 okey
	h := md5.New()
	io.WriteString(h, dokumen.NoAkta+dokumen.PublishDate)
	dokumen.ID = fmt.Sprintf("%x", h.Sum(nil))

	// mengirim kembali muatan data
	resp, err := json.MarshalIndent(dokumen, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("tidak bisa me-marshal muatan : %v", err)
		w.Write([]byte("tidak bisa menyimpan data koper"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func main() {
	// inisialisasi blockchain dan simpan di variable
	BlockChain = BlockchainBaru()

	// register router menggunakan mux
	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", MenulisKoperDokumen).Methods("POST")
	r.HandleFunc("/new", DokumenBaru).Methods("POST")

	// melihat status Blockchain ke konsol kita
	go func() {
		//for {
		for _, block := range BlockChain.kopers {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
		//}
	}()
	log.Println("server port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))
}
