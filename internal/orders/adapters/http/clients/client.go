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
	now     chan time.Time
	stop    chan bool
}

func New(logger *logger.Logger, address string, cl *resty.Client, usecase Usecase) *client {
	cl.SetRetryCount(3).SetRetryWaitTime(30 * time.Second).SetRetryMaxWaitTime(90 * time.Second)
	c := &client{
		address: address,
		cl:      cl,
		logger:  logger,
		usecase: usecase,
		now:     make(chan time.Time),
		stop:    make(chan bool),
	}
	go c.updOrders(c.stop, c.now)
	return c
}

func (c *client) updOrders(stop <-chan bool, now <-chan time.Time) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-stop:
			return
		case <-now:
			c.updOrdersNow()
		case <-ticker.C:
			c.updOrdersNow()
		}
	}
}

func (c *client) updOrdersNow() {
	ctx, gansel := context.WithTimeout(context.Background(), time.Second*30)
	defer gansel()
	orders, err := c.usecase.GetAllUactual(ctx)
	if err != nil {
		c.logger.Errorf("Orders: failed to get actual orders from db, err value: %s", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	for i, o := range orders {
		_, err := c.cl.R().SetResult(&orders[i]).SetPathParams(map[string]string{
			"ordNum": o.Ord,
		}).Get(c.address + "/api/orders/{ordNum}")
		if err != nil {
			c.logger.Errorf("Orders: failed to call accrulal, err value: %s", err)
			return
		}
	}

	ctx, gansel = context.WithTimeout(context.Background(), time.Second*30)
	defer gansel()
	err = c.usecase.Update(ctx, orders)
	if err != nil {
		c.logger.Errorf("Orders: failed to update orders, err value: %s", err)
		return
	}
}

func (c *client) UpdOrdersNow() {
	c.now <- time.Now()
}

func (c *client) Stop() {
	close(c.now)
	close(c.stop)
}
