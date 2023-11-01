package clients

import (
	"context"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/go-resty/resty/v2"
)

type Usecase interface {
	GetAllUactual(ctx context.Context) ([]orders.Order, error)
	Update(ctx context.Context, ordrs []orders.Order) error
}

type client struct {
	address string
	cl      *resty.Client
	logger  *logger.Logger
	usecase Usecase
	stop    chan bool
}

func New(logger *logger.Logger, address string, cl *resty.Client, usecase Usecase) *client {
	cl.SetRetryCount(3).SetRetryWaitTime(30 * time.Second).SetRetryMaxWaitTime(90 * time.Second).SetTimeout(time.Second*5)
	c := &client{
		address: address,
		cl:      cl,
		logger:  logger,
		usecase: usecase,
		stop:    make(chan bool),
	}
	go c.updOrders(c.stop)
	return c
}

func (c *client) updOrders(stop <-chan bool) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		c.updOrdersNow()
	}
}

func (c *client) updOrdersNow() {
	ctx, gansel := context.WithTimeout(context.Background(), time.Second*30)
	defer gansel()
	ordrs, err := c.usecase.GetAllUactual(ctx)
	if err != nil {
		c.logger.Errorf("Orders: failed to get actual orders from db, err value: %s", err)
		return
	}
	if len(ordrs) == 0 {
		return
	}

	ordersUpdated := make([]orders.Order, 0)
	for _, o := range ordrs {
		_, err := c.cl.R().SetResult(&o).SetPathParams(map[string]string{
			"ordNum": o.Ord,
		}).Get(c.address + "/api/orders/{ordNum}")
		if err != nil {
			c.logger.Errorf("Orders: failed to call accrulal, err value: %s", err)
			return
		}
		ordersUpdated = append(ordersUpdated, o)
	}

	ctx, gansel = context.WithTimeout(context.Background(), time.Second*30)
	defer gansel()
	err = c.usecase.Update(ctx, ordersUpdated)
	if err != nil {
		c.logger.Errorf("Orders: failed to update orders, err value: %s", err)
		return
	}
}
