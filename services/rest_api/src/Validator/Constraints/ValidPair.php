<?php

namespace App\Validator\Constraints;

use Symfony\Component\Validator\Constraint;

#[\Attribute]
class ValidPair extends Constraint
{
    public string $message = 'The pair "{{ pair }}" is not allowed.';

    public function __construct(
        public array $allowed = [],
        array $options = [],
        array $groups = null,
        mixed $payload = null
    ) {
        parent::__construct($options, $groups, $payload);
    }
}
