package ws_helper

import (
	"strings"
	"log"
	"fmt"
	"context"
	"time"
	"encoding/json"
	"net/url"

    "github.com/gorilla/websocket"

	"crypto_rates_collector/internal/config"
	redis_helper "crypto_rates_collector/internal/redis"
)

type Message struct {
    Stream string `json:"stream"`
    Data redis_helper.TradeEvent `json:"data"`
}

type WSClient struct {
    url string
    symbols []string
    redisW *redis_helper.Writer
}

func NewClient(cfg *config.Config, rw *redis_helper.Writer) *WSClient {
    streams := make([]string, 0, len(cfg.Symbols))
    for _, s := range cfg.Symbols {
        streams = append(streams, strings.ToLower(s)+"@trade")
    }


    u := url.URL{Scheme: "wss", Host: strings.TrimPrefix(cfg.BinanceWS, "wss://")}
    u.Path = "/stream"

    q := u.Query()
    q.Set("streams", strings.Join(streams, "/"))

    u.RawQuery = q.Encode()

    return &WSClient{
        url: u.String(),
        symbols: cfg.Symbols,
        redisW: rw,
    }
}

func (c *WSClient) Run(ctx context.Context) {
    backoff := time.Second

    for {
        if err := c.runOnce(ctx); err != nil {
            log.Printf("[ws] error: %v (reconnecting in %s)", err, backoff)

            t := time.NewTimer(backoff)
            select {
            case <-ctx.Done():
                return
            case <-t.C:
            }

            if backoff < 30 * time.Second {
                backoff *= 2
            }

            continue
        }

        return
    }
}

func (c *WSClient) runOnce(ctx context.Context) error {
    log.Printf("[ws] connecting: %s", c.url)

    dialer := websocket.Dialer{
        HandshakeTimeout: 10 * time.Second,
    }

    conn, _, err := dialer.DialContext(ctx, c.url, nil)
    if err != nil {
        return err
    }
    defer conn.Close()

    log.Printf("[ws] connected")


    conn.SetPongHandler(func(appData string) error {
        _ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

        return nil
    })
    _ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))


    for {
        select {
        case <-ctx.Done():
            return nil

        default:
            _, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }

            var msg Message
            if err := json.Unmarshal(data, &msg); err != nil {
                log.Printf("[ws] unmarshal error: %v", err)

                continue
            }

            if msg.Data.EventType != "trade" {
                continue
            }
            
            fmt.Printf("Tick: %v\n", msg.Data)
            if err := c.redisW.WriteTick(&msg.Data); err != nil {
                log.Printf("[redis] write error: %v", err)
            }
        }
    }
}
