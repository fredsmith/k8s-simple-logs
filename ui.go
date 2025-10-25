package main

func getHTMLUI() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kubernetes Logs Viewer</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .log-line {
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
            line-height: 1.5;
        }
        #logs-container {
            scroll-behavior: smooth;
        }
        .container-item:hover {
            background-color: #f3f4f6;
        }
        .container-item.active {
            background-color: #3b82f6;
            color: white;
        }
        .spinner {
            display: inline-block;
            width: 14px;
            height: 14px;
            border: 2px solid rgba(255, 255, 255, 0.3);
            border-radius: 50%;
            border-top-color: white;
            animation: spin 0.8s linear infinite;
        }
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
    </style>
</head>
<body class="bg-gray-100">
    <div class="flex h-screen">
        <!-- Sidebar -->
        <div class="w-80 bg-white shadow-lg flex flex-col">
            <!-- Header -->
            <div class="p-6 border-b border-gray-200">
                <h1 class="text-2xl font-bold text-gray-800">k8s-simple-logs</h1>
                <p class="text-sm text-gray-600 mt-1" id="namespace-info">Loading...</p>
                <p class="text-xs text-gray-500 mt-1" id="version-info">version: <span id="app-version">...</span></p>
            </div>

            <!-- Search -->
            <div class="p-4 border-b border-gray-200">
                <input
                    type="text"
                    id="search-containers"
                    placeholder="Search containers..."
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>

            <!-- Container List -->
            <div class="flex-1 overflow-y-auto p-4" id="containers-list">
                <div class="text-center text-gray-500 py-8">
                    <svg class="inline-block animate-spin h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <p class="mt-2">Loading containers...</p>
                </div>
            </div>

            <!-- Refresh Button -->
            <div class="p-4 border-t border-gray-200">
                <button
                    id="refresh-btn"
                    class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md transition"
                >
                    Refresh Containers
                </button>
            </div>
        </div>

        <!-- Main Content -->
        <div class="flex-1 flex flex-col">
            <!-- Top Bar -->
            <div class="bg-white shadow-sm p-4 border-b border-gray-200">
                <div class="flex justify-between items-center">
                    <div>
                        <h2 class="text-xl font-semibold text-gray-800" id="selected-container">
                            Select a container from the sidebar
                        </h2>
                        <p class="text-sm text-gray-600 mt-1">
                            <span id="selected-pod">No container selected</span>
                        </p>
                    </div>
                    <div class="flex space-x-2">
                        <button
                            id="auto-scroll-btn"
                            class="px-4 py-2 bg-green-500 hover:bg-green-600 text-white font-medium rounded-md transition"
                        >
                            Auto-scroll: ON
                        </button>
                        <button
                            id="clear-logs-btn"
                            class="px-4 py-2 bg-gray-500 hover:bg-gray-600 text-white font-medium rounded-md transition"
                        >
                            Clear
                        </button>
                    </div>
                </div>
            </div>

            <!-- Logs Display -->
            <div class="flex-1 overflow-y-auto bg-gray-900 text-gray-100 p-4" id="logs-container">
                <div class="text-center text-gray-500 py-20">
                    <svg class="inline-block h-16 w-16 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                    </svg>
                    <p class="mt-4 text-lg">Select a container to view logs</p>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentPod = null;
        let currentContainer = null;
        let ws = null;
        let autoScroll = true;
        let containers = [];
        let reconnectAttempts = 0;
        let maxReconnectAttempts = 5;
        let reconnectDelay = 2000; // Start with 2 seconds
        const API_KEY = new URLSearchParams(window.location.search).get('key') || '';

        // Fetch and display version
        async function loadVersion() {
            try {
                const response = await fetch('/version');
                const data = await response.json();
                document.getElementById('app-version').textContent = data.version;
            } catch (error) {
                console.error('Failed to load version:', error);
            }
        }

        // Fetch containers list
        async function loadContainers() {
            try {
                const url = API_KEY ? '/api/containers?key=' + encodeURIComponent(API_KEY) : '/api/containers';
                const response = await fetch(url);
                const data = await response.json();

                if (data.error) {
                    showError('Authentication required. Add ?key=YOUR_KEY to the URL.');
                    return;
                }

                containers = data.containers;
                document.getElementById('namespace-info').textContent = 'Namespace: ' + data.namespace;
                renderContainers(containers);
            } catch (error) {
                console.error('Failed to load containers:', error);
                showError('Failed to load containers: ' + error.message);
            }
        }

        // Render containers in sidebar
        function renderContainers(containersToRender) {
            const listElement = document.getElementById('containers-list');

            if (containersToRender.length === 0) {
                listElement.innerHTML = '<div class="text-center text-gray-500 py-8"><p>No containers found</p></div>';
                return;
            }

            listElement.innerHTML = containersToRender.map(c => {
                const isActive = currentPod === c.podName && currentContainer === c.containerName;
                return ` + "`" + `
                <div class="container-item p-3 mb-2 rounded-md cursor-pointer border border-gray-200 ${isActive ? 'active' : ''}"
                     data-pod="${c.podName}"
                     data-container="${c.containerName}">
                    <div class="flex justify-between items-center">
                        <div class="flex-1">
                            <div class="font-semibold text-sm">${c.containerName}</div>
                            <div class="text-xs ${isActive ? 'text-blue-200' : 'text-gray-500'}">${c.podName}</div>
                        </div>
                        ${isActive ? '<div class="spinner ml-2"></div>' : ''}
                    </div>
                </div>
                ` + "`" + `;
            }).join('');

            // Add click handlers
            document.querySelectorAll('.container-item').forEach(item => {
                item.addEventListener('click', () => {
                    const pod = item.getAttribute('data-pod');
                    const container = item.getAttribute('data-container');
                    selectContainer(pod, container);
                });
            });
        }

        // Select a container and start streaming logs
        function selectContainer(pod, container) {
            currentPod = pod;
            currentContainer = container;

            // Reset reconnection state
            reconnectAttempts = 0;
            reconnectDelay = 2000;

            // Update UI
            document.getElementById('selected-container').textContent = container;
            document.getElementById('selected-pod').textContent = 'Pod: ' + pod;

            // Re-render containers to show spinner on active container
            renderContainers(containers);

            // Clear existing logs
            clearLogs();

            // Close existing WebSocket
            if (ws) {
                ws.close();
            }

            // Connect WebSocket
            connectWebSocket(pod, container);
        }

        // Connect to WebSocket for real-time logs
        function connectWebSocket(pod, container, isReconnect = false) {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws/logs/' + encodeURIComponent(pod) + '/' + encodeURIComponent(container) + (API_KEY ? '?key=' + encodeURIComponent(API_KEY) : '');

            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                // Connection established - reset reconnect counter
                reconnectAttempts = 0;
                reconnectDelay = 2000;
                if (isReconnect) {
                    appendLog('--- Reconnected to log stream ---', 'text-green-400');
                }
                console.log('WebSocket connected for', pod, container);
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                if (data.error) {
                    appendLog('ERROR: ' + data.error, 'text-red-400');
                } else if (data.log) {
                    appendLog(data.log);
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            ws.onclose = (event) => {
                console.log('WebSocket closed for', pod, container, 'Code:', event.code, 'Reason:', event.reason);

                // Only attempt to reconnect if we're still viewing this container
                if (currentPod === pod && currentContainer === container) {
                    if (reconnectAttempts < maxReconnectAttempts) {
                        reconnectAttempts++;
                        const delay = reconnectDelay * reconnectAttempts;
                        appendLog('--- Connection lost. Reconnecting in ' + (delay / 1000) + 's... (attempt ' + reconnectAttempts + '/' + maxReconnectAttempts + ') ---', 'text-yellow-400');

                        setTimeout(() => {
                            if (currentPod === pod && currentContainer === container) {
                                connectWebSocket(pod, container, true);
                            }
                        }, delay);
                    } else {
                        appendLog('--- Connection lost. Maximum reconnection attempts reached. ---', 'text-red-400');
                        appendLog('--- Click the container again to reconnect. ---', 'text-gray-400');
                    }
                }
            };
        }

        // Append a log line
        function appendLog(text, colorClass = 'text-gray-100') {
            const logsContainer = document.getElementById('logs-container');
            const logLine = document.createElement('div');
            logLine.className = 'log-line ' + colorClass;
            logLine.textContent = text;
            logsContainer.appendChild(logLine);

            if (autoScroll) {
                logsContainer.scrollTop = logsContainer.scrollHeight;
            }
        }

        // Clear logs
        function clearLogs() {
            document.getElementById('logs-container').innerHTML = '';
        }

        // Show error message
        function showError(message) {
            const listElement = document.getElementById('containers-list');
            listElement.innerHTML = ` + "`" + `
                <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <p class="font-bold">Error</p>
                    <p class="text-sm">${message}</p>
                </div>
            ` + "`" + `;
        }

        // Search containers
        document.getElementById('search-containers').addEventListener('input', (e) => {
            const searchTerm = e.target.value.toLowerCase();
            const filtered = containers.filter(c =>
                c.containerName.toLowerCase().includes(searchTerm) ||
                c.podName.toLowerCase().includes(searchTerm)
            );
            renderContainers(filtered);
        });

        // Toggle auto-scroll
        document.getElementById('auto-scroll-btn').addEventListener('click', () => {
            autoScroll = !autoScroll;
            const btn = document.getElementById('auto-scroll-btn');
            btn.textContent = 'Auto-scroll: ' + (autoScroll ? 'ON' : 'OFF');
            btn.className = autoScroll
                ? 'px-4 py-2 bg-green-500 hover:bg-green-600 text-white font-medium rounded-md transition'
                : 'px-4 py-2 bg-gray-500 hover:bg-gray-600 text-white font-medium rounded-md transition';
        });

        // Clear logs button
        document.getElementById('clear-logs-btn').addEventListener('click', clearLogs);

        // Refresh containers button
        document.getElementById('refresh-btn').addEventListener('click', loadContainers);

        // Initialize
        loadVersion();
        loadContainers();

        // Cleanup on page unload
        window.addEventListener('beforeunload', () => {
            if (ws) {
                ws.close();
            }
        });
    </script>
</body>
</html>
`
}
