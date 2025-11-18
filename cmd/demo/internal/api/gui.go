package api

const guiTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Config Util Demo - Secure Configuration & Testing</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px 0;
            margin-bottom: 30px;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; }
        h2 { color: #667eea; margin: 20px 0 10px; }
        h3 { color: #555; margin: 15px 0 10px; }
        .section {
            background: white;
            padding: 25px;
            margin-bottom: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .card h3 { margin-top: 0; color: #667eea; }
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 8px;
        }
        .status-ok { background: #28a745; }
        .status-warning { background: #ffc107; }
        .status-error { background: #dc3545; }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #555;
        }
        input[type="text"], input[type="email"], input[type="number"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        button {
            background: #667eea;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: background 0.3s;
        }
        button:hover { background: #5568d3; }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .btn-danger {
            background: #dc3545;
        }
        .btn-danger:hover {
            background: #c82333;
        }
        .btn-secondary {
            background: #6c757d;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .results {
            margin-top: 20px;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 4px;
            max-height: 400px;
            overflow-y: auto;
        }
        .results pre {
            background: #2d2d2d;
            color: #f8f8f2;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background: #667eea;
            color: white;
            font-weight: 600;
        }
        tr:hover {
            background: #f8f9fa;
        }
        .alert {
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 4px;
        }
        .alert-success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .alert-error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .alert-info {
            background: #d1ecf1;
            color: #0c5460;
            border: 1px solid #bee5eb;
        }
        .config-item {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }
        .config-item:last-child {
            border-bottom: none;
        }
        .config-label {
            font-weight: 600;
            color: #555;
        }
        .config-value {
            color: #667eea;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔒 Config Util Demo</h1>
            <p>Secure Configuration Management & API Testing Harness</p>
        </header>

        <!-- Configuration Status -->
        <div class="section">
            <h2>📋 Current Configuration</h2>
            <div class="grid">
                <div class="card">
                    <h3>Database</h3>
                    <div class="config-item">
                        <span class="config-label">Status:</span>
                        <span><span class="status-indicator status-ok"></span>Connected</span>
                    </div>
                    <div class="config-item">
                        <span class="config-label">Type:</span>
                        <span class="config-value">SQLite</span>
                    </div>
                </div>
                <div class="card">
                    <h3>Server</h3>
                    <div class="config-item">
                        <span class="config-label">Port:</span>
                        <span class="config-value">{{.Port}}</span>
                    </div>
                    <div class="config-item">
                        <span class="config-label">Debug:</span>
                        <span class="config-value">{{if .Debug}}Enabled{{else}}Disabled{{end}}</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- User Management -->
        <div class="section">
            <h2>👥 User Management</h2>
            <p>Test secure CRUD operations with parameterized queries (SQL injection proof)</p>

            <div class="grid">
                <div class="card">
                    <h3>Create User</h3>
                    <div class="form-group">
                        <label>Username (3-32 alphanumeric)</label>
                        <input type="text" id="username" placeholder="john_doe">
                    </div>
                    <div class="form-group">
                        <label>Email</label>
                        <input type="email" id="email" placeholder="john@example.com">
                    </div>
                    <button onclick="createUser()">Create User</button>
                    <button class="btn-secondary" onclick="testInjection()">Test SQL Injection (Safe)</button>
                </div>

                <div class="card">
                    <h3>Search Users</h3>
                    <div class="form-group">
                        <label>Search Query</label>
                        <input type="text" id="searchQuery" placeholder="Search username or email">
                    </div>
                    <button onclick="searchUsers()">Search</button>
                    <button class="btn-secondary" onclick="listUsers()">List All Users</button>
                </div>
            </div>

            <div id="userResults" class="results" style="display:none;"></div>
        </div>

        <!-- API Testing -->
        <div class="section">
            <h2>🧪 API Testing</h2>
            <p>Test API endpoints and view responses</p>

            <div class="grid">
                <div class="card">
                    <h3>Health Check</h3>
                    <button onclick="testHealth()">Test Health Endpoint</button>
                </div>

                <div class="card">
                    <h3>Configuration Test</h3>
                    <button onclick="testConfig()">Test Config Endpoint</button>
                </div>

                <div class="card">
                    <h3>API Statistics</h3>
                    <button onclick="getStats()">View API Call Stats</button>
                </div>
            </div>

            <div id="apiResults" class="results" style="display:none;"></div>
        </div>

        <!-- Security Testing -->
        <div class="section">
            <h2>🛡️ Security Testing</h2>
            <p>Verify security measures are working correctly</p>

            <div class="alert alert-info">
                <strong>Note:</strong> All injection attempts below are safely handled by parameterized queries and input validation.
            </div>

            <div class="grid">
                <div class="card">
                    <h3>SQL Injection Tests</h3>
                    <button onclick="testSQLInjection1()">Test: ' OR '1'='1</button>
                    <button onclick="testSQLInjection2()">Test: '; DROP TABLE users;--</button>
                </div>

                <div class="card">
                    <h3>XSS Tests</h3>
                    <button onclick="testXSS()">Test: &lt;script&gt;alert('xss')&lt;/script&gt;</button>
                </div>

                <div class="card">
                    <h3>Input Validation</h3>
                    <button onclick="testInvalidEmail()">Test: Invalid Email</button>
                    <button onclick="testInvalidUsername()">Test: Invalid Username</button>
                </div>
            </div>

            <div id="securityResults" class="results" style="display:none;"></div>
        </div>
    </div>

    <script>
        // API Helper Functions
        async function apiCall(endpoint, options = {}) {
            try {
                const response = await fetch(endpoint, {
                    ...options,
                    headers: {
                        'Content-Type': 'application/json',
                        ...options.headers
                    }
                });
                const data = await response.json();
                return { success: response.ok, status: response.status, data };
            } catch (error) {
                return { success: false, error: error.message };
            }
        }

        function showResult(elementId, title, data, isError = false) {
            const element = document.getElementById(elementId);
            element.style.display = 'block';
            element.innerHTML = '<h3>' + title + '</h3><pre>' +
                JSON.stringify(data, null, 2) + '</pre>';
        }

        // User Management Functions
        async function createUser() {
            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;

            const result = await apiCall('/api/users', {
                method: 'POST',
                body: JSON.stringify({ username, email })
            });

            showResult('userResults', 'Create User Result', result, !result.success);
            if (result.success) listUsers();
        }

        async function listUsers() {
            const result = await apiCall('/api/users?limit=20');
            showResult('userResults', 'Users List', result.data);
        }

        async function searchUsers() {
            const query = document.getElementById('searchQuery').value;
            const result = await apiCall('/api/users/search?q=' + encodeURIComponent(query));
            showResult('userResults', 'Search Results', result.data);
        }

        // API Testing Functions
        async function testHealth() {
            const result = await apiCall('/api/health');
            showResult('apiResults', 'Health Check', result.data);
        }

        async function testConfig() {
            const result = await apiCall('/api/config/test');
            showResult('apiResults', 'Configuration Test', result.data);
        }

        async function getStats() {
            const result = await apiCall('/api/stats?limit=20');
            showResult('apiResults', 'API Statistics', result.data);
        }

        // Security Testing Functions
        async function testInjection() {
            const result = await apiCall('/api/users', {
                method: 'POST',
                body: JSON.stringify({
                    username: "admin' OR '1'='1",
                    email: "test@example.com"
                })
            });
            showResult('securityResults', 'SQL Injection Test (should fail validation)', result);
        }

        async function testSQLInjection1() {
            const result = await apiCall("/api/users/search?q=" + encodeURIComponent("' OR '1'='1"));
            showResult('securityResults', "SQL Injection Test: ' OR '1'='1", result.data);
        }

        async function testSQLInjection2() {
            const result = await apiCall("/api/users/search?q=" + encodeURIComponent("'; DROP TABLE users;--"));
            showResult('securityResults', "SQL Injection Test: '; DROP TABLE users;--", result.data);
        }

        async function testXSS() {
            const result = await apiCall('/api/users', {
                method: 'POST',
                body: JSON.stringify({
                    username: "<script>alert('xss')</script>",
                    email: "xss@example.com"
                })
            });
            showResult('securityResults', 'XSS Test (should fail validation)', result);
        }

        async function testInvalidEmail() {
            const result = await apiCall('/api/users', {
                method: 'POST',
                body: JSON.stringify({
                    username: "testuser",
                    email: "not-an-email"
                })
            });
            showResult('securityResults', 'Invalid Email Test', result);
        }

        async function testInvalidUsername() {
            const result = await apiCall('/api/users', {
                method: 'POST',
                body: JSON.stringify({
                    username: "a",
                    email: "test@example.com"
                })
            });
            showResult('securityResults', 'Invalid Username Test (too short)', result);
        }

        // Load initial data
        window.addEventListener('load', function() {
            listUsers();
        });
    </script>
</body>
</html>
`
