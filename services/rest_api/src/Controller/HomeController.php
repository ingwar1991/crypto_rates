<?php

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class HomeController extends AbstractController
{
    public function __construct(
        private readonly string $apiKey,
        private readonly array $symbols,
    ) {}

    #[Route('/', name: 'home')]
    public function index(): Response
    {
        return $this->render('home.html.twig', [
            'apiKey' => $this->apiKey,
            'symbols' => $this->symbols,
        ]);
    }
}
