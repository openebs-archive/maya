package server

// This is an adaptation of Hashicorp's Nomad library.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/ghodss/yaml"
	"github.com/openebs/mayaserver/lib/config"
	"github.com/ugorji/go/codec"
)

const (
	// ErrInvalidMethod is used if the HTTP method is not supported
	ErrInvalidMethod     = "Invalid method"
	ErrGetMethodRequired = "GET method required"
	ErrPutMethodRequired = "PUT/POST method required"
)

var (
	// jsonHandle and jsonHandlePretty are the codec handles to JSON encode
	// structs. The pretty handle will add indents for easier human consumption.
	jsonHandle       = &codec.JsonHandle{}
	jsonHandlePretty = &codec.JsonHandle{Indent: 4}
)

// HTTPServer is used to wrap maya api server and expose it over an HTTP interface
type HTTPServer struct {
	// TODO
	// Convert MayaApiServer as an interface with some public contracts
	// This interface can be embedded in HTTPServer struct
	maya *MayaApiServer

	mux      *http.ServeMux
	listener net.Listener
	logger   *log.Logger
	addr     string
}

// NewHTTPServer starts new HTTP server over Maya server
func NewHTTPServer(maya *MayaApiServer, config *config.MayaConfig, logOutput io.Writer) (*HTTPServer, error) {
	// Start the listener
	lnAddr, err := net.ResolveTCPAddr("tcp", config.NormalizedAddrs.HTTP)
	if err != nil {
		return nil, err
	}
	ln, err := config.Listener("tcp", lnAddr.IP.String(), lnAddr.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to start HTTP listener: %v", err)
	}

	// If TLS is enabled, wrap the listener with a TLS listener
	//if config.TLSConfig.EnableHTTP {
	//	tlsConf := &tlsutil.Config{
	//		VerifyIncoming:       false,
	//		VerifyOutgoing:       true,
	//		VerifyServerHostname: config.TLSConfig.VerifyServerHostname,
	//		CAFile:               config.TLSConfig.CAFile,
	//		CertFile:             config.TLSConfig.CertFile,
	//		KeyFile:              config.TLSConfig.KeyFile,
	//	}
	//	tlsConfig, err := tlsConf.IncomingTLSConfig()
	//	if err != nil {
	//		return nil, err
	//	}
	//	ln = tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, tlsConfig)
	//}

	// Create the mux
	mux := http.NewServeMux()

	// Create the server
	srv := &HTTPServer{
		maya:     maya,
		mux:      mux,
		listener: ln,
		logger:   maya.logger,
		addr:     ln.Addr().String(),
	}
	srv.registerHandlers(config.ServiceProvider, config.EnableDebug)

	// Start the server
	go http.Serve(ln, gziphandler.GzipHandler(mux))
	return srv, nil
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by NewHttpServer so
// dead TCP connections eventually go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(30 * time.Second)
	return tc, nil
}

// Shutdown is used to shutdown the HTTP server
func (s *HTTPServer) Shutdown() {
	if s != nil {
		s.logger.Printf("[DEBUG] http: Shutting down http server")
		s.listener.Close()
	}
}

// registerHandlers is used to attach handlers to the mux
func (s *HTTPServer) registerHandlers(serviceProvider string, enableDebug bool) {

	// NOTE - The curried func (due to wrap) is set as mux handler
	// NOTE - The original handler is passed as a func to the wrap method
	s.mux.HandleFunc("/latest/meta-data/", s.wrap(s.MetaSpecificRequest))

	// Request w.r.t to a single VSM entity is handled here
	s.mux.HandleFunc("/latest/volumes/", s.wrap(s.VSMSpecificRequest))
}

// HTTPCodedError is used to provide the HTTP error code
type HTTPCodedError interface {
	error
	Code() int
}

func CodedError(c int, s string) HTTPCodedError {
	return &codedError{s, c}
}

type codedError struct {
	s    string
	code int
}

func (e *codedError) Error() string {
	return e.s
}

func (e *codedError) Code() int {
	return e.code
}

