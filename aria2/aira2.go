package aria2

import (
	"bytes"
	"encoding/base64"
	"fmt"
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

type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Result interface{} `json:"result"`
}

func (a *aria2Cli) doRPC(req rpcRequest) (rpcResponse, error) {
	rpcResp := rpcResponse{}
	b, err := jsoniter.Marshal(req)
	if err != nil {
		log.Printf("unable to marshal request: %v", err)
		return rpcResp, err
	}

	resp, err := http.Post(a.RPCAddr, contentType, bytes.NewReader(b))
	if err != nil {
		log.Printf("unable to post request: %v", err)
		return rpcResp, err
	}

	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(rb, &rpcResp)
	if err != nil {
		log.Print("unable to unmarshal response")
		return rpcResp, err
	}

	if rpcResp.Error.Code != 0 {
		log.Printf("rpc error: %v", rpcResp.Error.Message)
		return rpcResp, fmt.Errorf("rpc error: %v", rpcResp.Error.Message)
	}
	return rpcResp, err
}

func (a *aria2Cli) AddUri(dir string, links ...string) error {
	req := rpcRequest{
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
	_, err := a.doRPC(req)
	return err
}

func (a *aria2Cli) AddTorrent(dir string, fileData []byte) error {
	fileDataEncoded := base64.StdEncoding.EncodeToString(fileData)
	req := rpcRequest{
		Jsonrpc: "2.0",
		ID:      "TelegramBot",
		Method:  "aria2.addTorrent",
		Params: []interface{}{
			a.Token,
			fileDataEncoded,
			[]interface{}{},
			struct {
				Dir string `json:"dir"`
			}{dir},
		},
	}
	_, err := a.doRPC(req)
	return err
}
