package protocol

import (
	"os"
	"bytes"
	"errors"
	"time"
	"os/exec"

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
WARNING: THIS IS VERY VERY VERY (Very) HACKY
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

	hash1 := blake2b.Sum256(data1)
	hash2 := blake2b.Sum256(data2)

	return &Update{
		IsDiff: true,
		Timestamp: time.Now().Unix(),
		Data: buf.Bytes(),
		OldHash: hash1[:],
		NewHash: hash2[:],
	}
}

func (u *Update) Apply(oldFile, newFile string) error {
	oldReader, err1 := os.Open(oldFile)
	if err1 != nil{
		return err1
	}

	defer oldReader.Close()

	data, err := ioutil.ReadAll(oldReader)
	hash := blake2b.Sum256(data)

	if !TestEq(hash[:], u.OldHash){
		return errors.New("Old hashes don't match!")
	}

	file, _ := ioutil.TempFile("/tmp", "dexmpatch")
	file.Write(u.Data)

	err = exec.Command("bspatch", oldFile, newFile, file.Name()).Run()
	if err != nil{
		return err
	}

	exec.Command("chmod", newFile, "+x").Run()

	newFileS, err := os.OpenFile(newFile, os.O_RDWR, 0666)
	if err != nil{
		return err
	}
	defer newFileS.Close()

	data, err = ioutil.ReadAll(newFileS)
	if err != nil{
		return err
	}

	hash = blake2b.Sum256(data)
	if !TestEq(hash[:], u.NewHash){
		return errors.New("New hashes don't match!")
	}

	return nil
}

// https://stackoverflow.com/questions/15311969/checking-the-equality-of-two-slices
func TestEq(a, b []byte) bool {
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