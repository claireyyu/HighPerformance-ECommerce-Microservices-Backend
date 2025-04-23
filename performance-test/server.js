// server.js
const express = require('express');
const app = express();
const PORT = 3000;

app.use(express.static('public'));
app.get('/results.json', (req, res) => res.sendFile(__dirname + '/results.json'));

app.listen(PORT, () => {
  console.log(`✅ Results UI: http://localhost:${PORT}`);
});
