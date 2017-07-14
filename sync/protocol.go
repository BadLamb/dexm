package protocol

import (
	"encoding/json"
	"net/http"
	"strconv"
	"net"
	"time"

	"github.com/badlamb/dexm/blockchain"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
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

	// This goroutine contacts known nodes and asks for their ip list
	go func(){
		iter := nodeDatabase.NewIterator(nil, nil)
		netTransport := &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		}
		netClient := &http.Client{
			Timeout: time.Second * 10,
			Transport: netTransport,
		}
		for iter.Next() {
			// Clean up given IP and avoid getting tricked into ddosing a server
			ip := string(iter.Key())

			resp, err := netClient.Get("http://" + ip + PORT + "/getaddr")
			if err != nil{
				log.Error(err)
				continue
			}


			ips := make(map[string][]byte)
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil{
				log.Error(err)
				continue
			}

			log.Info(string(data))

			json.Unmarshal(data, &ips)

			for k, v := range ips{
				e := net.ParseIP(k)
				if e == nil{
					continue
				}

				_, err = nodeDatabase.Get([]byte(k), nil)
				if err != nil{
					nodeDatabase.Put([]byte(k), v, nil)
					
					// Once a new IP has been found contact it and ask it for the len of it's chain
					go func(k string){
						resp, err := netClient.Get("http://" + k + PORT + "/getlen")
						if err != nil{
							log.Error(err)
							return
						}

						data, err := ioutil.ReadAll(resp.Body)
						if err != nil{
							log.Error(err)
							return
						}

						numOfBlocks, err := strconv.Atoi(string(data))
						if err != nil{
							log.Error(err)
							return
						}
						if numOfBlocks > bc.GetLen(){
							log.Info("Found peer with longer chain! Syncing...")
						}
					}(k)
					
					continue
				}

				// TODO Timestamp logic
			}
		}

		iter.Release()
	}()

	log.Info("Starting sync webserver...")
	http.HandleFunc("/getaddr", getAddr)
	http.HandleFunc("/getlen", getMaxBlock)
	http.HandleFunc("/getblock", getBlock)
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

		data, err := bc.GetBlock(index)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(data.GetBytes())
	}
}

/* Returns how many blocks the client knows */
func getMaxBlock(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(bc.GetLen())))
}

/* getMessage recives messages from other known peers about
   events(transactions, blocks etc)*/
func getMessage(w http.ResponseWriter, r *http.Request) {}

/* Insert IP and timestamp into DB */
func updateTimestamp(ip string) {
	// TODO fix local ips
	ip, _, err := net.SplitHostPort(ip)
	if err != nil || ip == "127.0.0.1" || ip == "::1" || ip == "" {
		return
	}
	stamp := []byte(string(time.Now().Unix()))
	nodeDatabase.Put([]byte(ip), stamp, nil)
}
