package contracts

import(
    "net/http"
    "io/ioutil"
    "encoding/json"

    "gopkg.in/olahol/melody.v1"
)

type Header struct{
    Index int `json:"index"`
    BlockSize int `json:"blocksize"`
    File string `json:"filename"`
    Owner string `json:"owner"`

    Proof string `json:"proof,omitempty"`
}

func StartCDNServer() {
    m := melody.New()
    m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

    http.HandleFunc("/register", func(w http.ResponseWriter, r* http.Request) {
        m.HandleRequest(w, r)
    })

    // Handle requests for part of files.
    m.HandleMessage(func(s *melody.Session, msg []byte) {
        var header Header
        
        err := json.Unmarshal(msg, &header)
        if err != nil{
            s.Write([]byte("Invalid header"))
            return
        }

        // Calculate max allowed size and see if BlockSize fits into it

        // If the index is right check the proofs

        // Write the result
        file, err := ioutil.ReadFile(FindCDNFilePath(header.File, header.Owner))

        if err != nil{
            s.Write([]byte(FindCDNFilePath(header.File, header.Owner)))
            return
        }
        
        // TODO Check if we aren't doing an out of bounds read here
        s.Write(file[header.Index : header.Index + header.BlockSize])
    })

    http.ListenAndServe(":8080", nil)
}