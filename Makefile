WORKDIR = .

HTTP_CLIENT=$(WORKDIR)/example/httpclient/
HTTP_SERVER=$(WORKDIR)/example/httpserver/
TCP_CLIENT=$(WORKDIR)/example/tcpclient/
TCP_SERVER=$(WORKDIR)/example/tcpserver/
WS_CLIENT=$(WORKDIR)/example/wsclient/
WS_SERVER=$(WORKDIR)/example/wsserver/

default: all

all: httpclient httpserver tcpclient tcpserver wsclient wsserver

httpclient:
	@go build -o $(WORKDIR)/bin/httpclient $(HTTP_CLIENT)/*.go >/dev/null;

httpserver:
	@go build -o $(WORKDIR)/bin/httpserver $(HTTP_SERVER)/*.go >/dev/null;

tcpclient:
	@go build -o $(WORKDIR)/bin/tcpclient $(TCP_CLIENT)/*.go >/dev/null;

tcpserver:
	@go build -o $(WORKDIR)/bin/tcpserver $(TCP_SERVER)/*.go >/dev/null;

wsclient:
	@go build -o $(WORKDIR)/bin/wsclient $(WS_CLIENT)/*.go >/dev/null;

wsserver:
	@go build -o $(WORKDIR)/bin/wsserver $(WS_SERVER)/*.go >/dev/null;

.PHONY:clean

clean:
	rm ./bin/httpclient ./bin/httpserver ./bin/tcpclient ./bin/tcpserver \
	./bin/wsclient ./bin/wsserver