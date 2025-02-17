package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/randomizedcoder/xtcp2/pkg/clickhouse_protolist"
	"google.golang.org/protobuf/proto"
)

//import "google.golang.org/protobuf/encoding/protowire"
//import "google.golang.org/protobuf/encoding/protodelim"

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
	envelope     bool
	db           bool
	filename     string
	value        uint
	debugDump    bool
	dumpFilename string
}

func main() {

	filename := flag.String("filename", "protoBytes.bin", "filename")
	value := flag.Uint("value", 1, "value uint -> uint32")

	envelope := flag.Bool("envelope", false, "envelope")
	db := flag.Bool("db", false, "db")

	dump := flag.Bool("dump", false, "dump proto for debug")
	dumpFilename := flag.String("dumpFileName", "dump.bin", "dump file name")

	v := flag.Bool("v", false, "show version")

	flag.Parse()

	// Print version information passed in via ldflags in the Makefile
	if *v {
		log.Printf("commit:%s\tdate(UTC):%s\tversion:%s", commit, date, version)
		os.Exit(0)
	}

	c := config{
		filename:     *filename,
		value:        *value,
		envelope:     *envelope,
		db:           *db,
		debugDump:    *dump,
		dumpFilename: *dumpFilename,
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

	serializedData, err := proto.Marshal(r)
	if err != nil {
		log.Fatal("serializedData, err := proto.Marshal(r):", err)
	}

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(serializedData)))

	binaryData = make([]byte, 0, n+len(serializedData))
	binaryData = append(binaryData, buf[:n]...)
	binaryData = append(binaryData, serializedData...)

	// 	[das@t:~/Downloads/xtcp2/cmd/clickhouse_protobuflist_db]$ hexdump -C ./fixed_output.bin
	// 00000000  06 08 ff ff ff ff 0f                              |.......|
	// 00000007
	if c.debugDump {
		errW := os.WriteFile(c.dumpFilename, binaryData, 0644)
		if errW != nil {
			log.Fatalf("Failed to write protobuf data: %v", errW)
		}
	}

	if !c.envelope {
		return binaryData
	}

	if c.envelope {
		log.Fatal("evenope not implemented")
	}

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
