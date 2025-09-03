<?php
namespace App\Command;

use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Predis\Client as RedisClient;
use Doctrine\DBAL\Connection;
use Symfony\Contracts\HttpClient\HttpClientInterface;
use Psr\Log\LoggerInterface;

abstract class RatesCollectCommandBase extends Command
{
    protected ?string $nameForLog = null;

    public function __construct(
        protected RedisClient $redis,
        protected Connection $db,
        protected HttpClientInterface $http,
        protected LoggerInterface $logger,
        protected array $symbols
    ) {
        parent::__construct();
    }

    protected abstract function fetchFromRedis(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array;

    protected abstract function fetchFromBinance(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end): array;

    protected abstract function saveToMySQL(string $symbol, array $candles, \DateTimeImmutable $start, \DateTimeImmutable $end): void;

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        list($start, $end) = $this->fiveMinWindowStartEnd();
        $this->logger->info("Collecting rates for {$this->getNameForLog($start, $end)}");

        foreach ($this->symbols as $symbol) {
            $data = $this->fetchFromRedis($symbol, $start, $end);

            if (empty($data)) {
                $this->logger->info("{$this->getNameForLog($start, $end)}: no redis data, trying binance api");
                $data = $this->fetchFromBinanceWithRetry($symbol, $start, $end);
            }

            if (!empty($data)) {
                $this->logger->info("{$this->getNameForLog($start, $end)}: saving");
                $output->writeln("{$this->getNameForLog($start, $end)}: saving");

                $this->saveToMySQL($symbol, $data, $start, $end);
            } else {
                $this->logger->info("{$this->getNameForLog($start, $end)}: no data");
                $output->writeln("{$this->getNameForLog($start, $end)}: no data");
            }
        }

        return Command::SUCCESS;
    }

    private function fetchFromBinanceWithRetry(string $symbol, \DateTimeImmutable $start, \DateTimeImmutable $end, int $maxRetries = 3): array
    {
        $attempts = 0;
        while ($attempts < $maxRetries) {
            try {
                return $this->fetchFromBinance($symbol, $start, $end);
            } catch (\Throwable $e) {
                $attempts++;
                if ($attempts >= $maxRetries) {
                    return [];
                }

                sleep(1);
            }
        }

        return [];
    }

    private function fiveMinWindowStartEnd(): array
    {
        $now = new \DateTimeImmutable('now', new \DateTimeZone('UTC'));

        // Truncate minutes to nearest multiple of 5
        $minute = (int) $now->format('i');
        $truncatedMinute = intdiv($minute, 5) * 5;

        $alignedNow = $now->setTime(
            (int) $now->format('H'),
            $truncatedMinute,
            0
        );
        $start = $alignedNow->modify('-5 minutes');

        return [$start, $alignedNow];
    }

    protected function getNameForLog(\DateTimeImmutable $start, \DateTimeImmutable $end) {
        if (!$this->nameForLog) {
            $this->nameForLog = "{$this->getName()}[{$start->format('Y-m-d H:i:s')}, {$end->format('Y-m-d H:i:s')}]";
        }

        return $this->nameForLog;
    }
}
