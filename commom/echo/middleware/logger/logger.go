package logger

import (
	"github.com/boxjan/misc/commom/echo/utils"
	"github.com/labstack/echo/v4"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasttemplate"
	"io"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
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

func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(defaultConfig)
}

//goland:noinspection GoNameStartsWithPackageName
func LoggerWithConfig(config Config) echo.MiddlewareFunc {
	if config.Format == "" {
		config.Format = defaultConfig.Format
	}

	// Create template parser
	tmpl := fasttemplate.New(config.Format, "${", "}")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			req := c.Request()
			res := c.Response()
			start := time.Now()
			chainErr := next(c)
			if chainErr != nil {
				c.Error(chainErr)
			}
			stop := time.Now()

			buf := bytebufferpool.Get()
			buf.Reset()
			defer bytebufferpool.Put(buf)

			_, err = tmpl.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case TagReferer:
					return buf.WriteString(req.Referer())
				case TagProtocol:
					return buf.WriteString(req.Proto)
				case TagPort:
					return buf.WriteString(strconv.Itoa(utils.Port(req)))
				case TagIP:
					return buf.WriteString(c.RealIP())
				case TagIPs:
					return buf.WriteString(req.Header.Get("X-Forwarded-For"))
				case TagHost:
					return buf.WriteString(req.Host)
				case TagPath:
					return buf.WriteString(c.Path())
				case TagURL:
					return buf.WriteString(req.URL.String())
				case TagUA:
					return buf.WriteString(req.UserAgent())
				case TagLatency:
					return buf.WriteString(stop.Sub(start).String())
				case TagBody:
					return buf.WriteString("[not support req body for echo]")
				case TagBytesReceived:
					return buf.WriteString(strconv.FormatInt(req.ContentLength, 10))
				case TagBytesSent:
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				case TagRoute:
					return buf.WriteString("[not support route match status for echo]")
				case TagStatus:
					return buf.WriteString(strconv.Itoa(c.Response().Status))
				case TagResBody:
					return buf.WriteString("[not support resp body for echo]")
				case TagReqHeaders:
					reqHeaders := make([]string, 0)
					for k, vs := range req.Header {
						for _, v := range vs {
							reqHeaders = append(reqHeaders, k+"="+v)
						}
					}
					return buf.Write([]byte(strings.Join(reqHeaders, "&")))
				case TagQueryStringParams:
					return buf.WriteString(req.URL.RawQuery)
				case TagMethod:
					return buf.WriteString(req.Method)
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
						return buf.WriteString(req.Header.Get(tag[10:]))
					case strings.HasPrefix(tag, TagHeader):
						return buf.WriteString(req.Header.Get(tag[7:]))
					case strings.HasPrefix(tag, TagRespHeader):
						return buf.WriteString(res.Header().Get(tag[11:]))
					case strings.HasPrefix(tag, TagQuery):
						return buf.WriteString(c.QueryParam(tag[6:]))
					case strings.HasPrefix(tag, TagForm):
						return buf.WriteString(c.FormValue(tag[5:]))
					case strings.HasPrefix(tag, TagCookie):
						cookie, err := req.Cookie(tag[7:])
						if err == nil {
							return buf.WriteString(cookie.String())
						} else {
							return 0, nil
						}
					case strings.HasPrefix(tag, TagLocals):
						return buf.WriteString("[not support local body for echo]")
					}
				}
				return 0, nil
			})
			klog.Info(buf.String())
			return nil
		}
	}
}
