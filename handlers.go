package main

import (
	"fmt"
	"github.com/dimonchik0036/Miniapps-wrapper"
	"github.com/dimonchik0036/NSUBot/news"
	"github.com/valyala/fasthttp"
	"log"
	"sync"
	"time"
)

type Handler struct {
	PermissionLevel int
	Handler         func(request *mapps.Request, user *User) string
}

type Handlers struct {
	Mux                sync.RWMutex
	Handler            map[string]Handler
	CommandInterpreter func(string) string
}

func (h Handlers) AddHandler(handler Handler, key ...string) {
	h.Mux.Lock()
	defer h.Mux.Unlock()

	if h.Handler == nil {
		h.Handler = map[string]Handler{}
	}

	for _, v := range key {
		h.Handler[v] = handler
	}
}

func (h *Handlers) GetHandler(key string) (Handler, bool) {
	h.Mux.RLock()
	defer h.Mux.RUnlock()
	handler, ok := h.Handler[h.CommandInterpreter(key)]
	return handler, ok
}

func NewHandlers(commandInterpreter func(string) string) Handlers {
	return Handlers{
		Handler:            map[string]Handler{},
		CommandInterpreter: commandInterpreter,
	}
}

func NewsHandler(subscribers []string, news []news.News, title string) {
	for _, s := range subscribers {
		subscriber := GlobalSubscribers.User(s)
		if subscriber == nil {
			log.Printf("%s %s", s, " WTF?! nil pointer in newshandler: "+title)
			continue
		}

		for _, n := range news {
			var response string
			var h func(string) string
			if subscriber.Protocol == mapps.Telegram {
				h = func(in string) string {
					return mapps.EscapeString(mapps.EscapeString(in))
				}
			} else {
				h = func(in string) string {
					return mapps.EscapeStringYesBr(in)
				}
			}

			response += h(title) + mapps.Br + h(n.URL) + mapps.Br + mapps.Br
			if n.Title != "" {
				response += h(n.Title) + mapps.Br
			}

			if n.Decryption != "" && n.Decryption != n.Title {
				response += h(n.Decryption) + mapps.Br
			}
			if n.Date != 0 {
				response += h(time.Unix(n.Date, 0).Format("02.01.2006"))
			}

			if err := subscriber.SendMessageBlock(mapps.Div("", response)); err != nil {
				log.Printf("%s %s", subscriber.String(), err.Error())
			}
			time.Sleep(time.Millisecond * 333)
		}
	}
}

func MainHandler(request *mapps.Request) {
	subscriber, ok := CheckNewSubscriber(request)
	subscriber.Queue.Lock()
	defer subscriber.Queue.Unlock()
	subscriber.MessageCount++
	printLog(request, subscriber)

	if ok {
		request.Ctx.WriteString(PageStart(request, subscriber))
		return
	}

	if !CommandHandler(request, subscriber) {
		PagesHandler(request, subscriber)
	}
}

func printLog(request *mapps.Request, subscriber *User) {
	SystemLogger.Print(request.AllFields())
	log.Print(subscriber.String() + " " + request.String())
	refreshSubscriber(subscriber)
}

func stringCheck(s string) string {
	if s == "" {
		return ""
	} else {
		return mapps.Data(s) + mapps.Br
	}
}

func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	req, err := mapps.Decode(string(ctx.RequestURI()))
	if err != nil {
		log.Print(err, " ", string(ctx.RequestURI()))
		fmt.Fprint(ctx, "404 Not Found")
		return
	}

	req.Ctx = ctx
	MainHandler(&req)
}

func HandlerStart() {
	Admin.SendMessageTelegram("Начинаю!")
	fmt.Sprintln()
	fasthttp.ListenAndServe(Port, fastHTTPHandler)
}
