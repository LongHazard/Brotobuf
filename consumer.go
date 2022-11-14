package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	eventpb "nat2es/event"

	//"github.com/elastic/go-elasticsearch"

	//"github.com/elastic/go-elasticsearch"

	//"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/olivere/elastic"
	"google.golang.org/protobuf/encoding/protojson"
)

var client *elasticsearch.Client
var docID int = 0

// var Elastic *elasticsearch
// delay for a while
func delay() {
	time.Sleep(time.Millisecond * 100)
}

// moveOnWithError to the next process
func moveOnWithError(err error) {
	log.Println(err.Error())
	log.Println("re-estabilishing nats subscription now ...")
	delay()
}

func KeepAliveSubscribe(subject string) {
	// infinity loop (with some delay) enable auto reconnection easier
	for {
		log.Printf("subscribing to nats server with subject '%s'\n", subject)
		// Connect to NATS
		nc, err := nats.Connect(NTP_NATS_SERVER)
		if err != nil {
			moveOnWithError(fmt.Errorf("ERROR: unable to create nats connection object. %s", err.Error()))
			continue
		}
		// create nats connection object server connection
		js, err := JetStreamInit(nc)
		if err != nil {
			moveOnWithError(fmt.Errorf("ERROR: unable to init JetStream connection object. %s", err.Error()))
			continue
		}

		// subscribe to nats server
		_, err = js.Subscribe(subject, func(msg *nats.Msg) {
			// gracefully handle incoming requestk
			err := msg.Ack()

			if err != nil {
				log.Println("Unable to Ack", err)
				return
			}

			protoMsg := &eventpb.EventLog{}
			err = proto.Unmarshal(msg.Data, protoMsg)
			if err != nil {
				log.Printf("Error Unmarshal message from binary %v\n", err)
			}

			jsonBytes, _ := protojson.Marshal(protoMsg)
			log.Printf("Got a message from subject %s\n", subject)
			jst := string(jsonBytes)

			//fmt.Println(jst)
			// cho phép định dạng
			log.SetFlags(0)

			// Sử dụng gói Olivere để lấy số phiên bản Elasticsearch
			fmt.Println("Version: ", elastic.Version)

			// Tạo một đối tượng ngữ cảnh cho các lệnh gọi API
			ctx := context.Background()

			// Khai báo một phiên bản ứng dụng khách của trình điều khiển Olivere
			client, err := elastic.NewClient(
				elastic.SetSniff(false),
				elastic.SetURL("http://192.168.14.151:9200"),
				elastic.SetBasicAuth("elastic", "OZpbNlZ1VsoIHh38gt4i"),
				elastic.SetHealthcheckInterval(5*time.Second),
			)

			// Kiểm tra xem phương thức NewClient () của olivere có trả lại lỗi
			if err != nil {
				fmt.Println("elastic.NewClient() ERROR: ", err)
				log.Fatalf("quiting connection...")
			} else {
				fmt.Println("client: ", client)
				fmt.Println("client TYPE:", reflect.TypeOf(client))
			}

			indexName := "student"

			indices := []string{indexName}

			existService := elastic.NewIndicesExistsService(client)

			existService.Index(indices)
			exist, err := existService.Do(ctx)

			// Kiểm tra xem phương thức IndicesExistsService.Do () có trả về bất kỳ lỗi nào
			if err != nil {
				log.Fatalf("IndicesExistsService.Do() ERROR:", err)

			} else if exist == false {
				fmt.Println("nOh no! The index", indexName, "doesn't exist.")
				fmt.Println("Create the index, and then run the Go script again")
			} else if exist == true {
				fmt.Println("Index name", indexName, " exist!")

				bulk := client.Bulk()
				docID++
				idStr := strconv.Itoa(docID)
				// doc.Timestamp = time.Now().Unix()
				// fmt.Println("ntime.Now().Unix():", doc.Timestamp)
				req := elastic.NewBulkIndexRequest()
				req.OpType("index")
				req.Index(indexName)
				req.Id(idStr)
				req.Doc(jst)
				fmt.Println("rep:", req)
				fmt.Println("req TYPE:", reflect.TypeOf(req))
				bulk = bulk.Add(req)
				fmt.Println("NewBulkIndexRequest().NumberOfActions():", bulk.NumberOfActions())

				// gửi hàng loạt các yêu cầu đến elastic

				bulkResp, err := bulk.Do(ctx)
				// Kiểm tra xem có bắn trả lỗi nào không
				if err != nil {
					log.Fatalf("bulk.Do(ctx) ERROR", err)
				} else {
					// nếu không thấy lỗi thì lấy phản hồi từ API
					indexed := bulkResp.Indexed()
					fmt.Println("nbulkResp.Indexed():", indexed)
					fmt.Println("bulkResp.Indexed() TYPE:", reflect.TypeOf(indexed))

					// Lặp lại đối tượng BulResp.Indexed () được trả về từ số lượng lớn.go
					t := reflect.TypeOf(indexed)
					fmt.Println("nt:", t)
					fmt.Println("NewBulkIndexRequest().NumberOfActions():", bulk.NumberOfActions())

					// Iterate over the document responses
					for i := 0; i < t.NumMethod(); i++ {
						method := t.Method(i)
						fmt.Println("nbulkResp.Indexed() METHOD NAME:", i, method.Name)
						fmt.Println("bulkResp.Indexed() method:", method)
					}

					// Return data on the documents indexed
					fmt.Println("nBulk response Index:", indexed)
					for _, info := range indexed {
						fmt.Println("nBulk response Index:", info)
					}
				}
			}
		})

		if err != nil {
			moveOnWithError(fmt.Errorf("ERROR: unable to subscribe to nats server with subject '%s'. %s", subject, err.Error()))
			continue
		}

		log.Println("connected")

		// watch nats connection state
		chSubIsClosed := make(chan bool)
		go func(chSubIsClosed chan bool, conn *nats.Conn) {
			// check nats conn every 0.1 second
			for nc.IsConnected() {
				delay()
			}
			// if it's not connected, then close the current conn
			conn.Close()
			// next, send signal to chSubIsClosed
			chSubIsClosed <- true
		}(chSubIsClosed, nc)

		// go to next loop for re-estabilishing subscription in case current conn is closed
		<-chSubIsClosed
		moveOnWithError(fmt.Errorf("ERROR: nats subscription is closed"))
	}
}
