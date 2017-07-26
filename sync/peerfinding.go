package protocol

import (
	"encoding/json"
	"encoding/binary"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const(
	DELAY_BETWEEN_CLEANUPS = 1 * time.Second
	EXPIRATION_PERIOD = 60
)

func findPeers() {
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
	
	for iter.Next() {
		/* TODO clean up given IP and avoid getting tricked into ddosing a server */
		ip := string(iter.Key())

		data, err := makeRequest("http://"+ip+PORT+"/getaddr", netClient)
		if err != nil {
			log.Error(err)
			continue
		}

		ips := make(map[string][]byte)

		log.Info(string(data))

		json.Unmarshal(data, &ips)

		for k, v := range ips {
			e := net.ParseIP(k)
			if e == nil {
				continue
			}

			_, err = nodeDatabase.Get([]byte(k), nil)
			if err != nil {
				nodeDatabase.Put([]byte(k), v, nil)

				/* Once a new IP has been found contact it and ask it for the len of it's chain */
				go func(k string) {
					data, err := makeRequest("http://"+k+PORT+"/getlen", netClient)
					numOfBlocks, err := strconv.Atoi(string(data))
					if err != nil {
						log.Error(err)
						return
					}

					// TODO Fix possible race condition
					if int64(numOfBlocks) > bc.GetLen() {
						log.Info("Found peer with longer chain! Need to sync this amount of blocks:", int64(numOfBlocks)-bc.GetLen())
					}
				}(k)

				continue
			}

			// TODO Timestamp logic
		}
	}

	iter.Release()
}

func AutoIPCleanup(){
	for {
		iter := nodeDatabase.NewIterator(nil, nil)

		for iter.Next() {
			stamp := int64(binary.LittleEndian.Uint64(iter.Value()))

			if stamp + EXPIRATION_PERIOD < time.Now().Unix(){
				nodeDatabase.Delete(iter.Key(), nil)
			}
		}

		time.Sleep(DELAY_BETWEEN_CLEANUPS)
	}
}