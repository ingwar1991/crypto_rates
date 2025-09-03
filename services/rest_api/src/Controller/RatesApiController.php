<?php

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\RedirectResponse;
use Symfony\Component\Validator\Validator\ValidatorInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\Routing\Annotation\Route;
use Doctrine\DBAL\Connection;
use App\Attribute\RequireAuth;
use App\Dto\RatesFilterDto;
use App\Service\LogService;
use Symfony\Component\Validator\ConstraintViolationList;
use Symfony\Component\Validator\ConstraintViolation;

#[RequireAuth]
#[Route('/api/rates')]
class RatesApiController extends AbstractController
{
    public function __construct(
        private Connection $db,
        private LogService $log,
    ) {}

    private function getRequestParamsAndValidate(Request $request, ValidatorInterface $validator, bool $withDefaultDate = true): array
    {
        $dto = new RatesFilterDto();
        $dto->pair  = $request->query->all('pair');
        // if no date tramitted - set it to today
        $defaultDate = $withDefaultDate
            ? (new \DateTimeImmutable())->format('Y-m-d')
            : null;
        $dto->date  = $request->query->get('date', $defaultDate);
        $dto->page  = (int)$request->query->get('page', 1);
        $dto->limit = (int)$request->query->get('limit', 100);

        return [
            // params
            [
                'pair' => $dto->pair,
                'date' => $dto->date,
                'limit' => $dto->limit,
                'offset' => ($dto->page - 1) * $dto->limit,
            ],
            // error
            $validator->validate($dto),
        ];
    }

    #[Route('/latest', name: 'api_rates_latest',  methods: ['GET'])]
    public function latestRates(Request $request, ValidatorInterface $validator): JsonResponse
    {
        list($params, $errors) = $this->getRequestParamsAndValidate($request, $validator);
        if (count($errors) > 0) {
            return $this->renderJson($request, $this->prepareErrorsForResponse($errors), 400, $params);
        }

        $qb = $this->db->createQueryBuilder()
            ->select('SQL_CALC_FOUND_ROWS *')
            ->from('latest_rates')
            ->where('DATE(time) = :date')
            ->setParameter('date', $params['date'])
            ->orderBy('time', 'asc')
            ->setFirstResult($params['offset'])
            ->setMaxResults($params['limit']);

        $qb->andWhere('pair IN (:pairs)')
           ->setParameter('pairs', $params['pair'], Connection::PARAM_STR_ARRAY);

        $rows = $qb->executeQuery()->fetchAllAssociative();
        $this->transformResponseData($rows);

        return $this->renderJson(
            $request,
            [
                'total_found'  => $this->db->fetchOne('SELECT FOUND_ROWS()'),
                'data'   => $rows,
            ],
            200,
            $params,
        );
    }

    #[Route('/avg', methods: ['GET'])]
    public function avgRates(Request $request, ValidatorInterface $validator): JsonResponse
    {
        list($params, $errors) = $this->getRequestParamsAndValidate($request, $validator);
        if (count($errors) > 0) {
            return $this->renderJson($request, $this->prepareErrorsForResponse($errors), 400, $params);
        }

        $qb = $this->db->createQueryBuilder()
            ->select('SQL_CALC_FOUND_ROWS *')
            ->from('avg_rates')
            ->where('DATE(start_time) = :date')
            ->setParameter('date', $params['date'])
            ->orderBy('start_time', 'asc')
            ->setFirstResult($params['offset'])
            ->setMaxResults($params['limit']);

        $qb->andWhere('pair IN (:pairs)')
           ->setParameter('pairs', $params['pair'], Connection::PARAM_STR_ARRAY);

        $rows = $qb->executeQuery()->fetchAllAssociative();
        $this->transformResponseData($rows);

        return $this->renderJson(
            $request,
            [
                'total_found'  => $this->db->fetchOne('SELECT FOUND_ROWS()'),
                'data'   => $rows,
            ],
            200,
            $params,
        );
    }

    #[Route('/last-24h', methods: ['GET'])]
    public function latestRates24h(Request $request): RedirectResponse
    {
        // add cur date param
        $queryParams = $request->query->all();
        $queryParams['date'] = (new \DateTimeImmutable())->format('Y-m-d');

        return $this->redirect(
            $this->generateUrl('api_rates_latest_day', $queryParams)
        );
    }

    #[Route('/day', name: 'api_rates_latest_day', methods: ['GET'])]
    public function latestRatesDay(Request $request, ValidatorInterface $validator): Response
    {
        list($params, $errors) = $this->getRequestParamsAndValidate($request, $validator, false);
        if (count($errors) > 0) {
            return $this->renderJson($request, $errors, 400, $params);
        }

        $queryParams = $request->query->all();

        return $this->redirect(
            $this->generateUrl('api_rates_latest', $queryParams)
        );
    }

    private function transformResponseData(array &$data)
    {
        $transformed = [];
        foreach($data as $entry) {
            if (!isset($transformed[$entry['pair']])) {
                $transformed[$entry['pair']] = [];
            }

            $transformed[$entry['pair']][] = [
                'rate' => $entry['rate'],
                'time' => $entry['time'],
            ];
        }

        $data = $transformed;
    }

    private function prepareErrorsForResponse(ConstraintViolationList $errors): array
    {
        return array_map(fn(ConstraintViolation $error) => $error->getMessage(), iterator_to_array($errors));
    }

    private function renderJson(Request $request, mixed $data, int $status = 200, array $params = []): JsonResponse
    {
        $this->log->setSession($request->getSession());
        $this->log->log(
            $request->get('_route'),
            $status,
            $params,
        );

        return $this->json($data, $status);
    }
}
