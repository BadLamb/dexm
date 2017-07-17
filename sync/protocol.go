package protocol

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/badlamb/dexm/blockchain"
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

func InitPartialNode() {
	chain := blockchain.NewBlockChain()
	bc = chain

	nodeDatabase, _ = leveldb.OpenFile("ips.db", nil)
}

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
		return
	}

	w.Write(value)
}

/* getBlock returns a block at index ?index */
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

/* Returns how many blocks the client knows */
func getMaxBlock(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(int(bc.GetLen()))))
}

/* getMessage recives messages from other known peers about
   events(transactions, blocks etc)*/
func getMessage(w http.ResponseWriter, r *http.Request) {
	log.Info(ioutil.ReadAll(r.Body))
}

func BroadcastMessage(class int, data []byte) {
	toSend := Message{Id: class, Data: data}
	InitPartialNode()
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

	jsonStr, _ := json.Marshal(toSend)

	for iter.Next() {
		req, err := http.NewRequest("POST", "http://"+string(iter.Key())+PORT+"/newmsg", bytes.NewBuffer(jsonStr))
		if err != nil {
			continue
		}

		netClient.Do(req)
	}

}

/* Insert IP and timestamp into DB */
func updateTimestamp(ip string) {
	// TODO fix local ips properly
	ip, _, err := net.SplitHostPort(ip)
	if err != nil || ip == "127.0.0.1" || ip == "::1" || ip == "" {
		return
	}

	stamp := []byte(string(time.Now().Unix()))
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
