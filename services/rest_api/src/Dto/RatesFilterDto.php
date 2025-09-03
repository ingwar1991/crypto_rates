<?php

namespace App\Dto;

use Symfony\Component\Validator\Constraints as Assert;
use App\Validator\Constraints as AppAssert;

class RatesFilterDto
{
    #[Assert\NotBlank(message: "pair is required")]
    #[Assert\All([
        new Assert\NotBlank(message: "pair element cannot be blank"),
        new Assert\Length(max: 7, maxMessage: "longest trading pair in crypto markets are 7 symbols"),
    ])]
    #[Assert\NotNull]
    #[AppAssert\ValidPair]
    public array $pair = [];

    #[Assert\NotBlank(message: "date is required")]
    #[Assert\Date(message: "date must be in YYYY-MM-DD format")]
    public ?string $date;

    #[Assert\Positive(message: "page must be greater than 0")]
    public int $page = 1;

    #[Assert\Range(min: 1, max: 500, notInRangeMessage: "limit must be between {{ min }} and {{ max }}")]
    public int $limit = 50;
}
