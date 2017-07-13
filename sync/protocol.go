package protocol

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/badlamb/dexm/blockchain"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	PORT = ":3141"
)

var nodeDatabase *leveldb.DB
var bc *blockchain.BlockChain

func StartSyncServer() {
	log.Info("Opening node db..")

	chain := blockchain.NewBlockChain()
	bc = chain

	nodeDatabase, _ = leveldb.OpenFile("ips.db", nil)
	defer nodeDatabase.Close()

	log.Info("Starting sync webserver...")
	http.HandleFunc("/getaddr", getAddr)
	http.HandleFunc("/getlen", getMaxBlock)
	http.ListenAndServe(PORT, nil)
}

/* getAddr is an http request that returns all known ips */
func getAddr(w http.ResponseWriter, r *http.Request) {
	iter := nodeDatabase.NewIterator(nil, nil)

	ips := make(map[string][]byte)

	for iter.Next() {
		ips[string(iter.Key())] = iter.Value()
	}
	iter.Release()

	updateTimestamp(r.RemoteAddr)

	value, err := json.Marshal(ips)
	if err != nil {
		log.Error(err)
	}

	w.Write(value)
}

/* getBlock returns a block at index ?index */
func getBlock(w http.ResponseWriter, r *http.Request) {
	if r.Form.Get("index") != "" {
		index, err := strconv.Atoi(r.Form.Get("index"))
		if err != nil {
			w.Write([]byte("Error"))
			return
		}

		data, err := bc.GetBlock(index)
		if err != nil {
			w.Write([]byte("Error"))
			return
		}

		w.Write(data.GetBytes())
	}
}

/* Returns how many blocks the client knows */
func getMaxBlock(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(string(bc.GetLen())))
}

/* getMessage recives messages from other known peers about
   events(transactions, blocks etc)*/
func getMessage(w http.ResponseWriter, r *http.Request) {}

/* Insert IP and timestamp into DB */
func updateTimestamp(ip string) {
	log.Info(ip)
	stamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(stamp, uint64(time.Now().Unix()))
	nodeDatabase.Put([]byte(ip), stamp, nil)
}
