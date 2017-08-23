package protocol

import (
	"bytes"
	"net"
	"os"
	"net/http"
	"encoding/json"
	"encoding/binary"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
	"github.com/badlamb/dexm/blockchain"
	"github.com/badlamb/dexm/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
)

const (
	PORT = ":3141"
)

type Message struct {
	Id   int
	Data []byte
}

var nodeDatabase *leveldb.DB
var bc *blockchain.BlockChain

// Opens the databases needed by many built in tools
func InitPartialNode() {
	if isFolder("blockchain.db"){
		bc = blockchain.OpenBlockchain()
	}else{
		blockchain.NewBlockChain()
	}

	// TODO Add first peer insertion
	nodeDatabase, _ = leveldb.OpenFile("ips.db", nil)
}

// Inspired by https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func isFolder(path string) bool{
	_, err := os.Stat(path)
	if err != nil { return true }
	if os.IsNotExist(err) { return false }

	return true
}

// Start a full node
func StartSyncServer() {
	log.Info("Opening node db..")
	InitPartialNode()

	/* This goroutine contacts known nodes and asks for their ip list */
	go findPeers()

	log.Info("Starting sync webserver...")
	http.HandleFunc("/getaddr", getAddr)
	http.HandleFunc("/getlen", getMaxBlock)
	http.HandleFunc("/getblock", getBlock)
	http.HandleFunc("/newmsg", getMessage)
	http.ListenAndServe(PORT, nil)
}

// getAddr is an http request that returns all known ips
func getAddr(w http.ResponseWriter, r *http.Request) {
	iter := nodeDatabase.NewIterator(nil, nil)

	ips := make(map[string][]byte)

	for iter.Next() {
		ips[string(iter.Key())] = iter.Value()
	}
	iter.Release()

	updateTimestamp(r.RemoteAddr)

	var value []byte
	var err error

	if r.URL.Query().Get("json") != "true"{
		value, err = bson.Marshal(ips)
	}else{
		value, err = json.Marshal(ips)
	}

	if err != nil {
		log.Error(err)
		return
	}
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(value)
}

// getBlock returns a block at index ?index
func getBlock(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("index") != "" {
		index, err := strconv.Atoi(r.FormValue("index"))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		data, err := bc.GetBlock(int64(index))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(data.GetBytes())
	}
}

// Returns how many blocks the client knows
func getMaxBlock(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(int(bc.GetLen()))))
}

// getMessage recives messages from other known peers about events(transactions, blocks etc)
func getMessage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		log.Error(err)
		return
	}

	var recived Message
	err = bson.Unmarshal(body, &recived)
	if err != nil{
		log.Error(err)
		return
	}

	res := false

	// Transaction
	if recived.Id == 1 {
		var t wallet.Transaction
		err = bson.Unmarshal(recived.Data, &t)
		if err != nil{
			log.Error(err)
			return
		}

		res, err = blockchain.VerifyTransactionSignature(t)
		if err != nil {
			return
		}
	}

	// New block
	if recived.Id == 2 {
		var newBlock blockchain.PoWBlock
		err = bson.Unmarshal(recived.Data, &newBlock)
		if err != nil{
			log.Error(err)
			return
		}

		res, err = bc.VerifyNewBlockValidity(&newBlock)
		if err != nil{
			log.Error(err)
			return
		}
	}

	if res{
		go BroadcastMessage(recived.Id, recived.Data)
	}
}

// BroadcastMessage functions shares a message with other peers.
// The algorithm isn't finalized yet and it's very innefficient.
func BroadcastMessage(class int, data []byte) {
	toSend := Message{Id: class, Data: data}
	iter := nodeDatabase.NewIterator(nil, nil)

	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	netClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	bsonStr, _ := bson.Marshal(toSend)

	for iter.Next() {
		req, err := http.NewRequest("POST", "http://"+string(iter.Key())+PORT+"/newmsg", bytes.NewBuffer(bsonStr))
		if err != nil {
			continue
		}

		netClient.Do(req)
	}

}

// Insert IP and timestamp into DB
func updateTimestamp(ip string) {
	// TODO fix local ips properly
	ip, _, err := net.SplitHostPort(ip)
	if err != nil || ip == "127.0.0.1" || ip == "::1" || ip == "" {
		return
	}

	stamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(stamp, uint64(time.Now().Unix()))
	log.Info(stamp)

	err = nodeDatabase.Put([]byte(ip), stamp, nil)
	log.Error(err)
}

func makeRequest(url string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return data, nil
}
