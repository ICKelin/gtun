package gtun

import (
	"encoding/json"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type HTTPServer struct {
	listenAddr string
}

func NewHTTPServer(listenAddr string) *HTTPServer {
	return &HTTPServer{listenAddr: listenAddr}
}

func (s *HTTPServer) ListenAndServe() error {
	srv := gin.Default()
	srv.POST("/sys/init", initSys)
	srv.POST("/sys/restart", restartSys)
	srv.GET("/meta", midInit(), loadMeta)

	srv.POST("/ip/add", midInit(), addIP)
	srv.DELETE("/ip/delete", midInit(), delIP)
	srv.GET("/ip/list", midInit(), listIP)

	srv.POST("/region/update", midInit(), updateRegionFile)

	srv.POST("/upload", uploadFile)

	srv.StaticFS("/", http.Dir("./web"))
	return srv.Run(s.listenAddr)
}

func midInit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if needSysInit {
			ctx.JSON(http.StatusForbidden, &response{Code: -1, Message: "need system initialize"})
			ctx.Abort()
		}
	}
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func initSys(ctx *gin.Context) {
	cnt, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logs.Error("read request fail: %v", err)
		reply(ctx, nil, err)
		return
	}

	fp, err := os.Open(confPath)
	if err != nil {
		logs.Error("write file fail: %v", err)
		reply(ctx, nil, err)
		return
	}
	defer fp.Close()

	_, err = fp.Write(cnt)
	if err != nil {
		logs.Error("write to file fail: %v", err)
		reply(ctx, nil, err)
		return
	}

	reply(ctx, nil, nil)
}

func restartSys(ctx *gin.Context) {
	// TODO: reload
	os.Exit(0)
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
	reply(ctx, body, nil)
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
		reply(ctx, nil, err)
		return
	}

	// write to proxy file
	reply(ctx, nil, nil)
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
		reply(ctx, nil, err)
		return
	}

	reply(ctx, nil, nil)
}

func listIP(ctx *gin.Context) {
	region := ctx.Query("region")
	ips := proxy.GetManager().IPList(region)
	reply(ctx, ips, nil)
}

func updateRegionFile(ctx *gin.Context) {
	type req struct {
		Region   string `json:"region"`
		Filename string `json:"filename"`
	}

	var form = req{}
	err := bindForm(ctx, &form)
	if err != nil {
		reply(ctx, nil, err)
		return
	}

	err = proxy.GetManager().UpdateRegionProxyFile(form.Region, form.Filename)
	reply(ctx, nil, err)
}

func uploadFile(ctx *gin.Context) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		reply(ctx, nil, err)
		return
	}

	err = ctx.SaveUploadedFile(fh, fh.Filename)
	if err != nil {
		reply(ctx, nil, err)
		return
	}

	reply(ctx, fh.Filename, nil)
}

func bindForm(ctx *gin.Context, obj interface{}) error {
	cnt, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(cnt, obj)
}

func reply(ctx *gin.Context, data interface{}, err error) {
	var code = 0
	var message = "success"
	if err != nil {
		code = -1
		message = err.Error()
		data = nil
	}

	ctx.JSON(http.StatusOK, &response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
