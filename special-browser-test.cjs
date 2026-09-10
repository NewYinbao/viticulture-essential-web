// Isolated seeded three-card and winter-start browser/HTTP regression.
process.env.QA_SPECIAL='1';process.env.QA_REPORT||='result-special.json';require('./visitor-browser-test.cjs');
