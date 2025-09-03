<?php
namespace App\Command;

use Symfony\Component\Console\Attribute\AsCommand;

#[AsCommand(name: 'rates:avg:collect')]
class AvgRatesCollectCommand extends RatesCollectCommandBase
{
    protected function fetchFromRedis(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array
    {
        $key = "candles:10s:{$symbol}";

        $entries = $this->redis->xread(0, 1000, [$key, '0']);
        $entries = !empty($entries)
            ? reset($entries)
            : [];

        $candles = [];
        foreach ($entries as $entry) {
            $entry = $this->createKeyValueEntity($entry[1]);
            $tfCheck = $this->checkTimeframe($entry, $start, $end);
            if ($tfCheck > 0) {
                // is after TF - stop reading stream
                break;
            } else if ($tfCheck < 0) {
                // is before TF - skip
                continue;
            }

            $candles[] = $entry;
        }

        return !empty($candles)
            ? ['price' => $this->calcVWAP($candles)]
            : [];
    }

    protected function fetchFromBinance(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array
    {
        // doc link: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/market-data-endpoints#klinecandlestick-data
        $resp = $this->http->request('GET', 'https://api.binance.com/api/v3/klines', [
            'query' => [
                'symbol' => strtoupper($symbol),
                'interval' => '5m',
                'startTime' => $start->getTimestamp() * 1000,
                'endTime' => $end->getTimestamp() * 1000,
                'limit' => 1
            ]
        ]);

        if ($resp->getStatusCode() === 200) {
            $data = $resp->toArray();
            if (!empty($data)) {
                // Format:
                // [
                //     1499040000000,      // Kline open time
                //     "0.01634790",       // Open price
                //     "0.80000000",       // High price
                //     "0.01575800",       // Low price
                //     "0.01577100",       // Close price
                //     "148976.11427815",  // Volume
                //     1499644799999,      // Kline Close time
                //     "2434.19055334",    // Quote asset volume
                //     308,                // Number of trades
                //     "1756.87402397",    // Taker buy base asset volume
                //     "28.46694368",      // Taker buy quote asset volume
                //     "0"                 // Unused field, ignore.
                // ]

                [$_, $_, $high, $low, $close] = $data[0];
                return [
                    'price' => $this->calcTP([
                        'h' => $high,
                        'l' => $low,
                        'c' => $close,
                    ]),
                ];
            }
        }

        return [];
    }

    protected function saveToMySQL(string $symbol, array $data, \DateTimeImmutable $start, \DateTimeImmutable $end): void
    {
        $sql = <<<SQL
            INSERT INTO avg_rates (pair, rate, start_time, end_time)
            VALUES (:pair, :rate, :start_time, :end_time)
            ON DUPLICATE KEY UPDATE
                rate = VALUES(rate)
        SQL;

        // we always use [$start,$end] params instead of $data time
        // to make sure we save the right 5m frame
        $this->db->executeStatement($sql, [
            'pair' => $symbol,
            'rate' => $data['price'],
            'start_time' => $start->format('Y-m-d H:i:s'),
            'end_time' => $end->format('Y-m-d H:i:s'),
        ]);
    }


    private function calcTP(array $candle): float {
        // Formula:
        // TP = (high + low + close) / 3

        return ($candle['h'] + $candle['l'] + $candle['c']) / 3;
    }

    private function calcTPV(array $candle): float {
        // Formula:
        // TPV = TP * volume

        return $this->calcTP($candle) * $candle['v'];
    }

    private function calcVWAP(array $candles): float {
        // Formulas:
        // TP = (high + low + close) / 3
        // TPV = TP * volume
        // VWAP = sum(TPV) / sum(volume)

        $totalTPV = 0;
        $totalVolume = 0;
        foreach($candles as $candle) {
            $totalTPV += $this->calcTPV($candle);
            $totalVolume += $candle['v'];
        }

        return $totalTPV / $totalVolume;
    }

    private function createKeyValueEntity(array $entity): array {
        $formatted = [];
        for ($i = 0; $i < count($entity); $i += 2) {
            $formatted[$entity[$i]] = $entity[$i + 1];
        }

        return $formatted;
    }

    /**
     * Returns:
     * -1 if entity is ealier than TF
     * 0 if entity is within TF
     * 1 if entity is after TF
     */
    private function checkTimeframe(array $entry, \DateTimeImmutable $start, \DateTimeImmutable $end): int {
        $enTime = new \DateTimeImmutable('@' . ($entry['ts']/1000));

        if ($enTime < $start) {
            return -1;
        }

        if ($enTime > $end) {
            return 1;
        }

        return 0;
    }
}
