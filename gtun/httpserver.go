package gtun

import (
	"encoding/json"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/gin-gonic/gin"
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
	srv := gin.Default()
	srv.GET("/meta", loadMeta)
	srv.POST("/ip/add", addIP)
	srv.DELETE("/ip/delete", delIP)
	srv.GET("/ip/list/:region", listIP)
	return http.ListenAndServe(s.listenAddr, nil)
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func loadMeta(ctx *gin.Context) {
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
	ctx.JSON(http.StatusOK, body)
}

func addIP(ctx *gin.Context) {
	type req struct {
		Region string `json:"region"`
		IP     string `json:"ip"`
	}

	var form = req{}
	err := bindForm(ctx, &form)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// add to ipset
	err = proxy.GetManager().AddIP(form.Region, form.IP)
	if err != nil {
		ctx.JSON(http.StatusOK, &response{
			Code:    -1,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// write to proxy file
	ctx.JSON(http.StatusOK, &response{Code: 0, Message: "success"})
}

func delIP(ctx *gin.Context) {
	type req struct {
		Region string `json:"region"`
		IP     string `json:"ip"`
	}

	var form = req{}
	err := bindForm(ctx, &form)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete from ipset
	err = proxy.GetManager().DelIP(form.Region, form.IP)
	if err != nil {
		ctx.JSON(http.StatusOK, &response{
			Code:    -1,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, &response{Code: 0, Message: "success"})
}

func listIP(ctx *gin.Context) {
	region := ctx.Param("region")
	ips := proxy.GetManager().IPList(region)
	ctx.JSON(http.StatusOK, &response{Code: 0, Message: "success", Data: ips})
}

func bindForm(ctx *gin.Context, obj interface{}) error {
	cnt, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(cnt, obj)
}

func reply(ctx *gin.Context, obj interface{}) {
	ctx.JSON(http.StatusOK, obj)
}
