import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const successfulCreates = new Counter('successful_creates');
const failedCreates = new Counter('failed_creates');
const successfulGets = new Counter('successful_gets');
const failedGets = new Counter('failed_gets');

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
    // Spike test scenario
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '1m', target: 500 },  // Quick ramp up to 500 users
        { duration: '30s', target: 500 }, // Stay at 500 users
        { duration: '1m', target: 0 },    // Quick ramp down to 0
      ],
      gracefulRampDown: '30s',
    }
  },
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Less than 1% of requests should fail
  },
};

const BASE_URL = 'http://localhost:8080'; // API Gateway URL

// Helper function to generate test product data
function generateProductData() {
  return {
    name: `Test Product ${randomString(8)}`,
    description: `Test Description ${randomString(16)}`,
    price: Math.floor(Math.random() * 1000) + 1,
    stock: Math.floor(Math.random() * 100) + 1
  };
}

export default function() {
  // Create Product Test
  const productData = generateProductData();
  const createResponse = http.post(`${BASE_URL}/products`, JSON.stringify(productData), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(createResponse, {
    'create product status is 200': (r) => r.status === 200,
  }) ? successfulCreates.add(1) : failedCreates.add(1);

  // Extract product ID from response
  let productId;
  try {
    productId = JSON.parse(createResponse.body).id;
  } catch (e) {
    console.log('Failed to parse create response:', e);
    return;
  }

  sleep(1); // Wait 1 second between requests

  // Get Product Test
  const getResponse = http.get(`${BASE_URL}/products/${productId}`);
  
  check(getResponse, {
    'get product status is 200': (r) => r.status === 200,
    'get product returns correct data': (r) => {
      const body = JSON.parse(r.body);
      return body.name === productData.name;
    },
  }) ? successfulGets.add(1) : failedGets.add(1);

  sleep(1);

  // List Products Test
  const listResponse = http.get(`${BASE_URL}/products`);
  
  check(listResponse, {
    'list products status is 200': (r) => r.status === 200,
    'list products returns array': (r) => Array.isArray(JSON.parse(r.body)),
  });

  sleep(1);

  // Update Product Test
  const updateData = {
    ...productData,
    name: `Updated ${productData.name}`,
    price: productData.price + 100
  };
  
  const updateResponse = http.put(
    `${BASE_URL}/products/${productId}`,
    JSON.stringify(updateData),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(updateResponse, {
    'update product status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // Delete Product Test
  const deleteResponse = http.del(`${BASE_URL}/products/${productId}`);
  
  check(deleteResponse, {
    'delete product status is 200': (r) => r.status === 200,
  });
} 