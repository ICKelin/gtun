package gtun

import (
	"encoding/json"
	"github.com/ICKelin/gtun/gtun/proxy"
	"io"
	"net/http"
)

type HTTPServer struct {
	listenAddr string
}

func NewHTTPServer(listenAddr string) *HTTPServer {
	return &HTTPServer{listenAddr: listenAddr}
}

func (s *HTTPServer) ListenAndServe() error {
	http.HandleFunc("/meta", loadMeta)
	http.HandleFunc("/ip/add", addIP)
	http.HandleFunc("/ip/delete", delIP)
	return http.ListenAndServe(s.listenAddr, nil)
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func loadMeta(w http.ResponseWriter, r *http.Request) {
	regionList := make([]string, 0)
	regions := GetConfig().Settings
	for region, _ := range regions {
		regionList = append(regionList, region)
	}

	type replyBody struct {
		Regions []string `json:"regions"`
		Cfg     *Config
	}

	body := &replyBody{
		Regions: regionList,
		Cfg:     GetConfig(),
	}
	reply(w, body)
}

func addIP(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Region string `json:"region"`
		IP     string `json:"ip"`
	}

	var form = req{}
	err := bindForm(r, &form)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// add to ipset
	err = proxy.AddIP(form.Region, form.IP)
	if err != nil {
		reply(w, &response{
			Code:    -1,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	reply(w, &response{Code: 0, Message: "success"})
}

func delIP(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Region string `json:"region"`
		IP     string `json:"ip"`
	}

	var form = req{}
	err := bindForm(r, &form)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete from ipset
	err = proxy.DelIP(form.Region, form.IP)
	if err != nil {
		reply(w, &response{
			Code:    -1,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	reply(w, &response{Code: 0, Message: "success"})
}

func bindForm(r *http.Request, obj interface{}) error {
	cnt, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(cnt, obj)
}

func reply(w http.ResponseWriter, obj interface{}) {
	buf, _ := json.Marshal(obj)
	_, _ = w.Write(buf)
}
