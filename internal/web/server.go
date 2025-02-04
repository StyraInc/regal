package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/arl/statsviz"

	"github.com/styrainc/regal/internal/explorer"
	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/util"

	_ "net/http/pprof" //nolint:gosec
)

const mainTemplate = "main.tpl"

//go:embed assets/*
var assets embed.FS

//nolint:gochecknoglobals
var tpl = template.Must(template.New("main.tpl").ParseFS(assets, "assets/main.tpl"))

type Server struct {
	cache        *cache.Cache
	logWriter    io.Writer
	logLevel     log.Level
	workspaceURI string
	baseURL      string
	client       clients.Identifier
}

func NewServer(cache *cache.Cache, logWriter io.Writer, level log.Level) *Server {
	return &Server{cache: cache, logWriter: logWriter, logLevel: level}
}

func (s *Server) SetWorkspaceURI(uri string) {
	s.workspaceURI = uri
}

func (s *Server) SetClient(client clients.Identifier) {
	s.client = client
}

func (s *Server) GetBaseURL() string {
	return s.baseURL
}

type state struct {
	Code   string
	Result []stringResult
}

type stringResult struct {
	Class  string
	Stage  string
	Output string
	Show   bool
}

func (s *Server) Start(_ context.Context) {
	mux := http.NewServeMux()

	err := statsviz.Register(mux)
	if err != nil {
		s.log(log.LevelMessage, fmt.Sprintf("failed to register statsviz handler: %v", err))
	}

	mux.HandleFunc("GET /explorer/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/explorer/")
		policyURI := s.workspaceURI + "/" + path

		policy, ok := s.cache.GetFileContents(policyURI)
		if !ok {
			http.NotFound(w, r)

			return
		}

		var enableStrict, enableAnnotationProcessing, enablePrint bool

		if err := r.ParseForm(); err == nil {
			enableStrict = r.Form.Get("strict") == "on"
			enableAnnotationProcessing = r.Form.Get("annotations") == "on"
			enablePrint = r.Form.Get("print") == "on"
		}

		cs := explorer.CompilerStages(path, policy, enableStrict, enableAnnotationProcessing, enablePrint)
		st := state{
			Code: policy,
		}
		st.Result = make([]stringResult, len(cs)+1)

		for i := range cs {
			st.Result[i].Stage = cs[i].Stage
			if cs[i].Error != "" {
				st.Result[i].Output = cs[i].Error
				st.Result[i].Class = "bad"
			} else {
				st.Result[i].Output = cs[i].Result.String()
			}

			st.Result[i].Show = i == 0 || st.Result[i-1].Output != st.Result[i].Output
			if st.Result[i].Class == "" {
				if st.Result[i].Show {
					st.Result[i].Class = "ok"
				} else {
					st.Result[i].Class = "plain"
				}
			}
		}

		n := len(cs)

		st.Result[n].Stage = "Plan"
		st.Result[n].Show = true

		p, err := explorer.Plan(r.Context(), path, policy, enablePrint)
		if err != nil {
			st.Result[n].Class = "bad"
			st.Result[n].Output = err.Error()
		} else {
			st.Result[n].Class = "ok"
			st.Result[n].Output = p
		}

		if err := tpl.ExecuteTemplate(w, mainTemplate, st); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle("/assets/", http.FileServer(http.FS(assets)))

	// pprof handlers
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)

	// root handler for those looking for what the server is
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`
<h1>Regal Language Server</h1>
<ul>
<li><a href="/debug/pprof/">pprof</a></li>
<li><a href="/debug/statsviz">statsviz</a></li>
</ul>`))
		if err != nil {
			s.log(log.LevelMessage, fmt.Sprintf("failed to write response: %v", err))
		}
	})

	freePort, err := util.FreePort(5052, 5053, 5054)
	if err != nil {
		s.log(log.LevelMessage, "preferred web server ports are not available, using random port")

		freePort = 0
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", freePort))
	if err != nil {
		s.log(log.LevelMessage, fmt.Sprintf("failed to start web server: %v", err))

		return
	}

	s.baseURL = "http://" + listener.Addr().String()

	s.log(log.LevelMessage, "starting web server for docs on "+s.baseURL)

	//nolint:gosec // this is a local server, no timeouts needed
	err = http.Serve(listener, mux)
	if err != nil {
		s.log(log.LevelMessage, fmt.Sprintf("failed to serve web server: %v", err))
	}
}

//nolint:unparam
func (s *Server) log(level log.Level, message string) {
	if !s.logLevel.ShouldLog(level) {
		return
	}

	if s.logWriter != nil {
		fmt.Fprintln(s.logWriter, message)
	}
}
