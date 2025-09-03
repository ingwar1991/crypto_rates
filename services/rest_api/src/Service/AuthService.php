<?php

namespace App\Service;

use Symfony\Contracts\HttpClient\HttpClientInterface;
use Symfony\Contracts\HttpClient\Exception\TransportExceptionInterface;
use Symfony\Component\HttpFoundation\Session\SessionInterface;

class AuthService
{
    private ?string $apiKey = null;
    private ?SessionInterface $session;

    public function __construct(
        private string $serverUrl,
        private HttpClientInterface $http,
    )
    {
        $this->serverUrl = rtrim($serverUrl, '/');
    }

    public function setApiKey(string $apiKey)
    {
        $this->apiKey = $apiKey;
    }

    public function setSession(SessionInterface $session)
    {
        $this->session = $session;
    }

    private function getSession()
    {
        if (empty($this->session)) {
            throw new \Exception("[AuthService:getSession(): no session]");
        }

        return $this->session;
    }

    private function getSavedJwtExpiresAt(): ?string
    {
        return $this->getSession()->get('jwtExpiresAt');
    }

    private function checkSavedJwtExpiration(): bool
    {
        $expiresAt = $this->getSavedJwtExpiresAt();
        if (!empty($expiresAt)) {
            return $expiresAt > time();
        }

        return false;
    }

    public function getSavedJwt(): ?string
    {
        return $this->checkSavedJwtExpiration()
            ? $this->getSession()->get('jwt')
            : null;
    }

    public function getJwt(): ?string
    {
        if (!$this->apiKey) {
            throw new \Exception("[AuthService:getJwt(): no api_key]");
        }

        if ($this->checkSavedJwtExpiration()) {
            // refresh if it's one min before expire
            if ($this->getSavedJwtExpiresAt() - 60 <= time()) {
                $jwt = $this->requestJwtRefresh();
                if (!$jwt) {
                    return null;
                }

                $this->setJwt($jwt);
            }

            return $this->getSavedJwt();
        }

        $jwt = $this->requestJwt();
        if (!$jwt) {
            return null;
        }
        $this->setJwt($jwt);

        return $this->getSavedJwt();
    }

    private function setJwt(string $jwt) {
        // Decode JWT to get expiry
        $parts = explode('.', $jwt);
        if (count($parts) !== 3) {
            throw new \Exception("[AuthService:setJwt()] Wrong jwt format");
        }

        $payload = json_decode(base64_decode($parts[1]), true);
        if (!$payload || empty($payload['exp'])) {
            throw new \Exception("[AuthService:setJwt()] Wrong jwt exp");
        }

        $this->getSession()->set('jwt', $jwt);
        $this->getSession()->set('jwtExpiresAt', $payload['exp']);
    }

    private function requestJwt(): ?string {
        if (!$this->apiKey) {
            throw new \Exception("[AuthService:getJwt()] no api_key]");
        }

        $url = $this->serverUrl . '/token';
        try {
            $response = $this->http->request('POST', $url, [
                'body' => [
                    'api_key' => $this->apiKey,
                ],
                'timeout' => 5,
            ]);

            $json = $response->toArray();
        } catch (TransportExceptionInterface $e) {
            return null;
        }

        if (!isset($json['jwt'])) {
            return null;
        }

        return $json['jwt'];
    }

    private function requestJwtRefresh(): ?string {
        $resp = $this->authedRequest('/token/refresh');

        return $resp && $resp['jwt']
            ? $resp['jwt']
            : null;
    }

    public function authedRequest(string $endpoint, ?array $body = null): ?array
    {
        $jwt = $this->getSavedJwt();
        if (!$jwt) {
            return null;
        }

        $url = $this->serverUrl . $endpoint;

        $method = 'GET';
        $data = [
            'headers' => [
                'Authorization' => "Bearer {$jwt}",
            ],
            'timeout' => 5,
        ];
        if ($body !== null) {
            $method = 'POST';
            $data['headers']['Content-Type'] = 'application/json';
            $data['body'] = json_encode($body);
        }

        try {
            $response = $this->http->request($method, $url, $data);

            return $response->toArray();
        } catch (TransportExceptionInterface $e) {
            return null;
        }
    }

    public function getServerUrl(): string
    {
        return $this->serverUrl;
    }
}
