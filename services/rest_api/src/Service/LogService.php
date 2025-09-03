<?php

namespace App\Service;

use Symfony\Contracts\HttpClient\HttpClientInterface;
use Symfony\Component\HttpFoundation\Session\SessionInterface;

class LogService
{
    public function __construct(
        private AuthService $authService,
        private HttpClientInterface $http,
    ) {}

    public function setSession(SessionInterface $session)
    {
        $this->authService->setSession($session);
    }

    public function log(string $endpoint, int $responseStatus, ?array $params)
    {
        if (!strlen($endpoint) || $responseStatus < 0) {
            return;
        }

        $this->authService->authedRequest('/log/rest', [
            'endpoint' => $endpoint,
            'response_status' => $responseStatus,
            'params' => !empty($params)
                ? $params
                : null,
            'timestamp' => time(),
        ]);
    }
}
