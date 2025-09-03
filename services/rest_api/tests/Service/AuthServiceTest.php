<?php

namespace App\Tests\Service;

use App\Service\AuthService;
use PHPUnit\Framework\TestCase;
use Symfony\Contracts\HttpClient\HttpClientInterface;
use Symfony\Contracts\HttpClient\ResponseInterface;
use Symfony\Component\HttpFoundation\Session\SessionInterface;

class AuthServiceTest extends TestCase
{
    public function testSetApiKeyAndGetServerUrl()
    {
        $http = $this->createMock(HttpClientInterface::class);
        $service = new AuthService('https://api.example.com/', $http);

        $this->assertEquals('https://api.example.com', $service->getServerUrl());

        $service->setApiKey('abc123');
        $this->assertTrue(true); // just confirming no exception
    }

    public function testGetSavedJwtReturnsNullIfExpired()
    {
        $http = $this->createMock(HttpClientInterface::class);
        $session = $this->createMock(SessionInterface::class);

        $session->method('get')->willReturnMap([
            ['jwtExpiresAt', time() - 100, null],
            ['jwt', 'fake.jwt.token'],
        ]);

        $service = new AuthService('https://api.example.com', $http);
        $service->setSession($session);

        $this->assertNull($service->getSavedJwt());
    }

    public function testRequestJwtReturnsToken()
    {
        $http = $this->createMock(HttpClientInterface::class);
        $response = $this->createMock(ResponseInterface::class);

        $response->method('toArray')->willReturn(['jwt' => 'abc.def.ghi']);
        $http->method('request')->willReturn($response);

        $service = new AuthService('https://api.example.com', $http);
        $service->setApiKey('abc123');

        $jwt = $this->invokePrivateMethod($service, 'requestJwt');
        $this->assertEquals('abc.def.ghi', $jwt);
    }

    private function generateFakeJwt(): string
    {
        $header = base64_encode(json_encode(['alg' => 'HS256', 'typ' => 'JWT']));
        $payload = base64_encode(json_encode(['exp' => time() + 3600]));
        $signature = base64_encode('signature');

        return "$header.$payload.$signature";
    }

    private function invokePrivateMethod($object, string $methodName, array $args = [])
    {
        $ref = new \ReflectionClass($object);
        $method = $ref->getMethod($methodName);
        $method->setAccessible(true);

        return $method->invokeArgs($object, $args);
    }
}
