// test.js
const axios = require('axios');
const fs = require('fs');
const express = require('express');
const { exec } = require('child_process');
const { performance } = require('perf_hooks');

const endpoints = {
  sync: 'http://localhost:8081/orders/sync',
  kafka: 'http://localhost:8081/orders/async/kafka',
  rabbitmq: 'http://localhost:8081/orders/async/rabbitmq'
};

const threadGroups = [2, 4, 6];
const threadsPerGroup = 100;
const requestsPerThread = 10;

const interval = 2000;
const maxDuration = 30000;

const queueName = 'orders';
const kafkaContainer = 'highperformance-ecommerce-microservices-backend-kafka-1';

// Watch queue sizes
function getRabbitQueueSize() {
  return axios
    .get(`http://localhost:15672/api/queues/%2F/${queueName}`, {
      auth: { username: 'guest', password: 'guest' }
    })
    .then(res => res.data.messages_ready || 0)
    .catch(() => null);
}

function getKafkaQueueSize() {
  return new Promise((resolve) => {
    exec(
      `docker exec ${kafkaContainer} kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group ecommerce-group`,
      (err, stdout, stderr) => {
        if (err || stderr || !stdout) return resolve(null);
        const match = stdout.match(/orders\s+\d+\s+\d+\s+(\d+)/);
        resolve(match ? parseInt(match[1]) : null);
      }
    );
  });
}

// Send one request
async function sendRequest(url) {
  const payload = {
    user_id: Math.floor(Math.random() * 1000),
    product_id: 1,
    quantity: 1
  };

  const start = performance.now();
  try {
    await axios.post(url, payload, { timeout: 5000 });
    const latency = performance.now() - start;
    return { success: true, latency };
  } catch (err) {
    const latency = performance.now() - start;
    return { success: false, latency };
  }
}

// Run one test group
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
  const avgLatency = successes.reduce((a, b) => a + b.latency, 0) / successes.length || 0;
  const throughput = successes.length / duration;

  return {
    endpoint: endpointKey,
    groups: groupCount,
    totalRequests: results.length,
    success: successes.length,
    failures: results.length - successes.length,
    avgLatency: Number(avgLatency.toFixed(2)),
    throughput: Number(throughput.toFixed(2))
  };
}

// Full workflow
(async () => {
  console.log(`📡 Watching Kafka + RabbitMQ queue size...`);
  const rabbitSamples = [];
  const kafkaSamples = [];

  const queueWatcher = setInterval(async () => {
    const time = ((Date.now() - startTime) / 1000).toFixed(1);
    const rabbit = await getRabbitQueueSize();
    const kafka = await getKafkaQueueSize();

    console.log(`[+${time}s] RabbitMQ: ${rabbit}, Kafka: ${kafka}`);
    rabbitSamples.push({ time: Number(time), size: rabbit });
    kafkaSamples.push({ time: Number(time), size: kafka });
  }, interval);

  const startTime = Date.now();
  setTimeout(() => {
    clearInterval(queueWatcher);
    fs.writeFileSync('public/queues-rabbitmq.json', JSON.stringify(rabbitSamples, null, 2));
    fs.writeFileSync('public/queues-kafka.json', JSON.stringify(kafkaSamples, null, 2));
    console.log('✅ Queue watching complete');
  }, maxDuration);

  // Performance test
  const allResults = [];
  for (const [key, url] of Object.entries(endpoints)) {
    for (const group of threadGroups) {
      console.log(`🚀 Testing ${key} with ${group} thread groups...`);
      const result = await runTest(key, url, group);
      allResults.push(result);
    }
  }

  fs.writeFileSync('public/results.json', JSON.stringify(allResults, null, 2));
  console.log('✅ Performance test complete');

  // Serve UI
  const app = express();
  app.use(express.static('public'));
  app.get('/results.json', (_, res) => res.sendFile(__dirname + '/public/results.json'));
  app.get('/queues-rabbitmq.json', (_, res) => res.sendFile(__dirname + '/public/queues-rabbitmq.json'));
  app.get('/queues-kafka.json', (_, res) => res.sendFile(__dirname + '/public/queues-kafka.json'));

  app.listen(3000, () => {
    console.log('📊 View results at: http://localhost:3000');
  });
})();
