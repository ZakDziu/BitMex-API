package api

//nolint:revive
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bitmex-api/docs"
	"bitmex-api/pkg/authmiddleware"
	"bitmex-api/pkg/config"
	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model/bitmex"
	"bitmex-api/pkg/store"
)

const (
	bitMexWebSocketURL     = "wss://testnet.bitmex.com/realtime"
	bitMexAPIActiveSymbols = "https://testnet.bitmex.com/api/v1/instrument/active"
	goroutineCount         = 2
	subscribeMessageStart  = `{"op": "subscribe", "args": [`
	subscribeMessageEnd    = `]}`
)

type Server struct {
	*http.Server
}

type api struct {
	postgresStore *store.Store
	router        *gin.Engine
	config        *config.ServerConfig
	auth          authmiddleware.AuthMiddleware

	bitMexWSConn *websocket.Conn

	allSymbols allSymbols
	symbolUser symbolUser
	userWSConn userWSConn

	authHandler          *AuthHandler
	userHandler          *UserHandler
	bitMexHandler        *BitMexHandler
	userWebSocketHandler *UserWebSocketHandler
}

type symbolUser struct {
	symbolUserSubscriptions map[string][]uuid.UUID

	mu sync.RWMutex
}

type allSymbols struct {
	allSymbols []string

	mu sync.RWMutex
}

type userWSConn struct {
	userConn map[uuid.UUID]*websocket.Conn
	connUser map[*websocket.Conn]uuid.UUID

	mu sync.RWMutex
}

func NewServer(
	ctx context.Context,
	config *config.ServerConfig,
	postgresStore *store.Store,
	auth authmiddleware.AuthMiddleware,
	wg *sync.WaitGroup,
) *Server {
	handler := newAPI(ctx, config, postgresStore, auth, wg)

	srv := &http.Server{
		Addr:              config.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: config.ReadTimeout.Duration,
	}

	return &Server{
		Server: srv,
	}
}

//nolint:varnamelen
func newAPI(
	ctx context.Context,
	config *config.ServerConfig,
	postgresStore *store.Store,
	auth authmiddleware.AuthMiddleware,
	wg *sync.WaitGroup,
) *api {
	api := &api{
		config:        config,
		postgresStore: postgresStore,
		auth:          auth,
		allSymbols: allSymbols{
			allSymbols: make([]string, 0),
			mu:         sync.RWMutex{},
		},
		symbolUser: symbolUser{
			symbolUserSubscriptions: make(map[string][]uuid.UUID),
			mu:                      sync.RWMutex{},
		},
		userWSConn: userWSConn{
			userConn: make(map[uuid.UUID]*websocket.Conn),
			connUser: make(map[*websocket.Conn]uuid.UUID),
			mu:       sync.RWMutex{},
		},
	}
	wg.Add(goroutineCount)

	channelConnected := make(chan struct{})
	go api.connectToBitMex(ctx, wg, channelConnected)
	<-channelConnected
	api.updateSymbols()
	api.subscribeToAllSymbols()
	api.updateUserSubscriptionFromDB()

	go api.BitMex().SendUsersDataOnUpdate(ctx, wg)

	api.router = configureRouter(api)

	api.router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return api
}

//nolint:varnamelen
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding,"+
			"X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)

			return
		}

		c.Next()
	}
}

func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *api) Auth() *AuthHandler {
	if a.authHandler == nil {
		a.authHandler = NewAuthHandler(a)
	}

	return a.authHandler
}

//nolint:gocritic
func (a *api) getUserIDFromHeader(c *gin.Context) (uuid.UUID, error) {
	token := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
	userID, err := a.auth.GetUserID(token)

	return userID, err
}

func (a *api) User() *UserHandler {
	if a.userHandler == nil {
		a.userHandler = NewUserHandler(a)
	}

	return a.userHandler
}

func (a *api) BitMex() *BitMexHandler {
	if a.bitMexHandler == nil {
		a.bitMexHandler = NewBitMexHandler(a)
	}

	return a.bitMexHandler
}

func (a *api) UserWebSocket() *UserWebSocketHandler {
	if a.userWebSocketHandler == nil {
		a.userWebSocketHandler = NewUserWebSocketHandler(a)
	}

	return a.userWebSocketHandler
}

//nolint:gosimple
func (a *api) connectToBitMex(ctx context.Context, wg *sync.WaitGroup, channelConnected chan struct{}) {
	defer wg.Done()

	c, resp, err := websocket.DefaultDialer.Dial(bitMexWebSocketURL, nil)
	defer func() {
		if err = resp.Body.Close(); err != nil {
			logger.Errorf("error close response body", err)
		}
		if err = c.Close(); err != nil {
			logger.Errorf("error close connection", err)
		}
	}()
	if err != nil {
		logger.Errorf("Dial error", err)
	}

	a.bitMexWSConn = c
	channelConnected <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			unsubscribeMessage := `{"op": "unsubscribe", "args": [`

			symbols := a.allSymbols.GetAll()
			for i, symbol := range symbols {
				if i == len(symbols)-1 {
					unsubscribeMessage += fmt.Sprintf(`"trade:%s"`, symbol)
				}
				unsubscribeMessage += fmt.Sprintf(`"trade:%s", `, symbol)
			}

			unsubscribeMessage += subscribeMessageEnd

			if err = c.WriteMessage(websocket.TextMessage, []byte(unsubscribeMessage)); err != nil {
				logger.Errorf("Unsubscribe error", err)
			}
			time.Sleep(1 * time.Second)

			logger.Infof("connectToBitMex done")

			return
		}
	}
}

