import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
    stages: [
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {
    const res = http.post(
        `${BASE_URL}/api/shorten`,
        JSON.stringify({ url: `https://google.com/${Math.random()}` }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    check(res, {
        'status 201': (r) => r.status === 201,
    });
}

export function handleSummary(data) {
    return {
        'tests/load/stress_summary.json': JSON.stringify({
            checks_succeeded: data.metrics.checks.values.passes,
            checks_failed: data.metrics.checks.values.fails,
            avg_duration: data.metrics.http_req_duration.values.avg,
            p95_duration: data.metrics.http_req_duration.values['p(95)'],
            rps: data.metrics.http_reqs.values.rate,
            error_rate: data.metrics.http_req_failed.values.rate,
        }, null, 2),
    };
}