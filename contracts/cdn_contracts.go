package contracts

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/badlamb/dexm/wallet"
	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/kothar/brotli-go.v0/enc"
	"gopkg.in/mgo.v2/bson"
)

func CreateCDNContract(files []string, maxCacheNodes uint16, w *wallet.Wallet) (Contract, error) {
	hashes := make(map[string][32]byte)

	// Make hashes of all files
	for i := 0; i < len(files); i++ {
		currFile, err := ioutil.ReadFile(files[i])
		if err != nil {
			return Contract{}, err
		}

		hash := blake2b.Sum256(currFile)
		hashes[files[i]] = hash
	}

	// Make body of contract
	cdn := CDNContract{
		Hashes:        hashes,
		MaxCacheNodes: maxCacheNodes,
	}

	encoded, err := bson.Marshal(cdn)
	if err != nil {
		return Contract{}, err
	}

	// Put it in the envelope
	toSign := Contract{
		Type:       CDN_CONTRACT,
		Definition: encoded,
	}

	toSign.AppendKeyAndSign(w)

	return toSign, nil
}

type CDNFileBundle struct {
	File     []byte `bson:"fb"`
	Filename string `bson:"fn"`
}

func (c Contract) SelectCDNNodes(w *wallet.Wallet) error {
	var body CDNContract
	err := bson.Unmarshal(c.Definition, &body)
	if err != nil {
		return err
	}

	// TODO Randomly select nodes based on proof of burn.

	// Compression settings, this way bundles are smaller
	params := enc.NewBrotliParams()
	params.SetQuality(4)

	for k, _ := range body.Hashes {
		// Compress the file
		rawFile, err := ioutil.ReadFile(k)
		if err != nil {
			return err
		}

		compressedFile, err := enc.CompressBuffer(params, rawFile, make([]byte, 0))
		if err != nil {
			log.Error(err)
		}

		thisBundle := CDNFileBundle{
			Filename: k,
			File:     compressedFile,
		}

		byt, err := bson.Marshal(thisBundle)
		if err != nil {
			return err
		}

		toSend := Contract{
			Type:       CDN_BUNDLE,
			Definition: byt,
		}

		toSend.AppendKeyAndSign(w)
		res, err := bson.Marshal(toSend)
		// TODO Send bundle to all selected nodes
		ProcessCDNBundle(res)
	}

	return nil
}

func ProcessCDNBundle(data []byte) error {
	var decoded Contract
	err := bson.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	var bundle CDNFileBundle
	bson.Unmarshal(decoded.Definition, &bundle)

	filepath := FindCDNFilePath(bundle.Filename, wallet.BytesToAddress(decoded.PubKey))
	log.Info(filepath)
	ioutil.WriteFile(filepath, bundle.File, 0644)

	return nil
}

func FindCDNFilePath(filename, ownerWallet string) string {
	archivePath := "~/.dexmarchive"

	// Look for DEXMARCHIVEPATH Environ and replace the archivePath if found
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if pair[0] == "DEXMARCHIVEPATH" {
			archivePath = pair[1]
		}
	}

	// Check if archivePath exsists, if it doesn't create said folder
	if !filepath.IsAbs(archivePath) {
		log.Fatal(archivePath + " is not an absolute filepath")
	}
	r, err := os.Stat(archivePath)

	if os.IsNotExist(err) {
		os.MkdirAll(archivePath, os.ModePerm)
	}

	if err == nil && !r.IsDir() {
		log.Fatal("File " + archivePath + " collides with archivePath")
	}

	// This part is very fragile, if the path is not properly escaped then it
	// could lead to a path traversal vulnerabilty. OWASP advides urlencoding
	// paths to avoid .. and ~
	return filepath.Join(archivePath, ownerWallet+url.QueryEscape(filename))
}
