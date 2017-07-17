package protocol

import (
	"os"
	"bytes"
	"errors"
	"time"

	"github.com/kr/binarydist"
	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Update struct {
	IsDiff    bool
	Timestamp int64
	Data      []byte
	OldHash   []byte
	NewHash   []byte
}

/*
WARNING: THIS DOES NOT WORK DO NOT USE IT 
OR IT WILL BREAK YOUR BINARY, JUST DON'T k?
*/

func FindDiff(f1, f2 string) *Update{
	reader1, err1 := os.Open(f1)
	reader2, err2 := os.Open(f2)

	if err1 != nil || err2 != nil {
		log.Error("Error opening files!")
		return nil
	}

	defer reader1.Close()
	defer reader2.Close()

	buf := new(bytes.Buffer)
	err := binarydist.Diff(reader1, reader2, buf)
	if err != nil{
		log.Error(err)
		return nil
	}

	data1, err1 := ioutil.ReadFile(f1)
	data2, err2 := ioutil.ReadFile(f2)
	if err1 != nil || err2 != nil {
		log.Error("Error opening files!")
		return nil
	}

	return &Update{
		IsDiff: true,
		Timestamp: time.Now().Unix(),
		Data: buf.Bytes(),
		OldHash: blake2b.New256().Sum(data1),
		NewHash: blake2b.New256().Sum(data2),
	}
}

func (u *Update) Apply(oldFile, newFile string) error {
	oldReader, err1 := os.Open(oldFile)
	if err1 != nil{
		return err1
	}

	defer oldReader.Close()

	data, err := ioutil.ReadAll(oldReader)
	if !testEq(blake2b.New256().Sum(data), u.OldHash){
		//return errors.New("Old hashes don't match!")
	}

	newFileS, err := os.Create(newFile)
	if err != nil{
		return err
	}
	defer newFileS.Close()

	err = binarydist.Patch(oldReader, newFileS, bytes.NewBuffer(u.Data))
	if err != nil{
		return err
	}

	data, err = ioutil.ReadAll(newFileS)
	if err != nil{
		return err
	}

	if !testEq(blake2b.New256().Sum(data), u.NewHash){
		return errors.New("New hashes don't match!")
	}

	return nil
}

func testEq(a, b []byte) bool {
    if a == nil && b == nil { 
        return true; 
    }

    if a == nil || b == nil { 
        return false; 
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}