import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    reserve_war: {
      executor: 'constant-arrival-rate',
      rate: __ENV.RATE ? Number(__ENV.RATE) : 200,
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      preAllocatedVUs: __ENV.PRE_VUS ? Number(__ENV.PRE_VUS) : 200,
      maxVUs: __ENV.MAX_VUS ? Number(__ENV.MAX_VUS) : 2000,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<200'],
    checks: ['rate>0.99'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const USER_TOKEN = __ENV.USER_TOKEN;
const EVENT_ID = __ENV.EVENT_ID;
const CATEGORY = __ENV.CATEGORY || 'VIP';

if (!USER_TOKEN || !EVENT_ID) {
  throw new Error('Set USER_TOKEN and EVENT_ID env vars before running k6');
}

export default function () {
  const payload = JSON.stringify({
    event_id: EVENT_ID,
    category: CATEGORY,
    qty: 1,
  });

  const res = http.post(`${BASE_URL}/reserve`, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${USER_TOKEN}`,
    },
  });

  check(res, {
    'status is 201 or 409/429': (r) => [201, 409, 429].includes(r.status),
  });

  sleep(0.05);
}