func (a *api) subscribeToAllSymbols() {
	subscribeMessage := subscribeMessageStart
	newCycle := true
	symbols := a.allSymbols.GetAll()
	for i, symbol := range symbols {
		if newCycle {
			subscribeMessage += fmt.Sprintf(`"trade:%s"`, symbol)
			newCycle = false
		}

		subscribeMessage += fmt.Sprintf(`, "trade:%s"`, symbol)
		if (i+1)%15 == 0 {
			subscribeMessage += subscribeMessageEnd

			logger.Infof(subscribeMessage)

			if err := a.bitMexWSConn.WriteMessage(websocket.TextMessage, []byte(subscribeMessage)); err != nil {
				logger.Errorf("Unsubscribe error", err)
			}

			subscribeMessage = subscribeMessageStart
			newCycle = true
		}
	}

	subscribeMessage += subscribeMessageEnd

	if err := a.bitMexWSConn.WriteMessage(websocket.TextMessage, []byte(subscribeMessage)); err != nil {
		logger.Errorf("Unsubscribe error", err)
	}
}

func (a *api) updateUserSubscriptionFromDB() {
	allUsers, err := a.postgresStore.User.GetAll()
	if err != nil {
		logger.Errorf("error get all users", err)
	}

	for _, user := range allUsers {
		if !user.Subscription {
			continue
		}
		if len(user.SubscriptionSymbols) == 0 {
			allSymbolNames := a.allSymbols.GetAll()
			for _, symbolName := range allSymbolNames {
				usersSubscribedToSymbol, _ := a.symbolUser.Get(symbolName)

				a.symbolUser.Insert(symbolName, append(usersSubscribedToSymbol, user.UserID))
			}
		}
		for _, symbol := range user.SubscriptionSymbols {
			usersSubscribedToSymbol, _ := a.symbolUser.Get(symbol)

			a.symbolUser.Insert(symbol, append(usersSubscribedToSymbol, user.UserID))
		}
	}
}

//nolint:noctx
func (a *api) updateSymbols() {
	response, err := http.Get(bitMexAPIActiveSymbols)
	if err != nil {
		logger.Errorf("HTTP request error", err)
	}
	defer response.Body.Close()

	var symbols []bitmex.SymbolInfo
	err = json.NewDecoder(response.Body).Decode(&symbols)
	if err != nil {
		logger.Errorf("JSON decoding error", err)
	}

	for _, symbol := range symbols {
		a.allSymbols.Update(symbol.Symbol)
	}

	symbolNames := a.allSymbols.GetAll()

	for _, symbolName := range symbolNames {
		_, ok := a.symbolUser.Get(symbolName)
		if !ok {
			a.symbolUser.Insert(symbolName, []uuid.UUID{})
		}
	}
}

func (m *allSymbols) GetAll() []string {
	m.mu.RLock()
	symbols := m.allSymbols
	m.mu.RUnlock()

	return symbols
}

func (m *allSymbols) Update(symbol string) {
	symbols := append(m.GetAll(), symbol)
	m.mu.Lock()
	m.allSymbols = symbols
	m.mu.Unlock()
}

func (m *symbolUser) Get(symbol string) ([]uuid.UUID, bool) {
	m.mu.RLock()
	users, ok := m.symbolUserSubscriptions[symbol]
	m.mu.RUnlock()

	return users, ok
}

func (m *symbolUser) Insert(symbol string, users []uuid.UUID) {
	m.mu.Lock()
	m.symbolUserSubscriptions[symbol] = users
	m.mu.Unlock()
}

func (a *api) DeleteSymbolUser(symbol string, userID uuid.UUID) {
	users, ok := a.symbolUser.Get(symbol)

	if !ok {
		a.updateSymbols()
		users, ok = a.symbolUser.Get(symbol)

		if !ok {
			return
		}
	}

	for i, u := range users {
		if u == userID {
			users = append(users[:i], users[i+1:]...)
		}
	}

	a.symbolUser.mu.Lock()
	a.symbolUser.symbolUserSubscriptions[symbol] = users
	a.symbolUser.mu.Unlock()
}

func (m *userWSConn) GetConn(userID uuid.UUID) (*websocket.Conn, bool) {
	m.mu.RLock()
	conn, ok := m.userConn[userID]
	m.mu.RUnlock()

	return conn, ok
}

func (m *userWSConn) Delete(conn *websocket.Conn) {
	m.mu.RLock()

	userID, ok := m.connUser[conn]
	if ok {
		delete(m.connUser, conn)
	}

	_, ok = m.userConn[userID]
	if ok {
		delete(m.userConn, userID)
	}

	m.mu.RUnlock()
}

func (m *userWSConn) Create(conn *websocket.Conn, userID uuid.UUID) {
	m.mu.Lock()
	m.connUser[conn] = userID
	m.userConn[userID] = conn
	m.mu.Unlock()
}
