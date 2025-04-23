const axios = require('axios');
const fs = require('fs');
const express = require('express');
const { exec } = require('child_process');
const { performance } = require('perf_hooks');

// Configuration
const baseUrl = process.env.BASE_URL || 'http://localhost:8081';

fs.mkdirSync('public', { recursive: true });

const endpoints = {
  sync: `${baseUrl}/orders/sync`,
  kafka: `${baseUrl}/orders/async/kafka`,
  rabbitmq: `${baseUrl}/orders/async/rabbitmq`
};

const threadGroups = [2, 4, 6];
const threadsPerGroup = 50;
const requestsPerThread = 10;

const kafkaStats = [];
const rabbitStats = [];
const summaryData = []; // 新增：存储汇总数据

const interval = 1000;
const queueName = 'orders';
const kafkaContainer = 'highperformance-ecommerce-microservices-backend-kafka-1';

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
        const match = stdout.match(/orders\s+\d+\s+(\d+)\s+(\d+)/); // current, end
        if (match) {
          const current = parseInt(match[1]);
          const end = parseInt(match[2]);
          return resolve(Math.max(0, end - current));
        }
        return resolve(0);
      }
    );
  });
}

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

async function runTest(endpointKey, url, groupCount) {
  const results = [];
  const totalThreads = groupCount * threadsPerGroup;
  const promises = [];
  const startTime = performance.now();

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

  await Promise.all(promises);
  const wallTime = (performance.now() - startTime) / 1000;

  const successes = results.filter(r => r.success);
  const failures = results.filter(r => !r.success);
  const avgLatency = successes.reduce((a, b) => a + b.latency, 0) / successes.length || 0;
  const throughput = successes.length / wallTime;

  // 新增：记录汇总数据
  summaryData.push({
    endpoint: endpointKey,
    groups: groupCount,
    totalRequests: results.length,
    successCount: successes.length,
    failureCount: failures.length,
    successRate: `${((successes.length / results.length) * 100).toFixed(2)}%`,
    wallTime: wallTime.toFixed(2),
    avgLatency: Number(avgLatency.toFixed(2)),
    throughput: Number(throughput.toFixed(2))
  });

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

(async () => {
  console.log(`📡 Watching Kafka + RabbitMQ queue size...`);
  const rabbitSamples = [];
  const kafkaSamples = [];

  let prevKafkaSize = 0;
  let prevRabbitSize = 0;
  const startTime = Date.now();

  const queueWatcher = setInterval(async () => {
    const time = ((Date.now() - startTime) / 1000).toFixed(1);
    const rabbit = await getRabbitQueueSize();
    const kafka = await getKafkaQueueSize();

    const kafkaInRate = kafka - prevKafkaSize;
    const kafkaOutRate = prevKafkaSize - kafka;
    const rabbitInRate = rabbit - prevRabbitSize;
    const rabbitOutRate = prevRabbitSize - rabbit;

    kafkaStats.push({ time: Number(time), in: Math.max(kafkaInRate, 0), out: Math.max(kafkaOutRate, 0) });
    rabbitStats.push({ time: Number(time), in: Math.max(rabbitInRate, 0), out: Math.max(rabbitOutRate, 0) });

    kafkaSamples.push({ time: Number(time), size: kafka });
    rabbitSamples.push({ time: Number(time), size: rabbit });

    prevKafkaSize = kafka;
    prevRabbitSize = rabbit;

    console.log(`[+${time}s] RabbitMQ: ${rabbit}, Kafka: ${kafka}`);
  }, interval);

  // 🧪 Start performance tests
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

  // 🧼 Wait for queues to drain or timeout fallback
  let zeroCount = 0;
  const maxWait = 60000;
  const drainCheck = setInterval(async () => {
    const rabbit = await getRabbitQueueSize();
    const kafka = await getKafkaQueueSize();
    if ((rabbit === 0 || rabbit === null) && (kafka === 0 || kafka === null)) {
      zeroCount++;
    } else {
      zeroCount = 0;
    }

    if (zeroCount >= 3) {
      finishWatching();
    }
  }, interval);

  setTimeout(() => {
    console.log("⏱️ Timeout reached, forcing queue watching to stop");
    finishWatching();
  }, maxWait);

  function finishWatching() {
    clearInterval(queueWatcher);
    clearInterval(drainCheck);
    fs.writeFileSync('public/queues-rabbitmq.json', JSON.stringify(rabbitSamples, null, 2));
    fs.writeFileSync('public/queues-kafka.json', JSON.stringify(kafkaSamples, null, 2));
    fs.writeFileSync('public/kafka-stats.json', JSON.stringify(kafkaStats, null, 2));
    fs.writeFileSync('public/rabbit-stats.json', JSON.stringify(rabbitStats, null, 2));
    console.log('✅ Queue watching complete');
  }

  // 📊 Start dashboard
  const app = express();
  app.use(express.static('public'));
  app.get('/results.json', (_, res) => res.sendFile(__dirname + '/public/results.json'));
  app.get('/queues-rabbitmq.json', (_, res) => res.sendFile(__dirname + '/public/queues-rabbitmq.json'));
  app.get('/queues-kafka.json', (_, res) => res.sendFile(__dirname + '/public/queues-kafka.json'));
  app.get('/kafka-stats.json', (_, res) => res.sendFile(__dirname + '/public/kafka-stats.json'));
  app.get('/rabbit-stats.json', (_, res) => res.sendFile(__dirname + '/public/rabbit-stats.json'));

  app.listen(3000, () => {
    console.log('📊 View results at: http://localhost:3000');
  });
})();
