import http from 'k6/http';
import { check } from 'k6';

export function setup() {
    const res = http.post(
        'http://localhost:8080/api/shorten',
        JSON.stringify({ url: 'https://google.com/cache-test' }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    const code = JSON.parse(res.body).short_url.split('/').pop();
    return { code };
}

export const options = {
    stages: [
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
    ],
};

export default function (data) {
    const res = http.get(`http://localhost:8080/${data.code}`, {
        redirects: 0,
    });

    check(res, {
        'status 301': (r) => r.status === 301,
    });
}

export function handleSummary(data) {
    return {
        'tests/load/resolve_summary.json': JSON.stringify({
            checks_succeeded: data.metrics.checks.values.passes,
            checks_failed: data.metrics.checks.values.fails,
            avg_duration: data.metrics.http_req_duration.values.avg,
            p95_duration: data.metrics.http_req_duration.values['p(95)'],
            rps: data.metrics.http_reqs.values.rate,
            error_rate: data.metrics.http_req_failed.values.rate,
        }, null, 2),
    };
}