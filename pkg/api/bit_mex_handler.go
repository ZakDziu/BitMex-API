package api

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model/bitmex"
)

type BitMexHandler struct {
	api *api
}

func NewBitMexHandler(a *api) *BitMexHandler {
	return &BitMexHandler{
		api: a,
	}
}

//nolint:nestif,gocognit,cyclop
func (h *BitMexHandler) SendUsersDataOnUpdate(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Infof("sendUsersDataOnUpdate done")

			return
		default:
			_, message, err := h.api.bitMexWSConn.ReadMessage()
			if err != nil {
				return
			}

			var tradeMessage bitmex.TradeMessage
			if err := json.Unmarshal(message, &tradeMessage); err != nil {
				logger.Errorf("JSON unmarshal error:", err)

				continue
			}

			if tradeMessage.Table == "trade" && tradeMessage.Action == "insert" {
				for _, record := range tradeMessage.Data {
					logger.Infof("Symbol: %s, Price: %f\n", record.Symbol, record.Price)
					var users []uuid.UUID
					users, ok := h.api.symbolUser.Get(record.Symbol)
					if !ok {
						h.api.updateSymbols()
						users, ok = h.api.symbolUser.Get(record.Symbol)
						if !ok {
							logger.Errorf("invalid symbol:", record.Symbol)

							continue
						}
					}

					for _, userID := range users {
						conn, ok := h.api.userWSConn.GetConn(userID)
						if !ok {
							continue
						}

						data, err := json.Marshal(record)
						if err != nil {
							logger.Errorf("JSON marshal:", err)

							continue
						}

						if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
							logger.Errorf("Subscribe error:", err)
						}
					}
				}
			}
		}
	}
}
