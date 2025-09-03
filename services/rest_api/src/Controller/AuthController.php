<?php

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\Routing\Annotation\Route;
use App\Service\AuthService;

class AuthController extends AbstractController
{
    private AuthService $authService;

    public function __construct(AuthService $authService)
    {
        $this->authService = $authService;
    }

    #[Route('/auth', name: 'auth', methods: ['POST'])]
    public function authenticate(Request $request): JsonResponse
    {
        $apiKey = $request->request->get('api_key');
        if (!$apiKey) {
            return new JsonResponse(['error' => 'api_key is required'], 400);
        }

        $this->authService->setApiKey($apiKey);
        $this->authService->setSession($request->getSession());
        $jwt = $this->authService->getJwt();

        if (!$jwt) {
            return new JsonResponse(['error' => 'failed to authenticate'], 404);
        }

        return new JsonResponse([
            'status' => 'success',
            'jwt' => $jwt,
        ]);
    }
}
