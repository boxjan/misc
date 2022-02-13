package logger

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"io"
	"k8s.io/klog/v2"
	"strings"
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
	TagBlack      = "black"
	TagRed        = "red"
	TagGreen      = "green"
	TagYellow     = "yellow"
	TagBlue       = "blue"
	TagMagenta    = "magenta"
	TagCyan       = "cyan"
	TagWhite      = "white"
	TagReset      = "reset"
)

// Color values
const (
	cBlack   = "\u001b[90m"
	cRed     = "\u001b[91m"
	cGreen   = "\u001b[92m"
	cYellow  = "\u001b[93m"
	cBlue    = "\u001b[94m"
	cMagenta = "\u001b[95m"
	cCyan    = "\u001b[96m"
	cWhite   = "\u001b[97m"
	cReset   = "\u001b[0m"
)

type Config struct {
	Format string
}

var defaultConfig = Config{
	Format: "${status} - ${ip} ${ips} ${method} ${host} ${path} ${latency} ${ua} route:{} err:${error} [${bytesReceived}:${bytesSent}] ",
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
			case TagReferer:
				return buf.WriteString(c.Get(fiber.HeaderReferer))
			case TagProtocol:
				return buf.WriteString(c.Protocol())
			case TagPort:
				return buf.WriteString(c.Port())
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
			case TagStatus:
				return appendInt(buf, c.Response().StatusCode())
			case TagResBody:
				return buf.Write(c.Response().Body())
			case TagReqHeaders:
				reqHeaders := make([]string, 0)
				for k, v := range c.GetReqHeaders() {
					reqHeaders = append(reqHeaders, k+"="+v)
				}
				return buf.Write([]byte(strings.Join(reqHeaders, "&")))
			case TagQueryStringParams:
				return buf.WriteString(c.Request().URI().QueryArgs().String())
			case TagMethod:
				return buf.WriteString(c.Method())
			case TagBlack:
				return buf.WriteString(cBlack)
			case TagRed:
				return buf.WriteString(cRed)
			case TagGreen:
				return buf.WriteString(cGreen)
			case TagYellow:
				return buf.WriteString(cYellow)
			case TagBlue:
				return buf.WriteString(cBlue)
			case TagMagenta:
				return buf.WriteString(cMagenta)
			case TagCyan:
				return buf.WriteString(cCyan)
			case TagWhite:
				return buf.WriteString(cWhite)
			case TagReset:
				return buf.WriteString(cReset)
			case TagError:
				if chainErr != nil {
					return buf.WriteString(chainErr.Error())
				}
				return buf.WriteString("-")
			default:
				// Check if we have a value tag i.e.: "reqHeader:x-key"
				switch {
				case strings.HasPrefix(tag, TagReqHeader):
					return buf.WriteString(c.Get(tag[10:]))
				case strings.HasPrefix(tag, TagHeader):
					return buf.WriteString(c.Get(tag[7:]))
				case strings.HasPrefix(tag, TagRespHeader):
					return buf.WriteString(c.GetRespHeader(tag[11:]))
				case strings.HasPrefix(tag, TagQuery):
					return buf.WriteString(c.Query(tag[6:]))
				case strings.HasPrefix(tag, TagForm):
					return buf.WriteString(c.FormValue(tag[5:]))
				case strings.HasPrefix(tag, TagCookie):
					return buf.WriteString(c.Cookies(tag[7:]))
				case strings.HasPrefix(tag, TagLocals):
					switch v := c.Locals(tag[7:]).(type) {
					case []byte:
						return buf.Write(v)
					case string:
						return buf.WriteString(v)
					case nil:
						return 0, nil
					default:
						return buf.WriteString(fmt.Sprintf("%v", v))
					}
				}
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
