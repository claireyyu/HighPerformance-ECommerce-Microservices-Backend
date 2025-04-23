import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Custom metrics for each endpoint type
const syncDuration = new Trend('sync_duration');
const kafkaDuration = new Trend('kafka_duration');
const rabbitmqDuration = new Trend('rabbitmq_duration');

const syncErrors = new Rate('sync_errors');
const kafkaErrors = new Rate('kafka_errors');
const rabbitmqErrors = new Rate('rabbitmq_errors');

const syncThroughput = new Counter('sync_throughput');
const kafkaThroughput = new Counter('kafka_throughput');
const rabbitmqThroughput = new Counter('rabbitmq_throughput');

// Track total requests and successes for each endpoint and group
let testSummary = {};

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://52.12.143.55:8082';

const endpoints = {
  sync: `${BASE_URL}/orders`,
  kafka: `${BASE_URL}/orders/async?queue=kafka`,
  rabbitmq: `${BASE_URL}/orders/async?queue=rabbitmq`
};

// Test configuration to mimic the original test with thread groups
export const options = {
  scenarios: {
    // Group 1: 2 thread groups (100 VUs)
    sync_group1: {
      executor: 'constant-vus',
      vus: 100,
      duration: '30s',
      exec: 'syncTest',
      startTime: '0s',
      tags: { endpoint: 'sync', group: '1' }
    },
    kafka_group1: {
      executor: 'constant-vus',
      vus: 100,
      duration: '30s',
      exec: 'kafkaTest',
      startTime: '30s',
      tags: { endpoint: 'kafka', group: '1' }
    },
    rabbitmq_group1: {
      executor: 'constant-vus',
      vus: 100,
      duration: '30s',
      exec: 'rabbitmqTest',
      startTime: '1m',
      tags: { endpoint: 'rabbitmq', group: '1' }
    },

    // Group 2: 4 thread groups (200 VUs)
    sync_group2: {
      executor: 'constant-vus',
      vus: 200,
      duration: '30s',
      exec: 'syncTest',
      startTime: '1m30s',
      tags: { endpoint: 'sync', group: '2' }
    },
    kafka_group2: {
      executor: 'constant-vus',
      vus: 200,
      duration: '30s',
      exec: 'kafkaTest',
      startTime: '2m',
      tags: { endpoint: 'kafka', group: '2' }
    },
    rabbitmq_group2: {
      executor: 'constant-vus',
      vus: 200,
      duration: '30s',
      exec: 'rabbitmqTest',
      startTime: '2m30s',
      tags: { endpoint: 'rabbitmq', group: '2' }
    },

    // Group 3: 6 thread groups (300 VUs)
    sync_group3: {
      executor: 'constant-vus',
      vus: 300,
      duration: '30s',
      exec: 'syncTest',
      startTime: '3m',
      tags: { endpoint: 'sync', group: '3' }
    },
    kafka_group3: {
      executor: 'constant-vus',
      vus: 300,
      duration: '30s',
      exec: 'kafkaTest',
      startTime: '3m30s',
      tags: { endpoint: 'kafka', group: '3' }
    },
    rabbitmq_group3: {
      executor: 'constant-vus',
      vus: 300,
      duration: '30s',
      exec: 'rabbitmqTest',
      startTime: '4m',
      tags: { endpoint: 'rabbitmq', group: '3' }
    },
  },
  thresholds: {
    'sync_duration': ['p(95)<500'],
    'kafka_duration': ['p(95)<200'],
    'rabbitmq_duration': ['p(95)<200'],
    'sync_errors': ['rate<0.05'],
    'kafka_errors': ['rate<0.05'],
    'rabbitmq_errors': ['rate<0.05'],
  }
};

// Initialize test summary record
export function setup() {
  // Create summary structure for each endpoint and group
  const groups = [1, 2, 3];
  const endpoints = ['sync', 'kafka', 'rabbitmq'];
  
  let summary = {};
  
  endpoints.forEach(endpoint => {
    summary[endpoint] = {};
    groups.forEach(group => {
      summary[endpoint][group] = {
        totalRequests: 0,
        successCount: 0,
        failureCount: 0,
        wallTimeStart: null,
        wallTimeEnd: null,
        latencies: []
      };
    });
  });
  
  return summary;
}

// Function to generate random payload
function generatePayload() {
  const quantity = Math.floor(Math.random() * 5) + 1;
  const price = 49.99;
  return {
    user_id: Math.floor(Math.random() * 1000),
    status: "pending",
    total_amount: quantity * price,
    items: [
      {
        product_id: 1,
        quantity: quantity,
        price: price
      }
    ]
  };
}

