<?php

namespace App\Tests\Validator\Constraints;

use App\Validator\Constraints\ValidPair;
use App\Validator\Constraints\ValidPairValidator;
use Symfony\Component\Validator\Test\ConstraintValidatorTestCase;

class ValidPairValidatorTest extends ConstraintValidatorTestCase
{
    protected function createValidator(): ValidPairValidator
    {
        return new ValidPairValidator(['BTCEUR', 'ETHEUR']);
    }

    public function testValidPairsPassValidation(): void
    {
        $constraint = new ValidPair();

        $this->validator->validate(['BTCEUR', 'ETHEUR'], $constraint);

        $this->assertNoViolation();
    }

    public function testSingleInvalidPairTriggersViolation(): void
    {
        $constraint = new ValidPair();

        $this->validator->validate(['DOGEEUR'], $constraint);

        $this->buildViolation($constraint->message)
            ->setParameter('{{ pair }}', 'DOGEEUR')
            ->assertRaised();
    }

    public function testMixedPairsTriggersViolationForInvalid(): void
    {
        $constraint = new ValidPair();

        $this->validator->validate(['BTCEUR', 'DOGEEUR'], $constraint);

        $this->buildViolation($constraint->message)
            ->setParameter('{{ pair }}', 'DOGEEUR')
            ->assertRaised();
    }

    public function testNullValueDoesNotTriggerViolation(): void
    {
        $constraint = new ValidPair();

        $this->validator->validate(null, $constraint);

        $this->assertNoViolation();
    }

    public function testEmptyArrayDoesNotTriggerViolation(): void
    {
        // $this->assertNull('blo');
        $constraint = new ValidPair();

        $this->validator->validate([], $constraint);

        $this->assertNoViolation();
    }
}
