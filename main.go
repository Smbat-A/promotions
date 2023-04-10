package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"taskProject/infrastructure"
	"time"
)

type VerveGroup struct {
	client *mongo.Client
}

func main() {
	client := infrastructure.InitDataLayer()

	go bgTask(client)

	verveGroup := &VerveGroup{client: infrastructure.InitDataLayer()}
	http.HandleFunc("/promotions/", verveGroup.getPromotionByID)
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer client.Disconnect(context.Background())
}

func (verveGroup *VerveGroup) getPromotionByID(w http.ResponseWriter, r *http.Request) {
	// Parse the ID from the URL
	idStr := r.URL.Path[len("/promotions/"):]

	log.Printf(idStr)
	// Find the promotion with the given ID
	var promotions = infrastructure.FindPromotions(idStr, verveGroup.client)
	var promotion *infrastructure.Promotion
	for _, p := range promotions {
		if p.ID == idStr {
			promotion = &p
			break
		}
	}
	if promotion == nil {
		http.NotFound(w, r)
		return
	}

	// Marshal the promotion as JSON and write it to the response
	promotionJSON, err := json.Marshal(promotion)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(promotionJSON)
}

func bgTask(client *mongo.Client) {
	ticker := time.NewTicker(30 * time.Minute)
	for _ = range ticker.C {
		
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		file, err := os.Open(path + "/promotions.csv")
		if err != nil {
			log.Fatalf("Failed to open CSV file: %v", err)
		}
		defer file.Close()
		infrastructure.DeletePromotionsCollection(client)
		Process(file, client)
	}
}

func Process(f *os.File, client *mongo.Client) error {

	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, 1024*1024)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	r := bufio.NewReader(f)

	var wg sync.WaitGroup

	for {
		buf := linesPool.Get().([]byte)

		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				break
			}
			return err
		}

		nextUntillNewline, err := r.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntillNewline...)
		}

		wg.Add(1)
		go func() {
			ProcessChunk(buf, &linesPool, &stringPool, client)
			wg.Done()
		}()

	}

	wg.Wait()
	return nil
}

func ProcessChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, client *mongo.Client) {

	var wg2 sync.WaitGroup

	records := stringPool.Get().(string)
	records = string(chunk)

	linesPool.Put(chunk)

	recordsSlice := strings.Split(records, "\n")

	stringPool.Put(records)

	chunkSize := 300
	n := len(recordsSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {

		wg2.Add(1)
		go func(s int, e int) {
			defer wg2.Done() //to avaoid deadlocks

			promotions := make([]infrastructure.Promotion, 0, e-s)
			for i := s; i < e; i++ {
				text := recordsSlice[i]
				if len(text) == 0 {
					continue
				}
				recordSlice := strings.SplitN(text, ",", 3)

				price, err := strconv.ParseFloat(recordSlice[1], 64)
				if err != nil {
					log.Printf("Failed to parse price: %v", err)
					continue
				}

				promotion := infrastructure.Promotion{
					ID:             recordSlice[0],
					Price:          price,
					ExpirationDate: recordSlice[2],
				}
				promotions = append(promotions, promotion)
				log.Println(promotion)

			}
			infrastructure.AddPromotions(promotions, client)

		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(recordsSlice)))))
	}

	wg2.Wait()
	recordsSlice = nil
}
