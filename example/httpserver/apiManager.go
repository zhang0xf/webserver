package main

import (
	"fmt"
	"net/http"
	"nw/common/httpserver"
	"nw/internel/util"
	"runtime/debug"
	"strconv"
	"strings"
)

const ResponseFomat = `{"code":%d,"message":"%s"}`

type ApiManager struct {
	// util.DefaultModule // 略（用于模块管理）
	server *httpserver.HttpServer
}

func NewApiManager() *ApiManager {
	return &ApiManager{server: httpserver.NewHttpServer()}
}

func (apiManager *ApiManager) Init() error {
	apiManager.server.Router.Handle("/api/whNotice", http.HandlerFunc(apiManager.whNotice))
	apiManager.server.Router.Handle("/api/banAccount", http.HandlerFunc(apiManager.banAccount))
	return nil
}

func (apiManager *ApiManager) Start() error {
	if err := apiManager.server.Start(":8505"); err != nil {
		return err
	}
	return nil
}

func (apiManager *ApiManager) Run() { util.WaitForTerminate() }

func (apiManager *ApiManager) Stop() { apiManager.server.Stop() }

func (apiManager *ApiManager) whNotice(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			stackBytes := debug.Stack()
			fmt.Printf("panic whNotice : %v, %s\n", r, stackBytes)
		}
	}()

	// POST方法
	var whNotice = strings.TrimSpace(r.PostFormValue("info"))

	switch whNotice {
	case "":
		w.Write([]byte(fmt.Sprintf(ResponseFomat, 0, "发布公告失败！")))
		return
	default:
		w.Write([]byte(fmt.Sprintf(ResponseFomat, 1, "发布公告成功！")))
	}

	//发送所有在线玩家 弹窗效果
	//m.ClientManager.BroadcastAll(&pb.WeihuNoticeNtf{Content: whNotice})
}

func (apiManager *ApiManager) banAccount(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			stackBytes := debug.Stack()
			fmt.Printf("panic banAccount : %v, %s\n", r, stackBytes)
		}
	}()

	var openId = strings.TrimSpace(r.PostFormValue("openId"))
	var banTimeStr = strings.TrimSpace(r.PostFormValue("bantime"))
	bantime, err := strconv.Atoi(banTimeStr)
	if err != nil {
		bantime = 3600
	}
	var reason = strings.TrimSpace(r.PostFormValue("reason"))
	fmt.Printf("api manager banAccount, openId : %v, banTime : %v, reason : %v", openId, bantime, reason)
	// m.Gm.BanAccount(openId, bantime, reason)
}
