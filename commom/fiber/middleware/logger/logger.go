package logger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"io"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

const (
	TagReferer           = "referer"
	TagProtocol          = "protocol"
	TagPort              = "port"
	TagIP                = "ip"
	TagIPs               = "ips"
	TagHost              = "host"
	TagMethod            = "method"
	TagPath              = "path"
	TagURL               = "url"
	TagUA                = "ua"
	TagLatency           = "latency"
	TagStatus            = "status"
	TagResBody           = "resBody"
	TagReqHeaders        = "reqHeaders"
	TagQueryStringParams = "queryParams"
	TagBody              = "body"
	TagBytesSent         = "bytesSent"
	TagBytesReceived     = "bytesReceived"
	TagRoute             = "route"
	TagError             = "error"
	// DEPRECATED: Use TagReqHeader instead
	TagHeader     = "header:"
	TagReqHeader  = "reqHeader:"
	TagRespHeader = "respHeader:"
	TagLocals     = "locals:"
	TagQuery      = "query:"
	TagForm       = "form:"
	TagCookie     = "cookie:"
)

type Config struct {
	Format string
}

var defaultConfig = Config{
	Format: "${status} - ${ip} ${ips} ${latency} ${method} ${host} ${path} [${bytes_in}:${bytes_out}] ",
}

func Logger() fiber.Handler {
	return LoggerWithConfig(defaultConfig)
}

//goland:noinspection GoNameStartsWithPackageName
func LoggerWithConfig(config Config) fiber.Handler {
	if config.Format == "" {
		config.Format = defaultConfig.Format
	}

	// Create template parser
	tmpl := fasttemplate.New(config.Format, "${", "}")

	// Set variables
	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	return func(c *fiber.Ctx) error {
		once.Do(func() {
			errHandler = c.App().ErrorHandler
		})

		start := time.Now()

		// Handle request, store err for logging
		chainErr := c.Next()

		// Manually call error handler
		if chainErr != nil {
			if err := errHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		stop := time.Now()

		buf := bytebufferpool.Get()
		buf.Reset()
		defer bytebufferpool.Put(buf)

		_, err := tmpl.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case TagIP:
				return buf.WriteString(c.IP())
			case TagIPs:
				return buf.WriteString(c.Get(fiber.HeaderXForwardedFor))
			case TagHost:
				return buf.WriteString(c.Hostname())
			case TagPath:
				return buf.WriteString(c.Path())
			case TagURL:
				return buf.WriteString(c.OriginalURL())
			case TagUA:
				return buf.WriteString(c.Get(fiber.HeaderUserAgent))
			case TagLatency:
				return buf.WriteString(stop.Sub(start).String())
			case TagBody:
				return buf.Write(c.Body())
			case TagBytesReceived:
				return appendInt(buf, len(c.Request().Body()))
			case TagBytesSent:
				return appendInt(buf, len(c.Response().Body()))
			case TagRoute:
				return buf.WriteString(c.Route().Path)
			}
			return 0, nil
		})

		// Also write errors to the buffer
		if err != nil {
			_, _ = buf.WriteString(err.Error())
		}

		klog.Info(buf.String())
		return nil
	}
}

func appendInt(buf *bytebufferpool.ByteBuffer, v int) (int, error) {
	old := len(buf.B)
	buf.B = fasthttp.AppendUint(buf.B, v)
	return len(buf.B) - old, nil
}
