package protocol

import(
    "net/http"
    "time"
    "encoding/binary"
    "encoding/json"

    "github.com/syndtr/goleveldb/leveldb"
    "github.com/badlambd/dexm/blockchain"
    log "github.com/sirupsen/logrus"
)

const (
    PORT = ":3141"
)

var nodeDatabase *leveldb.DB

func StartSyncServer() {
    log.Info("Opening node db..")
    
    chain := blockchain.NewBlockChain()
    
    nodeDatabase, _ = leveldb.OpenFile("ips.db", nil)
    defer nodeDatabase.Close()

    log.Info("Starting sync webserver...")
    http.HandleFunc("/getaddr", getAddr)
    http.ListenAndServe(PORT, nil)
}

/* getAddr is an http request that returns all known ips */
func getAddr(w http.ResponseWriter, r* http.Request) {
    iter := nodeDatabase.NewIterator(nil, nil)

    ips := make(map[string][]byte)

    for iter.Next() {
        ips[string(iter.Key())] = iter.Value()
    }
    iter.Release()

    updateTimestamp(r.RemoteAddr)
    
    value, err := json.Marshal(ips)
    if err != nil{
        log.Error(err)
    }

    w.Write(value)
}

/* getBlock returns a block at index ?index */
func getBlock(w http.ResponseWriter, r* http.Request) {
}

/* getMessage recives messages from other known peers about
   events(transactions, blocks etc)*/
func getMessage(w http.ResponseWriter, r* http.Request) {}


/* Insert IP and timestamp into DB */
func updateTimestamp(ip string) {
    log.Info(ip)
    stamp := make([]byte, 8)
    binary.LittleEndian.PutUint64(stamp, uint64(time.Now().Unix()))
    nodeDatabase.Put([]byte(ip), stamp, nil)
}
