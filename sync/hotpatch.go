package protocol

import (
	"os"
	"bytes"

	"github.com/kr/binarydist"
	log "github.com/sirupsen/logrus"
)

type Update struct {
	IsDiff    bool
	Timestamp int64
	Data      []byte
}

func FindDiff(f1, f2 string) string{
	reader1, err1 := os.Open(f1)
	reader2, err2 := os.Open(f2)

	if err1 != nil || err2 != nil {
		log.Error("Error opening files!")
		return ""
	}

	buf := new(bytes.Buffer)
	err := binarydist.Diff(reader1, reader2, buf)
	if err != nil{
		log.Error(err)
		return ""
	}

	return buf.String()
}

