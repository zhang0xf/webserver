package main

import (
	"fmt"
	"net/url"
	"nw/common/httpclient"
)

const (
	WATCH_URL = "http://watcher.XXXXXXX.com/"
)

type WatchResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type WatchManager struct {
	// util.DefaultModule // 略（用于模块管理）
	token string
}

func NewWatchManager() *WatchManager {
	return &WatchManager{token: ""}
}

// 为*WatchManager实现Moudle（略）接口，使其成为Moudle类型
func (watchManager *WatchManager) Init() error {
	if watchManager.token != "" {
		return nil
	}

	// http://watcher.hoodinn.com/api/sign?project=xxkx&system=linux&service=game
	// {"code":1,"message":"\u6ce8\u518c\u6210\u529f","data":{"token":"cb209fc97d785bf9bcecb75e98c5481b5c448e12"}}

	watchManager.getToken()
	watchManager.SyncStatus()
	return nil
}

func (watchManager *WatchManager) Start() error { return nil }

func (watchManager *WatchManager) Run() {}

func (watchManager *WatchManager) Stop() {}

// 为*WatchManager实现其他一些接口，处理本模块业务逻辑。
func (watchManager *WatchManager) getToken() {
	params := url.Values{
		"project": {"某某游戏"},
		"system":  {"wx"},
		"service": {"GameServer" + "1服" + "登录并创建角色"},
	}

	var result WatchResult
	err := httpclient.DoGetEx(WATCH_URL+"api/sign", params, &result) // 发送http的Get请求到另一台server
	if err != nil {
		fmt.Printf("httpclient DoGetEx error : %v\n", err)
		return
	}

	if result.Code != 1 {
		fmt.Printf("watch getToken fail, code: %d, message: %v\n", result.Code, result.Message)
		return
	}

	watchManager.token = result.Data.Token
}

// 同步状态信息
func (watchManager *WatchManager) SyncStatus() error {
	// 暂略
	return nil
}