// wrap is a convenient method used to wrap the handler function &
// return this handler curried with common logic.
func (s *HTTPServer) wrap(handler func(resp http.ResponseWriter, req *http.Request) (interface{}, error)) func(resp http.ResponseWriter, req *http.Request) {
	// curry the handler
	f := func(resp http.ResponseWriter, req *http.Request) {
		// some book keeping stuff
		setHeaders(resp, s.maya.config.HTTPAPIResponseHeaders)
		reqURL := req.URL.String()
		start := time.Now()
		defer func() {
			s.logger.Printf("[DEBUG] http: Request %v (%v)", reqURL, time.Now().Sub(start))
		}()

		s.logger.Printf("[DEBUG] http: Request %v (%v)", reqURL, req.Method)
		// Original handler is invoked
		obj, err := handler(resp, req)

		// Check for an error & set it as an http error
		// Below err block for re-usability
	HAS_ERR:
		if err != nil {
			s.logger.Printf("[ERR] http: Request %v %v, error: %v", req.Method, reqURL, err)
			code := 500
			if http, ok := err.(HTTPCodedError); ok {
				code = http.Code()
			}
			resp.WriteHeader(code)
			resp.Write([]byte(err.Error()))
			return
		}

		prettyPrint := false
		if v, ok := req.URL.Query()["pretty"]; ok {
			if len(v) > 0 && (len(v[0]) == 0 || v[0] != "0") {
				prettyPrint = true
			}
		}

		// Transform the response structure to its JSON equivalent
		if obj != nil {
			var buf bytes.Buffer
			if prettyPrint {
				enc := codec.NewEncoder(&buf, jsonHandlePretty)
				err = enc.Encode(obj)
				if err == nil {
					buf.Write([]byte("\n"))
				}
			} else {
				enc := codec.NewEncoder(&buf, jsonHandle)
				err = enc.Encode(obj)
			}

			// err is handled for both pretty & plain
			if err != nil {
				goto HAS_ERR
			}
			// no error, set the response as json
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(buf.Bytes())
		}
	}
	return f
}

// Get the value of Content-Type that is set in http request header
func getContentType(req *http.Request) (string, error) {

	if req.Header == nil {
		return "", fmt.Errorf("Request does not have any header")
	}

	return req.Header.Get("Content-Type"), nil
}

// Decode the request body to appropriate structure based on content
// type
func decodeBody(req *http.Request, out interface{}) error {

	cType, err := getContentType(req)
	if err != nil {
		return err
	}

	if strings.Contains(cType, "yaml") {
		return decodeYamlBody(req, out)
	}

	// default is assumed to be json content
	return decodeJsonBody(req, out)
}

// decodeJsonBody is used to decode a JSON request body
func decodeJsonBody(req *http.Request, out interface{}) error {
	dec := json.NewDecoder(req.Body)
	return dec.Decode(&out)
}

// decodeYamlBody is used to decode a YAML request body
func decodeYamlBody(req *http.Request, out interface{}) error {
	// Get []bytes from io.Reader
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &out)
}

// setIndex is used to set the index response header
func setIndex(resp http.ResponseWriter, index uint64) {
	resp.Header().Set("X-Maya-Index", strconv.FormatUint(index, 10))
}

// setLastContact is used to set the last contact header
func setLastContact(resp http.ResponseWriter, last time.Duration) {
	lastMsec := uint64(last / time.Millisecond)
	resp.Header().Set("X-Maya-LastContact", strconv.FormatUint(lastMsec, 10))
}

// setMeta is used to set the query response meta data
//func setMeta(resp http.ResponseWriter, qm *structs.QueryMeta) {
//setIndex(resp, qm.Index)
//setLastContact(resp, qm.LastContact)
//}

// setHeaders is used to set canonical response header fields
func setHeaders(resp http.ResponseWriter, headers map[string]string) {
	for field, value := range headers {
		resp.Header().Set(http.CanonicalHeaderKey(field), value)
	}
}

// parseWait is used to parse the ?wait and ?index query params
// Returns true on error
//func parseWait(resp http.ResponseWriter, req *http.Request, qo *structs.QueryOptions) bool {
//	query := req.URL.Query()
//	if wait := query.Get("wait"); wait != "" {
//		dur, err := time.ParseDuration(wait)
//		if err != nil {
//			resp.WriteHeader(400)
//			resp.Write([]byte("Invalid wait time"))
//			return true
//		}
//		qo.MaxQueryTime = dur
//	}
//	if idx := query.Get("index"); idx != "" {
//		index, err := strconv.ParseUint(idx, 10, 64)
//		if err != nil {
//			resp.WriteHeader(400)
//			resp.Write([]byte("Invalid index"))
//			return true
//		}
//		qo.MinQueryIndex = index
//	}
//	return false
//}

// parseConsistency is used to parse the ?stale query params.
//func parseConsistency(req *http.Request, qo *structs.QueryOptions) {
//	query := req.URL.Query()
//	if _, ok := query["stale"]; ok {
//		qo.AllowStale = true
//	}
//}

// parsePrefix is used to parse the ?prefix query param
//func parsePrefix(req *http.Request, qo *structs.QueryOptions) {
//	query := req.URL.Query()
//	if prefix := query.Get("prefix"); prefix != "" {
//		qo.Prefix = prefix
//	}
//}

// parseRegion is used to parse the ?region query param
func (s *HTTPServer) parseRegion(req *http.Request, r *string) {
	if other := req.URL.Query().Get("region"); other != "" {
		*r = other
	} else if *r == "" {
		*r = s.maya.config.Region
	}
}

// parse is a convenience method for endpoints that need to parse multiple flags
//func (s *HTTPServer) parse(resp http.ResponseWriter, req *http.Request, r *string, qo *structs.QueryOptions) bool {
//s.parseRegion(req, r)
//parseConsistency(req, qo)
//parsePrefix(req, qo)
//return parseWait(resp, req, qo)
//}