// Track request in the summary data
function trackRequest(endpoint, group, success, latency) {
  // Get or initialize group entry
  if (!testSummary[endpoint]) {
    testSummary[endpoint] = {};
  }
  
  if (!testSummary[endpoint][group]) {
    testSummary[endpoint][group] = {
      totalRequests: 0,
      successCount: 0,
      failureCount: 0,
      wallTimeStart: Date.now(),
      wallTimeEnd: Date.now(),
      latencies: []
    };
  }
  
  let stats = testSummary[endpoint][group];
  
  // Update metrics
  stats.totalRequests += 1;
  if (success) {
    stats.successCount += 1;
    stats.latencies.push(latency);
  } else {
    stats.failureCount += 1;
  }
  
  // Update end time
  stats.wallTimeEnd = Date.now();
}

// Sync endpoint test
export function syncTest() {
  const payload = generatePayload();
  const startTime = Date.now();
  
  // Get group from execution context tags
  const group = __ENV.SCENARIO ? __ENV.SCENARIO.match(/sync_group(\d+)/)[1] : '1';
  
  const res = http.post(endpoints.sync, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' }
  });

  const latency = Date.now() - startTime;
  syncDuration.add(latency);
  
  const ok = check(res, { 
    'status is 200 or 201 or 202': (r) => r.status === 200 || r.status === 201 || r.status === 202
  });
  
  syncErrors.add(!ok);
  
  if (ok) {
    syncThroughput.add(1);
  }
  
  // Track in summary
  trackRequest('sync', group, ok, latency);

  // Add small sleep to avoid overwhelming the system
  sleep(0.1);
}

// Kafka endpoint test
export function kafkaTest() {
  const payload = generatePayload();
  const startTime = Date.now();
  
  // Get group from execution context tags
  const group = __ENV.SCENARIO ? __ENV.SCENARIO.match(/kafka_group(\d+)/)[1] : '1';
  
  const res = http.post(endpoints.kafka, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' }
  });

  const latency = Date.now() - startTime;
  kafkaDuration.add(latency);
  
  const ok = check(res, { 
    'status is 200 or 201 or 202': (r) => r.status === 200 || r.status === 201 || r.status === 202 
  });
  
  kafkaErrors.add(!ok);
  
  if (ok) {
    kafkaThroughput.add(1);
  }
  
  // Track in summary
  trackRequest('kafka', group, ok, latency);

  // Add small sleep to avoid overwhelming the system
  sleep(0.1);
}

// RabbitMQ endpoint test
export function rabbitmqTest() {
  const payload = generatePayload();
  const startTime = Date.now();
  
  // Get group from execution context tags
  const group = __ENV.SCENARIO ? __ENV.SCENARIO.match(/rabbitmq_group(\d+)/)[1] : '1';

  const res = http.post(endpoints.rabbitmq, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' }
  });

  const latency = Date.now() - startTime;
  rabbitmqDuration.add(latency);
  
  const ok = check(res, { 
    'status is 200 or 201 or 202': (r) => r.status === 200 || r.status === 201 || r.status === 202
  });
  
  rabbitmqErrors.add(!ok);
  
  if (ok) {
    rabbitmqThroughput.add(1);
  }
  
  // Track in summary
  trackRequest('rabbitmq', group, ok, latency);

  // Add small sleep to avoid overwhelming the system
  sleep(0.1);
}

// Generate summary data in the format requested
function generateSummaryData() {
  const summaryData = [];
  
  for (const [endpoint, groups] of Object.entries(testSummary)) {
    for (const [groupNum, stats] of Object.entries(groups)) {
      // Skip empty groups
      if (stats.totalRequests === 0) continue;
      
      // Calculate metrics
      const wallTime = (stats.wallTimeEnd - stats.wallTimeStart) / 1000;
      const avgLatency = stats.latencies.length > 0 
        ? stats.latencies.reduce((a, b) => a + b, 0) / stats.latencies.length 
        : 0;
      const throughput = stats.successCount / wallTime;
      
      summaryData.push({
        endpoint: endpoint,
        groups: parseInt(groupNum),
        totalRequests: stats.totalRequests,
        successCount: stats.successCount,
        failureCount: stats.failureCount,
        successRate: `${((stats.successCount / stats.totalRequests) * 100).toFixed(2)}%`,
        wallTime: wallTime.toFixed(2),
        avgLatency: Number(avgLatency.toFixed(2)),
        throughput: Number(throughput.toFixed(2))
      });
    }
  }
  
  return summaryData;
}

// Generate HTML and text summary
export function handleSummary(data) {
  // Generate custom summary data in the requested format
  const summaryData = generateSummaryData();

  console.log("\n📊 Custom Summary");
  console.log(summaryData.map(s =>
    `➡️  ${s.endpoint} group ${s.groups}: ` +
    `${s.totalRequests} reqs, ` +
    `${s.successRate} success, ` +
    `avg ${s.avgLatency}ms, ` +
    `throughput ${s.throughput}/s`
  ).join('\n'));

  
  return {
    "results/messagingtest-summary.html": htmlReport(data),
    "results/messagingtest-summary.txt": textSummary(data, { indent: " ", enableColors: true }),
    "results/messagingtest-summary.json": JSON.stringify(data),
    "results/custom-summary.json": JSON.stringify(summaryData, null, 2)
  };
}