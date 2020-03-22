package aria2

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

const (
	moviesFolder  = "/data/Movies"
	bangumiFolder = "/data/Bangumi"
)

const (
	contentType = "application/json-rpc"
)

type aria2Cli struct {
	RPCAddr string `json:"rpc"`
	Token   string `json:"token"`
}

func newAria2Cli(rpc, token string) *aria2Cli {
	return &aria2Cli{
		RPCAddr: rpc,
		Token:   token,
	}
}

type addUriRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type addUriResponse struct {
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

func (a *aria2Cli) AddUri(dir string, links ...string) error {
	req := addUriRequest{
		Jsonrpc: "2.0",
		ID:      "TelegramBot",
		Method:  "aria2.addUri",
		Params: []interface{}{
			a.Token,
			links,
			struct {
				Dir string `json:"dir"`
			}{dir},
		},
	}
	b, err := jsoniter.Marshal(req)
	if err != nil {
		log.Printf("unable to marshal request: %v", err)
		return err
	}

	resp, err := http.Post(a.RPCAddr, contentType, bytes.NewReader(b))
	if err != nil {
		log.Printf("unable to post request: %v", err)
		return err
	}

	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	jresp := new(addUriResponse)
	err = jsoniter.Unmarshal(rb, jresp)
	if err == nil && jresp.Result != "" {
		log.Printf("addUri successfully with gid: %v", jresp.Result)
	}
	return err
}
