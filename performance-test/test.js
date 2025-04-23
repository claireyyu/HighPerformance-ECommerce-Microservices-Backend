// test.js
const axios = require('axios');
const fs = require('fs');
const { performance } = require('perf_hooks');

// 需要修改以匹配端点
const endpoints = {
  sync: 'http://localhost:8081/orders/sync',
  kafka: 'http://localhost:8081/orders/async/kafka',
  rabbitmq: 'http://localhost:8081/orders/async/rabbitmq'
};

const threadGroups = [2, 4, 6];
const threadsPerGroup = 100;
const requestsPerThread = 10;

async function sendRequest(url) {
  const payload = {
    user_id: Math.floor(Math.random() * 1000),
    product_id: 1,
    quantity: 1
  };

  const start = performance.now();
  try {
    const res = await axios.post(url, payload, { timeout: 5000 });
    const latency = performance.now() - start;
    return { success: true, latency };
  } catch (err) {
    const latency = performance.now() - start;
    return { success: false, latency };
  }
}

async function runTest(endpointKey, url, groupCount) {
  const results = [];
  const totalThreads = groupCount * threadsPerGroup;
  const promises = [];

  for (let i = 0; i < totalThreads; i++) {
    promises.push(
      (async () => {
        for (let j = 0; j < requestsPerThread; j++) {
          const result = await sendRequest(url);
          results.push(result);
        }
      })()
    );
  }

  const start = performance.now();
  await Promise.all(promises);
  const duration = (performance.now() - start) / 1000;

  const successes = results.filter(r => r.success);
  const failures = results.length - successes.length;
  const avgLatency = successes.reduce((a, b) => a + b.latency, 0) / successes.length;
  const throughput = successes.length / duration;

  return {
    endpoint: endpointKey,
    groups: groupCount,
    totalRequests: results.length,
    success: successes.length,
    failures,
    avgLatency: Number(avgLatency.toFixed(2)),
    throughput: Number(throughput.toFixed(2))
  };
}

(async () => {
  const allResults = [];

  for (const [key, url] of Object.entries(endpoints)) {
    for (const group of threadGroups) {
      console.log(`Testing ${key} with ${group} thread groups...`);
      const result = await runTest(key, url, group);
      allResults.push(result);
    }
  }

  fs.writeFileSync('results.json', JSON.stringify(allResults, null, 2));
  console.log('✅ Test completed. Results written to results.json');
})();
