<?php

namespace App\Tests\EventListener;

use App\Attribute\RequireAuth;
use App\EventListener\AuthListener;
use App\Service\AuthService;
use PHPUnit\Framework\TestCase;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Session\SessionInterface;
use Symfony\Component\HttpKernel\Event\ControllerEvent;
use Symfony\Component\HttpKernel\HttpKernelInterface;

#[RequireAuth]
class DummyController
{
    #[RequireAuth]
    public function protectedAction() {}
}

class AuthListenerTest extends TestCase
{
    public function testSkipsIfNoRequireAuth()
    {
        $authService = $this->createMock(AuthService::class);
        $listener = new AuthListener($authService);

        $controller = [new class {
            public function index() {}
        }, 'index'];

        $request = new Request();
        $event = new ControllerEvent(
            $this->createMock(HttpKernelInterface::class),
            $controller,
            $request,
            HttpKernelInterface::MAIN_REQUEST
        );

        $listener->onKernelController($event);

        $this->assertSame($controller, $event->getController());
    }

    public function testRejectsMissingAuthorizationHeader()
    {
        $authService = $this->createMock(AuthService::class);
        $listener = new AuthListener($authService);

        $controller = [new DummyController(), 'protectedAction'];
        $request = new Request(); // no Authorization header

        $event = new ControllerEvent(
            $this->createMock(HttpKernelInterface::class),
            $controller,
            $request,
            HttpKernelInterface::MAIN_REQUEST
        );

        $listener->onKernelController($event);

        $response = call_user_func($event->getController());
        $event->setController(function () {
            return new JsonResponse(['error' => 'Authorization Bearer token is required'], 401);
        });
        $this->assertInstanceOf(\Symfony\Component\HttpFoundation\JsonResponse::class, $response);
        $this->assertEquals(401, $response->getStatusCode());
    }

    public function testRejectsJwtMismatch()
    {
        $authService = $this->createMock(AuthService::class);
        $session = $this->createMock(SessionInterface::class);

        $authService->method('getSavedJwt')->willReturn('valid.jwt.token');
        $authService->expects($this->once())->method('setSession');

        $listener = new AuthListener($authService);

        $controller = [new DummyController(), 'protectedAction'];
        $request = new Request();
        $request->headers->set('Authorization', 'Bearer invalid.jwt.token');
        $request->setSession($session);

        $event = new ControllerEvent(
            $this->createMock(HttpKernelInterface::class),
            $controller,
            $request,
            HttpKernelInterface::MAIN_REQUEST
        );

        $listener->onKernelController($event);

        $response = call_user_func($event->getController());
        $this->assertEquals(401, $response->getStatusCode());
    }

    public function testAllowsValidJwt()
    {
        $authService = $this->createMock(AuthService::class);
        $session = $this->createMock(SessionInterface::class);

        $authService->method('getSavedJwt')->willReturn('valid.jwt.token');
        $authService->expects($this->once())->method('setSession');

        $listener = new AuthListener($authService);

        $controller = [new DummyController(), 'protectedAction'];
        $request = new Request();
        $request->headers->set('Authorization', 'Bearer valid.jwt.token');
        $request->setSession($session);

        $event = new ControllerEvent(
            $this->createMock(HttpKernelInterface::class),
            $controller,
            $request,
            HttpKernelInterface::MAIN_REQUEST
        );

        $listener->onKernelController($event);

        $this->assertSame($controller, $event->getController());
    }
}
