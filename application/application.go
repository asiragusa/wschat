// Package application is the entry point for the whole application.
package application

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/asiragusa/wschat/controller"
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/middleware"
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/services"
	"github.com/asiragusa/wschat/validator"
	"github.com/asiragusa/wschat/ws"
	"github.com/facebookgo/inject"
	"github.com/jonboulle/clockwork"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/websocket"
)

// AppConfig contains the app configuration
type AppConfig struct {
	// JwtSecret is the secret for encrypting JWT Tokens
	JwtSecret string

	// JwtIssuer is the issuer for the JWT Token eg. http://api.example.com
	JwtIssuer string

	// DatastoreClient is the client for google cloud's datastore
	DatastoreClient *datastore.Client

	// PubsubClient is the client for google cloud's pubsub
	PubsubClient *pubsub.Client
}

// The Controller interface defines the interface for the request handlers
type Controller interface {
	Handle(context.Context)
}

// The route struct contains the HTTP routes of the application
type Route struct {
	// HTTP method (GET, POST, etc)
	Method string

	// HTTP path
	Path string

	// Can contain an Iris party or *iris.Application
	Party router.Party

	// Request handler
	Controller Controller
}

type Application struct {
	config  *AppConfig
	irisApp *iris.Application

	routes []Route

	graph []*inject.Object
}

// Creates a new Application
func NewApplication(config *AppConfig) (*Application, error) {
	app := &Application{
		config:  config,
		irisApp: iris.New(),
	}

	app.getInjectObjects()
	app.getRoutes()
	app.initWs()
	app.initStatic()

	if err := app.init(); err != nil {
		return nil, err
	}

	return app, nil
}

// Init static routes
func (a *Application) initStatic() {
	a.irisApp.Get("/", func(ctx context.Context) {
		ctx.ServeFile("./public/index.html", false)
	})
	a.irisApp.StaticWeb("/js", "./public/js")
}

// Prepares all the objects to be injected
func (a *Application) getInjectObjects() {
	a.injectNamed(
		"accessTokenGenerator",
		services.NewTokenGenerator(a.config.JwtSecret, a.config.JwtIssuer, "access"))

	a.injectNamed(
		"wsTokenGenerator",
		services.NewTokenGenerator(a.config.JwtSecret, a.config.JwtIssuer, "ws"))

	a.inject(a.config.DatastoreClient)
	a.inject(a.config.PubsubClient)

	a.inject(clockwork.NewRealClock())

	a.inject(repository.NewUserRepository())
	a.inject(repository.NewMessageRepository())
	a.inject(repository.NewSubscriptionRepository())

	a.inject(services.NewPubsubClient())

	a.inject(validator.NewValidator())

	a.inject(interactor.NewRegisterInteractor())
	a.inject(interactor.NewLoginInteractor())
	a.inject(interactor.NewListMessagesInteractor())
	a.inject(interactor.NewListUsersInteractor())
	a.inject(interactor.NewCreateMessageInteractor())
	a.inject(interactor.NewWsTokenInteractor())
}

// Initializes the websocket endpoint
func (a *Application) initWs() {
	wsMiddleware := middleware.NewWsMiddleware()
	wsHandler := ws.NewWsHandler()

	a.inject(wsMiddleware)
	a.inject(wsHandler)

	s := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	s.OnConnection(wsHandler.HandleConnection)

	wsParty := a.irisApp.Party("/ws", wsMiddleware.Handle)
	wsParty.Get("/", s.Handler())

	a.irisApp.Any("/iris-ws.js", func(ctx context.Context) {
		ctx.Write(websocket.ClientSource)
	})
}

// Initializes the routes
func (a *Application) getRoutes() {
	authenticatedMiddleware := middleware.NewAuthenticatedMiddleware()
	a.inject(authenticatedMiddleware)

	messagesParty := a.irisApp.Party("/messages", authenticatedMiddleware.Handle)
	usersParty := a.irisApp.Party("/users", authenticatedMiddleware.Handle)
	wsTokenParty := a.irisApp.Party("/wsToken", authenticatedMiddleware.Handle)

	a.routes = []Route{
		{
			Method:     iris.MethodPost,
			Path:       "/register",
			Party:      a.irisApp,
			Controller: controller.NewRegisterController(),
		},
		{
			Method:     iris.MethodPost,
			Path:       "/login",
			Party:      a.irisApp,
			Controller: controller.NewLoginController(),
		},
		{
			Method:     iris.MethodGet,
			Path:       "/",
			Party:      messagesParty,
			Controller: controller.NewListMessagesController(),
		},
		{
			Method:     iris.MethodPost,
			Path:       "/",
			Party:      messagesParty,
			Controller: controller.NewCreateMessageController(),
		},
		{
			Method:     iris.MethodGet,
			Path:       "/",
			Party:      usersParty,
			Controller: controller.NewListUsersController(),
		},
		{
			Method:     iris.MethodPost,
			Path:       "/",
			Party:      wsTokenParty,
			Controller: controller.NewWsTokenController(),
		},
	}
}

// Initializes the dependency graph
func (a *Application) init() error {
	for _, route := range a.routes {
		a.inject(route.Controller)
	}

	var g inject.Graph
	if err := g.Provide(a.graph...); err != nil {
		return err
	}

	if err := g.Populate(); err != nil {
		return err
	}

	for _, route := range a.routes {
		route.Party.Handle(route.Method, route.Path, route.Controller.Handle)
	}

	return nil
}

// Helper method to inject objects
func (a *Application) inject(v interface{}) {
	a.graph = append(a.graph, &inject.Object{Value: v})
}

// Helper method to inject named objects
func (a *Application) injectNamed(name string, v interface{}) {
	a.graph = append(a.graph, &inject.Object{Value: v, Name: name})
}

// Returns the configured *iris.Application
func (a *Application) GetRouter() *iris.Application {
	return a.irisApp
}
