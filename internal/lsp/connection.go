// The implementation of logMessages, is a heavily modified version of the original implementation
// in https://github.com/sourcegraph/jsonrpc2
// The original license for that code is as follows:
// Copyright (c) 2016 Sourcegraph Inc
//
// # MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions
// of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
// THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

type ConnectionOptions struct {
	LoggingConfig ConnectionLoggingConfig
}

type ConnectionLoggingConfig struct {
	Writer io.Writer

	LogInbound  bool
	LogOutbound bool

	// IncludeMethods is a list of methods to include in the request log.
	// If empty, all methods are included. IncludeMethods takes precedence
	// over ExcludeMethods.
	IncludeMethods []string
	// ExcludeMethods is a list of methods to exclude from the request log.
	ExcludeMethods []string
}

func (cfg *ConnectionLoggingConfig) ShouldLog(method string) bool {
	if len(cfg.IncludeMethods) > 0 {
		for _, m := range cfg.IncludeMethods {
			if m == method {
				return true
			}
		}

		return false
	}

	for _, m := range cfg.ExcludeMethods {
		if m == method {
			return false
		}
	}

	return true
}

type ConnectionHandlerFunc func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request) (result interface{}, err error)

type connectionLogger struct {
	writer io.Writer
}

func (c *connectionLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(c.writer, format, v...)
}

func NewConnectionFromLanguageServer(
	ctx context.Context,
	handler ConnectionHandlerFunc,
	opts *ConnectionOptions,
) *jsonrpc2.Conn {
	return jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(StdOutReadWriteCloser{}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(handler),
		logMessages(opts.LoggingConfig),
	)
}

func logMessages(cfg ConnectionLoggingConfig) jsonrpc2.ConnOpt {
	logger := &connectionLogger{writer: cfg.Writer}

	return func(c *jsonrpc2.Conn) {
		// Remember reqs we have received so that we can helpfully show the
		// request method in OnSend for responses.
		var (
			mu         sync.Mutex
			reqMethods = map[jsonrpc2.ID]string{}
		)

		setMethod := func(id jsonrpc2.ID, method string) {
			mu.Lock()
			defer mu.Unlock()

			reqMethods[id] = method
		}

		getMethod := func(id jsonrpc2.ID) string {
			mu.Lock()
			defer mu.Unlock()

			return reqMethods[id]
		}

		deleteMethod := func(id jsonrpc2.ID) {
			mu.Lock()
			defer mu.Unlock()

			delete(reqMethods, id)
		}

		if cfg.LogInbound {
			jsonrpc2.OnRecv(buildRecvHandler(setMethod, logger, cfg))(c)
		}

		if cfg.LogOutbound {
			jsonrpc2.OnSend(buildSendHandler(getMethod, deleteMethod, logger, cfg))(c)
		}
	}
}

func buildRecvHandler(
	setMethod func(jsonrpc2.ID, string),
	logger *connectionLogger,
	cfg ConnectionLoggingConfig,
) func(req *jsonrpc2.Request, resp *jsonrpc2.Response) {
	return func(req *jsonrpc2.Request, resp *jsonrpc2.Response) {
		switch {
		case req != nil && resp == nil:
			setMethod(req.ID, req.Method)

			if !cfg.ShouldLog(req.Method) {
				return
			}

			params, _ := json.Marshal(req.Params)
			if req.Notif {
				logger.Printf("--> notif: %s: %s\n", req.Method, params)
			} else {
				logger.Printf("--> request #%s: %s: %s\n", req.ID, req.Method, params)
			}

		case resp != nil:
			var method string
			if req != nil {
				method = req.Method
			} else {
				method = "(no matching request)"
			}

			if !cfg.ShouldLog(method) {
				return
			}

			switch {
			case resp.Result != nil:
				result, _ := json.Marshal(resp.Result)
				logger.Printf("--> response #%s: %s: %s\n", resp.ID, method, result)
			case resp.Error != nil:
				errBs, _ := json.Marshal(resp.Error)
				logger.Printf("--> response error #%s: %s: %s\n", resp.ID, method, errBs)
			}
		}
	}
}

func buildSendHandler(
	getMethod func(jsonrpc2.ID) string,
	deleteMethod func(jsonrpc2.ID),
	logger *connectionLogger,
	cfg ConnectionLoggingConfig,
) func(req *jsonrpc2.Request, resp *jsonrpc2.Response) {
	return func(req *jsonrpc2.Request, resp *jsonrpc2.Response) {
		switch {
		case req != nil && resp == nil:
			if !cfg.ShouldLog(req.Method) {
				return
			}

			params, _ := json.Marshal(req.Params)
			if req.Notif {
				logger.Printf("<-- notif: %s: %s\n", req.Method, params)
			} else {
				logger.Printf("<-- request #%s: %s: %s\n", req.ID, req.Method, params)
			}

		case resp != nil:
			method := getMethod(resp.ID)

			deleteMethod(resp.ID)

			if method == "" {
				method = "(no previous request)"
			}

			if !cfg.ShouldLog(method) {
				return
			}

			if resp.Result != nil {
				result, _ := json.Marshal(resp.Result)
				logger.Printf("<-- response #%s: %s: %s\n", resp.ID, method, result)
			} else {
				errBs, _ := json.Marshal(resp.Error)
				logger.Printf("<-- response error #%s: %s: %s\n", resp.ID, method, errBs)
			}
		}
	}
}
