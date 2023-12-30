package api

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/model/ui/subscription"
)

const (
	// ReadBufferSize is buffer sizes for read.
	ReadBufferSize int = 1024
	// WriteBufferSize is buffer sizes for write.
	WriteBufferSize int = 1024
)

type UserWebSocketHandler struct {
	api *api
}

func NewUserWebSocketHandler(a *api) *UserWebSocketHandler {
	return &UserWebSocketHandler{
		api: a,
	}
}

// SubscribeAction
// @Summary subscribe or unsubscribe on bitMex price update
// @Produce json
// @Tags User
// @Security ApiKeyAuth
// @Param User  body subscription.Request  true "Subscription Request"
// @Success 200 {object} subscription.Response
// @Failure 400 {object} errors.UIResponseErrorBadRequest
// @Router /api/v1/bit-mex/subscription [patch]
//
//nolint:varnamelen
func (h *UserWebSocketHandler) SubscribeAction(c *gin.Context) {
	action := &subscription.Request{}
	err := c.ShouldBindJSON(&action)
	if err != nil {
		logger.Errorf("Subscribe.ShouldBindJSON", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	userID, err := h.api.getUserIDFromHeader(c)
	if err != nil {
		logger.Errorf("Subscribe.getUserRoleFromHeader", err)
		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	user, err := h.api.postgresStore.User.Get(userID)
	if err != nil {
		logger.Errorf("Subscribe.Get", err)
		c.JSON(http.StatusBadRequest, model.ErrUnhealthy)

		return
	}

	if action.Action == subscription.Subscribe {
		if err = h.subscribeActions(user, action); err != nil {
			c.JSON(http.StatusBadRequest, err)

			return
		}
	} else {
		if !user.Subscription {
			c.JSON(http.StatusBadRequest, model.ErrAlreadyUnsubscribed)

			return
		}

		allSymbolNames := h.api.allSymbols.GetAll()

		for _, symbol := range allSymbolNames {
			h.api.DeleteSymbolUser(symbol, userID)
		}

		user.Subscription = false
		user.SubscriptionSymbols = []string{}
	}

	if err = h.api.postgresStore.User.Update(user); err != nil {
		logger.Errorf("Subscribe.Update", err)
		c.JSON(http.StatusBadRequest, model.ErrUnhealthy)

		return
	}

	c.JSON(http.StatusOK, subscription.Response{Success: true})
}

func (h *UserWebSocketHandler) subscribeActions(user *model.User, action *subscription.Request) error {
	if len(action.Symbols) == 0 {
		symbolNames := h.api.allSymbols.GetAll()
		for _, symbolName := range symbolNames {
			users, ok := h.api.symbolUser.Get(symbolName)
			if !ok {
				continue
			}

			h.api.symbolUser.Insert(symbolName, append(users, user.UserID))
		}

		user.SubscriptionSymbols = []string{}
	}

	for _, symbol := range action.Symbols {
		users, ok := h.api.symbolUser.Get(symbol)

		if !ok {
			return model.ErrIncorrectSymbol
		}

		if slices.Contains(users, user.UserID) || slices.Contains(user.SubscriptionSymbols, symbol) {
			return model.ErrAlreadySubscribed
		}

		h.api.symbolUser.Insert(symbol, append(users, user.UserID))

		user.SubscriptionSymbols = append(user.SubscriptionSymbols, symbol)
	}

	return nil
}

//nolint:varnamelen
func (h *UserWebSocketHandler) Connect(c *gin.Context) {
	userID, err := h.api.getUserIDFromHeader(c)
	if err != nil {
		logger.Errorf("can't get user from header", nil)
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("can't create connection:", userID)

		return
	}

	defer func() {
		h.api.userWSConn.Delete(conn)
		conn.Close()
	}()

	h.api.userWSConn.Create(conn, userID)

	logger.Infof("user connected %v", userID)

	for {
		messageType, _, err := conn.ReadMessage()
		if messageType == -1 || err != nil {
			return
		}
	}
}
