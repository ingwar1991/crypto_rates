<?php

namespace App\EventListener;

use App\Attribute\RequireAuth;
use Symfony\Component\HttpKernel\Event\ControllerEvent;
use Symfony\Component\HttpFoundation\JsonResponse;
use App\Service\AuthService;
use ReflectionClass;
use ReflectionMethod;

class AuthListener
{
    private AuthService $authService;

    public function __construct(AuthService $authService)
    {
        $this->authService = $authService;
    }

    public function onKernelController(ControllerEvent $event)
    {
        $controller = $event->getController();

        if (!is_array($controller)) {
            return;
        }

        [$controllerObject, $methodName] = $controller;

        $reflectionMethod = new ReflectionMethod($controllerObject, $methodName);
        $reflectionClass  = new ReflectionClass($controllerObject);

        $hasAttribute = $reflectionMethod->getAttributes(RequireAuth::class) ||
                        $reflectionClass->getAttributes(RequireAuth::class);

        if ($hasAttribute) {
            $request = $event->getRequest();
            $authorizationHeader = $request->headers->get('Authorization');

            // Check if Authorization header exists and is in Bearer token format
            $matches = [];
            if (!$authorizationHeader || !preg_match('/Bearer\s(\S+)/', $authorizationHeader, $matches)) {
                $event->setController(function () {
                    return new JsonResponse(['error' => 'Authorization Bearer token is required'], 401);
                });

                return;
            }
            $jwtCurr = !empty($matches[1])
                ? $matches[1]
                : null;
            if (!$jwtCurr) {
                $event->setController(function() {
                    return new JsonResponse(['error' => 'Unauthorized'], 401);
                });

                return;
            }

            // set session to authService
            $this->authService->setSession($request->getSession());

            $jwtSaved = $this->authService->getSavedJwt();
            if (!$jwtSaved || $jwtCurr != $jwtSaved) {
                $event->setController(function() {
                    return new JsonResponse(['error' => 'Unauthorized'], 401);
                });
            }
        }
    }
}
