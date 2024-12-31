package caddy_analytics

import (
	"fmt"
	"io"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"golang.org/x/text/transform"
)

var SupportedProviders = []string{"plausible"}

type Analytics struct {
	// The analytics provider to use. Currently available: plausible
	Provider string `json:"provider,omitempty"`

	// Domain name on which the analytics provider is hosted. Defaults to the official instance.
	Server string `json:"host,omitempty"`

	// An API token. Used to automate registering of websites on the analytics provider. Does nothing for now.
	AdminToken string `json:"admin_token,omitempty"`
}

func init() {
	caddy.RegisterModule(Analytics{})
	httpcaddyfile.RegisterHandlerDirective("analytics", parseCaddyfileHandler)
}

func defaultConfig() *Analytics {
	return &Analytics{
		Server: Analytics{Provider: "plausible"}.DefaultServer(),
	}
}

func (a Analytics) DefaultServer() string {
	switch a.Provider {
	case "plausible":
		return "plausible.io"
	}
	return ""
}

func (Analytics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.analytics",
		New: func() caddy.Module {
			return defaultConfig()
		},
	}
}

func (a *Analytics) Validate() error {
	if a.Provider == "" {
		return fmt.Errorf("provider must be set")
	}
	if !stringInSlice(a.Provider, SupportedProviders) {
		return fmt.Errorf("provider must be one of %v", SupportedProviders)
	}

	return nil
}

func (a *Analytics) Provision(ctx caddy.Context) error {
	if a.Server == "" {
		a.Server = a.DefaultServer()
	}
	return nil
}

var _ caddy.Validator = (*Analytics)(nil)
var _ caddyhttp.MiddlewareHandler = (*Analytics)(nil)
var _ caddy.Provisioner = (*Analytics)(nil)

type analyticsWriter struct {
	*caddyhttp.ResponseWriterWrapper
	wroteHeader             bool
	transformer             transform.Transformer
	additionalContentLength int
	combinedWriter          io.WriteCloser
}

func (fw *analyticsWriter) WriteHeader(status int) {
	if fw.wroteHeader {
		return
	}
	fw.wroteHeader = true
	fw.combinedWriter = transform.NewWriter(fw.ResponseWriterWrapper, fw.transformer)

	// FIXME
	// oldLength, err := strconv.ParseInt(fw.Header().Get("Content-Length"), 10, 32)
	// fmt.Println(fw.Header().Get("Content-Length"))
	// if err == nil && oldLength != 0 {
	// 	fmt.Printf("oldLength: %d\n", oldLength)
	// 	fmt.Printf("additionalContentLength: %d\n", fw.additionalContentLength)
	// 	fw.Header().Set("Content-Length", fmt.Sprintf("%d", oldLength+int64(fw.additionalContentLength)))
	// }

	fw.Header().Del("Content-Length")

	fw.ResponseWriterWrapper.WriteHeader(status)
}

func (fw *analyticsWriter) Write(d []byte) (int, error) {
	if !fw.wroteHeader {
		fw.WriteHeader(http.StatusOK)
	}

	if fw.combinedWriter == nil {
		return fw.ResponseWriterWrapper.Write(d)
	}

	return fw.combinedWriter.Write(d)
}

func (fw *analyticsWriter) Close() error {

	return nil
}

func (a *Analytics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	fw := &analyticsWriter{
		ResponseWriterWrapper:   &caddyhttp.ResponseWriterWrapper{ResponseWriter: w},
		transformer:             firstStringReplacer("<head>", fmt.Sprintf(`<head>%s`, a.additionalContent(r))),
		additionalContentLength: len(a.additionalContent(r)),
	}

	err := next.ServeHTTP(fw, r)
	if err != nil {
		return err
	}

	fw.Close()
	return nil
}

// additionalContent returns the content to be added to the head of the HTML document
func (a *Analytics) additionalContent(r *http.Request) string {
	switch a.Provider {
	case "plausible":
		return fmt.Sprintf(`<script defer data-domain="%s" src="https://%s/js/script.js"></script>`, r.Host, a.Server)
	}
	return ""
}
