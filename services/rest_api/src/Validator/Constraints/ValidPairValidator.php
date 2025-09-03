<?php

namespace App\Validator\Constraints;

use Symfony\Component\Validator\Constraint;
use Symfony\Component\Validator\ConstraintValidator;

class ValidPairValidator extends ConstraintValidator
{
    public function __construct(private readonly array $symbols) {}

    public function validate($value, Constraint $constraint): void
    {
        if (!$constraint instanceof ValidPair) {
            return;
        }

        if ($value === null) {
            return;
        }

        foreach ((array)$value as $pair) {
            if (!in_array($pair, $this->symbols, true)) {
                $this->context->buildViolation($constraint->message)
                    ->setParameter('{{ pair }}', $pair)
                    ->addViolation();
            }
        }
    }
}
