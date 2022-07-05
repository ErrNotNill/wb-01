package handler

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"text/template"
	"wb-l0/models"
)

func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	w.Write([]byte("Hello"))
	t, err := template.ParseFiles("./templates/get_request.html")
	if err != nil {
		zap.L().Error("Failed to parse files", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse files"))

		return
	}

	t.Execute(w, &models.Page{Title: "Data", Msg: "Введите id"})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DataResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")

	orderId := models.OrderId{
		Id: r.FormValue("id"),
	}

	data, found := h.cache.Get(orderId.Id)
	if !found {
		data, err := h.repo.GetData(orderId.Id)
		if err != nil {
			zap.L().Error("Failed to get data")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get data"))

			return
		}

		err = SendData(w, r, data)
		if err != nil {
			zap.L().Error("Failed to send data", zap.Error(err))

			return
		}
	}

	if data != nil {
		err := SendData(w, r, data)
		if err != nil {
			zap.L().Error("Failed to send data", zap.Error(err))

			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func SendData(w http.ResponseWriter, r *http.Request, d interface{}) error {
	msg, err := json.Marshal(d)
	if err != nil {
		zap.L().Error("Failed to send data", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to send data"))

		return err
	}

	var data models.Data
	err = json.Unmarshal(msg, &data)
	if err != nil {
		zap.L().Error("Failed to send data", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to send data"))

		return err
	}

	w.Header().Set("Content-type", "text/html")

	t, err := template.ParseFiles("./templates/table.html")
	if err != nil {
		zap.L().Error("Failed to parse files", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse files"))

		return err
	}

	t.Execute(w, &models.Data{OrderUID: data.OrderUID, TrackNumber: data.TrackNumber, Entry: data.Entry, Locale: data.Locale,
		InternalSignature: data.InternalSignature, CustomerID: data.CustomerID, DeliveryService: data.DeliveryService,
		Shardkey: data.Shardkey, SmID: data.SmID, DateCreated: data.DateCreated, OofShard: data.OofShard, Delivery: data.Delivery, Payment: data.Payment, Items: data.Items})

	w.WriteHeader(http.StatusOK)

	return nil
}
