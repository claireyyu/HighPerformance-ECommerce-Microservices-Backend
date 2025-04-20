import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const syncDuration = new Trend('sync_duration');
const asyncDuration = new Trend('async_duration');
const syncErrors = new Rate('sync_errors');
const asyncErrors = new Rate('async_errors');
const productErrors = new Rate('product_errors'); // ✅ NEW

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 50 },   // Warm-up
    { duration: '1m', target: 200 },    // Light load
    { duration: '1m', target: 400 },    // Medium load
    { duration: '1m', target: 450 },    // Heavy load
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'],
    'errors': ['rate<0.01'],
    'sync_duration': ['p(95)<500'],
    'async_duration': ['p(95)<100'],
    'sync_errors': ['rate<0.01'],
    'async_errors': ['rate<0.01'],
    'product_errors': ['rate<0.01'], 
  },
};

// Use correct ports for services
const PRODUCT_SERVICE_URL = 'http://localhost:8081';
const ORDER_SERVICE_URL = 'http://localhost:8082';

// Helper function to generate test product data
function generateProductData() {
  return {
    name: `Test Product ${Date.now()}`,
    description: 'A test product for performance testing',
    price: Math.floor(Math.random() * 100) + 1,
    stock: Math.floor(Math.random() * 100) + 1,
  };
}

// Helper function to generate test order data
function generateOrderData(productId, price) {
  return {
    user_id: 1,
    items: [
      {
        product_id: productId,
        quantity: 1,
        price: price,
      },
    ],
  };
}

export default function() {
  // Create a product first
  const productData = generateProductData();
  const productRes = http.post(`${PRODUCT_SERVICE_URL}/products`, JSON.stringify(productData), {
    headers: { 'Content-Type': 'application/json' },
  });

  const productCheck = check(productRes, {
    'product creation successful': (r) => r.status === 201,
  });

  if (!productCheck) {
    errorRate.add(1);
    productErrors.add(1); 
    return;
  } else {
    productErrors.add(0);
  }

  const product = JSON.parse(productRes.body);

  // Create order synchronously
  group('Synchronous Order Creation', function() {
    const syncOrderData = generateOrderData(product.id, product.price);
    const syncRes = http.post(`${ORDER_SERVICE_URL}/orders`, JSON.stringify(syncOrderData), {
      headers: { 'Content-Type': 'application/json' },
    });

    syncDuration.add(syncRes.timings.duration);
    const syncCheck = check(syncRes, {
      'sync order creation successful': (r) => r.status === 201,
    });

    if (!syncCheck) {
      syncErrors.add(1);
      errorRate.add(1);
    } else {
      syncErrors.add(0);
    }
  });

  // Create order asynchronously
  group('Asynchronous Order Creation', function() {
    const asyncOrderData = generateOrderData(product.id, product.price);
    const asyncRes = http.post(`${ORDER_SERVICE_URL}/orders/async`, JSON.stringify(asyncOrderData), {
      headers: { 'Content-Type': 'application/json' },
    });

    asyncDuration.add(asyncRes.timings.duration);
    const asyncCheck = check(asyncRes, {
      'async order creation successful': (r) => r.status === 202,
    });

    if (!asyncCheck) {
      asyncErrors.add(1);
      errorRate.add(1);
    } else {
      asyncErrors.add(0);
    }
  });

  sleep(1); // Sleep removed to maximize throughput
}

// Generate both HTML and text summary reports
export function handleSummary(data) {
  return {
    "tests/results/load_test/summary.html": htmlReport(data),
    "tests/results/load_test/summary.txt": textSummary(data),
  };
}
