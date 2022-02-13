package route_list

import (
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
	"path"
	"reflect"
	"runtime"
	"strings"
)

type Route struct {
	Pos         uint32
	Use         bool
	Star        bool
	Root        bool
	Method      string
	Name        string
	Path        string
	Params      []string
	HandlerName []string
}

func (r Route) Key() string {
	return ""
}

func RouteList(appPtr *fiber.App) (uint64, []*Route) {
	appPtr.Handler()

	app := reflect.ValueOf(appPtr).Elem()

	routeCount := app.FieldByName("handlersCount").Uint()

	treeStack := app.FieldByName("treeStack")
	treeStackLen := treeStack.Len()

	for i := 0; i < treeStackLen; i++ {
		methodRoutesMap := treeStack.Index(i)

		prefixRoutesIt := methodRoutesMap.MapRange()
		for prefixRoutesIt.Next() {
			prefixRoutes := prefixRoutesIt.Value()
			prefixRoutesLen := prefixRoutes.Len()

			for j := 0; j < prefixRoutesLen; j++ {
				route := prefixRoutes.Index(j).Elem()
				r := &Route{}

				r.Pos = (uint32)(route.FieldByName("pos").Uint())
				r.Use = route.FieldByName("use").Bool()
				r.Star = route.FieldByName("star").Bool()
				r.Root = route.FieldByName("root").Bool()
				r.Method = route.FieldByName("Method").String()
				r.Name = route.FieldByName("Name").String()
				r.Path = route.FieldByName("Path").String()

				paramsValue := route.FieldByName("Params")
				r.Params = make([]string, paramsValue.Len())
				for i := 0; i < paramsValue.Len(); i++ {
					r.Params[i] = paramsValue.Index(i).String()
				}

				handlersValue := route.FieldByName("Handlers")
				r.HandlerName = make([]string, handlersValue.Len())
				for i := 0; i < handlersValue.Len(); i++ {
					fullName := runtime.FuncForPC(handlersValue.Index(i).Pointer()).Name()
					funcName := strings.TrimPrefix(fullName, path.Dir(fullName))
					r.HandlerName[i] = funcName
				}

				klog.Info(r)

			}
		}
	}

	return routeCount, nil
}
