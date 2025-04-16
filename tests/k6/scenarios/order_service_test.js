import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const successfulOrderCreates = new Counter('successful_order_creates');
const failedOrderCreates = new Counter('failed_order_creates');
const successfulAsyncCreates = new Counter('successful_async_creates');
const failedAsyncCreates = new Counter('failed_async_creates');

// Test configuration
export const options = {
  scenarios: {
    // Load test scenario
    load_test: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '1m', target: 50 },  // Ramp up to 50 users
        { duration: '3m', target: 50 },  // Stay at 50 users
        { duration: '1m', target: 0 },   // Ramp down to 0
      ],
      gracefulRampDown: '30s',
    },
    // Stress test scenario
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 50,
      stages: [
        { duration: '2m', target: 100 }, // Ramp up to 100 users
        { duration: '5m', target: 100 }, // Stay at 100 users
        { duration: '2m', target: 200 }, // Ramp up to 200 users
        { duration: '5m', target: 200 }, // Stay at 200 users
        { duration: '2m', target: 0 },   // Ramp down to 0
      ],
      gracefulRampDown: '30s',
    },
    // Spike test scenario for async orders
    async_spike_test: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 1000 },  // Quick ramp up to 1000 users
        { duration: '1m', target: 1000 },   // Stay at 1000 users
        { duration: '30s', target: 0 },     // Quick ramp down to 0
      ],
      gracefulRampDown: '30s',
    }
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests should be below 1000ms
    http_req_failed: ['rate<0.01'],    // Less than 1% of requests should fail
  },
};

const BASE_URL = 'http://localhost:8080'; // API Gateway URL

// Helper function to generate test order data
function generateOrderData() {
  return {
    user_id: Math.floor(Math.random() * 1000) + 1,
    items: [
      {
        product_id: Math.floor(Math.random() * 100) + 1,
        quantity: Math.floor(Math.random() * 5) + 1
      },
      {
        product_id: Math.floor(Math.random() * 100) + 1,
        quantity: Math.floor(Math.random() * 5) + 1
      }
    ]
  };
}

export default function() {
  // Create Order Test (Synchronous)
  const orderData = generateOrderData();
  const createResponse = http.post(`${BASE_URL}/orders`, JSON.stringify(orderData), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(createResponse, {
    'create order status is 200': (r) => r.status === 200,
  }) ? successfulOrderCreates.add(1) : failedOrderCreates.add(1);

  // Extract order ID from response
  let orderId;
  try {
    orderId = JSON.parse(createResponse.body).id;
  } catch (e) {
    console.log('Failed to parse create response:', e);
    return;
  }

  sleep(1);

  // Create Order Test (Asynchronous)
  const asyncOrderData = generateOrderData();
  const asyncCreateResponse = http.post(`${BASE_URL}/orders/async`, JSON.stringify(asyncOrderData), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(asyncCreateResponse, {
    'create async order status is 202': (r) => r.status === 202,
  }) ? successfulAsyncCreates.add(1) : failedAsyncCreates.add(1);

  sleep(1);

  // Get Order Test
  const getResponse = http.get(`${BASE_URL}/orders/${orderId}`);
  
  check(getResponse, {
    'get order status is 200': (r) => r.status === 200,
    'get order returns correct user_id': (r) => {
      const body = JSON.parse(r.body);
      return body.user_id === orderData.user_id;
    },
  });

  sleep(1);

  // List Orders Test
  const listResponse = http.get(`${BASE_URL}/orders?user_id=${orderData.user_id}`);
  
  check(listResponse, {
    'list orders status is 200': (r) => r.status === 200,
    'list orders returns array': (r) => Array.isArray(JSON.parse(r.body)),
  });

  sleep(1);

  // Update Order Status Test
  const updateStatusResponse = http.put(
    `${BASE_URL}/orders/${orderId}/status`,
    JSON.stringify({ status: 'paid' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(updateStatusResponse, {
    'update order status is 200': (r) => r.status === 200,
  });
} 