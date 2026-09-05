import http from 'k6/http';
import { check } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {
    const res = http.post(
        'http://localhost:8080/api/shorten',
        JSON.stringify({ url: `https://google.com/${Math.random()}` }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    check(res, {
        'status 201': (r) => r.status === 201,
    });
}