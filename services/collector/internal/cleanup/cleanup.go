package cleanup

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	redis_helper "crypto_rates_collector/internal/redis"
)

func cleanupSymbol(rw *redis_helper.Writer, stream string, ttl time.Duration, errCh chan error) {
    entries, err := rw.Client.XRange(rw.GetContext(), stream, "-", "+").Result()
    if err != nil {
        errCh <- err

        return
    }

    // remove only entries twice older than ttl 
	// for aggragators to have data
    ttlPlusGap := -ttl.Seconds() * 2
	cutoffTime := time.Now().Truncate(ttl).Add(time.Duration(ttlPlusGap) * time.Second)
    var trimID string
    for _, entry := range entries {
        timestampStr, ok := entry.Values["ts"].(string)
        if !ok {
            log.Printf("[cleanup] Failed to get timestamp for entry %v", entry)

            continue
        }

        ts, err := strconv.ParseInt(timestampStr, 10, 64)
        if err != nil {
            log.Printf("[cleanup] Failed to parse timestamp for entry with ID %v: %v", entry.ID, err)

            continue
        }

		timestamp := time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond))
        if timestamp.Compare(cutoffTime) <= 0 {
            trimID = entry.ID
        } else {
            break
        }
    }

	fmt.Printf("[cleanup] cleaning %s up to ID: %s\n", stream, trimID)
    if _, err = rw.Client.XTrimMinID(rw.GetContext(), stream, trimID).Result(); err != nil {
        errCh <- err

        return
    }

    errCh <- nil
}

func cleanup(ctx context.Context, rw *redis_helper.Writer, streams []string, ttl time.Duration) {
    if err := rw.Client.Ping(rw.GetContext()).Err(); err != nil {
        log.Printf("[cleanup] redis ping failed: %v", err)

        return
    }

    errCh := make(chan error, len(streams))

    go func() {
        <- ctx.Done()
        log.Printf("[cleanup] received cancel signal, closing err chan\n")

        close(errCh)
    }()
        

mainLoop:
    for {
        for _, stream := range streams {
            select {
            case <- ctx.Done():
                log.Printf("[cleanup] received cancel signal, shutting down\n")
                break mainLoop 
            default:
                go cleanupSymbol(rw, stream, ttl, errCh)
            }
        }

        i := 1
        for err := range errCh {
            i++

            if err != nil {
                log.Printf("[cleanup] failed to cleanup: %v", err)
            }

            if i > len(streams) {
                break
            }
        }

		timeToSleep := calcTimeToNextTf(ttl)
        fmt.Printf("[cleanup] Sleep for %v\n", timeToSleep)
        time.Sleep(timeToSleep)
    }
}

func calcTimeToNextTf(ttl time.Duration) time.Duration {
    now := time.Now()

    secondsPast := now.Second()
    remainder := secondsPast % int(ttl.Seconds()) 

    return time.Duration(int(ttl.Seconds())-remainder) * time.Second
}

func CleanupTicks(ctx context.Context, rw *redis_helper.Writer) {
    cleanup(ctx, rw, rw.GetAvailableTicksStreams(), rw.GetTickTTL())
}

func CleanupCandles(ctx context.Context, rw *redis_helper.Writer) {
    cleanup(ctx, rw, rw.GetAvailableCandlesStreams(), rw.GetCandleTTL())
}
