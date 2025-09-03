<?php
namespace App\Command;

use Symfony\Component\Console\Attribute\AsCommand;

#[AsCommand(name: 'rates:latest:collect')]
class LatestRatesCollectCommand extends RatesCollectCommandBase
{
    protected function fetchFromRedis(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array
    {
        $key = "ticks:{$symbol}";
        $entries = $this->redis->xrevrange($key, '+', '-', 1);

        $entry = [];
        if (!empty($entries)) {
            $entr = reset($entries);

            // only use redis data if info is not older than 5 seconds
            $eTime = new \DateTimeImmutable('@' . ($entr['ts']/1000));
            if ($eTime > $end->modify('-5 seconds')) {
                $entry = [
                    'price' => $entr['price'],
                    // converting ms to seconds
                    'time' => $eTime,
                ];
            }
        }

        return $entry;
    }

    protected function fetchFromBinance(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array
    {
        // doc link: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/market-data-endpoints#symbol-price-ticker
        $resp = $this->http->request('GET', 'https://api.binance.com/api/v3/ticker/price', [
            'query' => [
                'symbol' => strtoupper($symbol),
            ]
        ]);

        if ($resp->getStatusCode() === 200) {
            $data = $resp->toArray();
            if (!empty($data)) {
                return [
                    'price' => $data['price'],
                    'time' => $end,
                ];
            }
        }

        return [];
    }

    protected function saveToMySQL(string $symbol, array $data, \DateTimeImmutable $start, \DateTimeImmutable $end): void
    {
        $sql = <<<SQL
            INSERT INTO latest_rates (pair, rate, time)
            VALUES (:pair, :rate, :time)
            ON DUPLICATE KEY UPDATE
                rate = VALUES(rate)
        SQL;

        // we always use $end param instead of $data time
        // to make sure we save the right 5m frame
        $this->db->executeStatement($sql, [
            'pair' => $symbol,
            'rate' => $data['price'],
            'time' => $end->format('Y-m-d H:i:s'),
        ]);
    }
}
