package repository

import (
	"context"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"wb-l0/models"
)

func NewPostgresDb(uri string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), uri)
	return conn, err
}

type Repository struct {
	Service
}

type Service interface {
	GetData(id string) (*models.Data, error)
	InsertData(data *models.Data) error
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{
		Service: NewDatadb(conn),
	}
}

type Datadb struct {
	db *pgx.Conn
}

func NewDatadb(conn *pgx.Conn) *Datadb {
	return &Datadb{
		db: conn,
	}
}

func (r *Datadb) GetData(id string) (*models.Data, error) {

	var data models.Data
	var items models.Items
	var payment models.Payment
	var delivery models.Delivery

	query := "SELECT * FROM data WHERE order_uid=$1;"
	err := r.db.QueryRow(context.Background(), query, id).Scan(&data.OrderUID, &data.TrackNumber, &data.Entry, &data.Locale,
		&data.InternalSignature, &data.CustomerID, &data.DeliveryService, &data.Shardkey, &data.SmID, &data.DateCreated, &data.OofShard)
	if err != nil {
		zap.L().Error("Failed to get data", zap.Error(err))

		return nil, err
	}

	query = "SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid=$1;"
	err = r.db.QueryRow(context.Background(), query, id).Scan(&items.ChrtID, &items.TrackNumber, &items.Price,
		&items.Rid, &items.Name, &items.Sale, &items.Size, &items.TotalPrice, &items.NmID, &items.Brand, &items.Status)
	if err != nil {
		zap.L().Error("Failed to get items", zap.Error(err))

		return nil, err
	}

	query = "SELECT * FROM payment WHERE transaction=$1;"
	err = r.db.QueryRow(context.Background(), query, id).Scan(&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
		&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
	if err != nil {
		zap.L().Error("Failed to get payment", zap.Error(err))

		return nil, err
	}

	query = "SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid=$1;"
	err = r.db.QueryRow(context.Background(), query, id).Scan(&delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		zap.L().Error("Failed to get delivery", zap.Error(err))

		return nil, err
	}

	data.Delivery = delivery
	data.Payment = payment
	data.Items = []models.Items{items}

	return &data, err
}

func (r *Datadb) InsertData(data *models.Data) error {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO data VALUES ($1, $2, $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10 , $11);"

	_, err = tx.Exec(ctx, query, data.OrderUID, data.TrackNumber, data.Entry, data.Locale, data.InternalSignature,
		data.CustomerID, data.DeliveryService, data.Shardkey, data.SmID, data.DateCreated, data.OofShard)
	if err != nil {
		zap.L().Error("Failed to insert into data", zap.Error(err))
		return err
	}

	query = "INSERT INTO payment VALUES ($1, $2, $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10);"
	_, err = tx.Exec(ctx, query, data.Payment.Transaction, data.Payment.RequestID, data.Payment.Currency, data.Payment.Provider, data.Payment.Amount,
		data.Payment.PaymentDt, data.Payment.Bank, data.Payment.DeliveryCost, data.Payment.GoodsTotal, data.Payment.CustomFee)
	if err != nil {
		zap.L().Error("Failed to insert into payment", zap.Error(err))
		return err
	}

	query = "INSERT INTO items VALUES ($1, $2, $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10 , $11, $12);"
	_, err = tx.Exec(ctx, query, data.Items[0].ChrtID, data.Items[0].TrackNumber, data.Items[0].Price, data.Items[0].Rid, data.Items[0].Name,
		data.Items[0].Sale, data.Items[0].Size, data.Items[0].TotalPrice, data.Items[0].NmID, data.Items[0].Brand, data.Items[0].Status, data.OrderUID)
	if err != nil {
		zap.L().Error("Failed to insert into items", zap.Error(err))
		return err
	}

	query = "INSERT INTO delivery VALUES ($1, $2, $3 , $4 , $5 , $6 , $7 , $8);"
	_, err = tx.Exec(ctx, query, data.Delivery.Name, data.Delivery.Phone, data.Delivery.Zip, data.Delivery.City,
		data.Delivery.Address, data.Delivery.Region, data.Delivery.Email, data.OrderUID)
	if err != nil {
		zap.L().Error("Failed to insert into delivery", zap.Error(err))
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
