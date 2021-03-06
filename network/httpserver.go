package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/duanhf2012/origin/service"
	"github.com/duanhf2012/origin/sysmodule"
)

type CA struct {
	certfile string
	keyfile  string
}

type HttpServer struct {
	port uint16

	handler      http.Handler
	readtimeout  time.Duration
	writetimeout time.Duration

	httpserver *http.Server
	caList     []CA

	ishttps bool
}

func (slf *HttpServer) Init(port uint16, handler http.Handler, readtimeout time.Duration, writetimeout time.Duration) {
	slf.port = port
	slf.handler = handler
	slf.readtimeout = readtimeout
	slf.writetimeout = writetimeout
}

func (slf *HttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (slf *HttpServer) Start() {
	go slf.startListen()
}

func (slf *HttpServer) startListen() error {
	listenPort := fmt.Sprintf(":%d", slf.port)

	var tlscatList []tls.Certificate
	var tlsConfig *tls.Config
	for _, cadata := range slf.caList {
		cer, err := tls.LoadX509KeyPair(cadata.certfile, cadata.keyfile)
		if err != nil {
			service.GetLogger().Printf(sysmodule.LEVER_FATAL, "load CA  [%s]-[%s] file is error :%s", cadata.certfile, cadata.keyfile, err.Error())
			os.Exit(1)
			return nil
		}
		tlscatList = append(tlscatList, cer)
	}

	if len(tlscatList) > 0 {
		tlsConfig = &tls.Config{Certificates: tlscatList}
	}

	slf.httpserver = &http.Server{
		Addr:           listenPort,
		Handler:        slf.handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,
	}

	var err error
	if slf.ishttps == true {
		err = slf.httpserver.ListenAndServeTLS("", "")
	} else {
		err = slf.httpserver.ListenAndServe()
	}

	if err != nil {
		service.GetLogger().Printf(sysmodule.LEVER_FATAL, "http.ListenAndServe(%d, nil) error:%v\n", listenPort, err)
		fmt.Printf("http.ListenAndServe(%d, %v) error\n", slf.port, err)
		os.Exit(1)
	}

	return nil
}

func (slf *HttpServer) SetHttps(certfile string, keyfile string) bool {
	if certfile == "" || keyfile == "" {
		return false
	}
	slf.caList = append(slf.caList, CA{certfile, keyfile})
	slf.ishttps = true
	return true
}
