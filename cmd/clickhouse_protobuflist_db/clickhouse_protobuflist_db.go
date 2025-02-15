package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/randomizedcoder/xtcp2/pkg/clickhouse_protolist"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const (
	clickhouseConnectString = "127.0.0.1:9001"
	clickhouseUser          = "dave"
	clickhousePassword      = "dave"
)

var (
	// Passed by "go build -ldflags" for the show version
	commit  string
	date    string
	version string
)

type config struct {
	envelope bool
	db       bool
	filename string
	value    uint
}

func main() {

	filename := flag.String("filename", "protoBytes.bin", "filename")
	value := flag.Uint("value", 1, "value uint -> uint32")

	envelope := flag.Bool("envelope", false, "envelope")
	db := flag.Bool("db", false, "db")

	v := flag.Bool("v", false, "show version")

	flag.Parse()

	// Print version information passed in via ldflags in the Makefile
	if *v {
		log.Printf("commit:%s\tdate(UTC):%s\tversion:%s", commit, date, version)
		os.Exit(0)
	}

	c := config{
		filename: *filename,
		value:    *value,
		envelope: *envelope,
		db:       *db,
	}

	primaryFunction(c)
}

func primaryFunction(c config) {

	binaryData := prepareBinary(c)

	fileOrDB(c, binaryData)

}

func prepareBinary(c config) (binaryData []byte) {

	r := &clickhouse_protolist.Record{}
	r.MyUint32 = uint32(c.value)

	encodedData, err := encodeLengthDelimitedProtobufList(r)
	if err != nil {
		log.Println("Error encoding:", err)
		return
	}

	if !c.envelope {
		binaryData = encodedData
		return binaryData
	}

	e := &clickhouse_protolist.Envelope{}
	e.Rows = append(e.Rows, r)

	encodedEnvelope, err := encodeLengthDelimitedEnvelope(encodedData)
	if err != nil {
		fmt.Println("Error encoding Envelope:", err)
		return
	}

	binaryData = encodedEnvelope

	return binaryData
}

func fileOrDB(c config, binaryData []byte) {

	if !c.db {
		errW := writeDataToFile(c.filename, binaryData)
		if errW != nil {
			log.Println("Error:", errW)
		}
		os.Exit(0)
	}

	errDB := insertIntoCH(c, binaryData)
	if errDB != nil {
		log.Println("Error:", errDB)
	}

}

func encodeLengthDelimitedProtobufList(r *clickhouse_protolist.Record) (result []byte, err error) {

	// for _, record := range e.Rows {
	// 	recordBytes, err := proto.Marshal(record)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error marshaling Record: %v", err)
	// 	}

	recordBytes, err := proto.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Record: %v", err)
	}

	log.Printf("AppendVarint of length:%d", len(recordBytes))
	protowire.AppendVarint(result, uint64(len(recordBytes)))

	result = append(result, recordBytes...)

	// }

	return result, nil
}

func encodeLengthDelimitedEnvelope(encodedData []byte) (result []byte, err error) {

	result = append(result, protowire.AppendVarint(nil, uint64(len(encodedData)))...)
	result = append(result, encodedData...)

	return result, nil
}

func writeDataToFile(filename string, data []byte) error {

	err := os.WriteFile(filename, data, 0644) // 0644 permissions (rw-r--r--)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err) // Wrap the error
	}
	return nil

}

func insertIntoCH(c config, binaryData []byte) error {

	ctx := context.TODO()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{clickhouseConnectString}, // Update with your ClickHouse address
		Auth: clickhouse.Auth{
			Database: "",
			Username: clickhouseUser,
			Password: clickhousePassword,
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()

	insertQuery := `
INSERT INTO clickhouse_protolist.clickhouse_protolist (my_uint32)
FORMAT Protobuf
SETTINGS format_schema = '/var/lib/clickhouse/format_schemas/clickhouse_protolist.proto:clickhouse_protolist.v1.Record;
`

	if c.envelope {
		insertQuery = strings.Replace(insertQuery, "FORMAT Protobuf", "FORMAT ProtobufList", 1)
	}

	insertQuery = strings.Replace(insertQuery, "\n", " ", -1)

	log.Printf("insertQuery:%s", insertQuery)

	// batch, err := conn.PrepareBatch(ctx, insertQuery)
	// if err != nil {
	// 	log.Fatalf("Failed to prepare batch: %v", err)
	// }

	// if err := batch.Append(binaryData); err != nil {
	// 	log.Fatalf("Failed to append data: %v", err)
	// }

	// if err := batch.Send(); err != nil {
	// 	log.Fatalf("Failed to send batch: %v", err)
	// }
	err = conn.Exec(ctx, insertQuery, binaryData)
	if err != nil {
		log.Fatalf("Failed to execute Protobuf insert: %v", err)
	}

	return nil
}
