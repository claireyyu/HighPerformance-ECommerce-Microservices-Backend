import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Custom metrics
const asyncDuration = new Trend('async_duration');
const asyncErrors = new Rate('async_errors');
const throughput = new Counter('throughput'); 

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '1m', target: 200 },
    { duration: '1m', target: 300 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'async_duration': ['p(95)<150'],
    'async_errors': ['rate<0.01'],
    'throughput': ['count>0'], // Optional: make sure some requests were successful
  }
};

const ORDER_SERVICE_URL = 'http://localhost:8082';

export default function () {
  const orderData = {
    user_id: 1,
    items: [{ product_id: 1, quantity: 1, price: 99.99 }]
  };

  const res = http.post(`${ORDER_SERVICE_URL}/orders/async?queue=kafka`, JSON.stringify(orderData), {
    headers: { 'Content-Type': 'application/json' }
  });

  asyncDuration.add(res.timings.duration);

  const ok = check(res, { 'status is 202': (r) => r.status === 202 });
  asyncErrors.add(!ok);

  if (ok) {
    throughput.add(1); 
  }
}

export function handleSummary(data) {
  return {
    "tests/results/kafka_tuning/summary.html": htmlReport(data),
    "tests/results/kafka_tuning/summary.txt": textSummary(data),
  };
}