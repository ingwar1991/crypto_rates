<?php

namespace App;

use Symfony\Component\Scheduler\Attribute\AsSchedule;
use Symfony\Component\Scheduler\Schedule as SymfonySchedule; use Symfony\Component\Scheduler\ScheduleProviderInterface;
use Symfony\Contracts\Cache\CacheInterface;
use Symfony\Component\Scheduler\RecurringMessage;
use Symfony\Component\Console\Messenger\RunCommandMessage;


#[AsSchedule]
class Schedule implements ScheduleProviderInterface
{
    public function __construct(
        private CacheInterface $cache,
    ) {
    }

    public function getSchedule(): SymfonySchedule
    {
        return (new SymfonySchedule())
            ->add(
                RecurringMessage::cron(
                    '*/5 * * * *', // every 5 mins
                    new RunCommandMessage('rates:latest:collect'),
                ),
            )
            ->add(
                RecurringMessage::cron(
                    '*/5 * * * *', // every 5 mins
                    new RunCommandMessage('rates:avg:collect'),
                ),
            )
            ->stateful($this->cache) // ensure missed tasks are executed
            ->processOnlyLastMissedRun(true) // ensure only last missed task is run

            // add your own tasks here
            // see https://symfony.com/doc/current/scheduler.html#attaching-recurring-messages-to-a-schedule
        ;
    }
}
