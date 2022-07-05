package handler

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"time"
	"wb-l0/models"
)

func (h *Handler) natsHandler() {
	h.nats.Subscribe("data", func(m *nats.Msg) {
		data := &models.Data{}
		err := json.Unmarshal(m.Data, data)
		if err != nil {
			zap.L().Error("error unmarshal data", zap.Error(err))
			return
		}

		h.cache.Set(data.OrderUID, data, 10*time.Minute)

		err = h.repo.InsertData(data)
		if err != nil {
			zap.L().Error("error insert data into postgres", zap.Error(err))

			return
		}

		zap.L().Info("insert data into postgres")
	})
}
