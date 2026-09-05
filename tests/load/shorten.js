import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {
    // тест shorten
    const shortenRes = http.post(
        'http://localhost:8080/api/shorten',
        JSON.stringify({ url: `https://google.com/${Math.random()}` }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    check(shortenRes, {
        'shorten status 201': (r) => r.status === 201,
        'shorten has short_url': (r) => JSON.parse(r.body).short_url !== '',
    });

    sleep(1);
}