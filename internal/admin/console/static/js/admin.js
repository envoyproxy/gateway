// Envoy Gateway Admin Console JavaScript

// Global app object
window.EnvoyGatewayAdmin = {
    // Configuration
    config: {
        refreshInterval: 30000, // 30 seconds
        apiTimeout: 10000 // 10 seconds
    },
    
    // State
    state: {
        autoRefresh: false,
        currentPage: null
    },
    
    // Initialize the application
    init: function() {
        this.setupNavigation();
        this.setupAutoRefresh();
        this.detectCurrentPage();
        this.loadPageData();
    },
    
    // Setup navigation highlighting
    setupNavigation: function() {
        const currentPath = window.location.pathname;
        const navLinks = document.querySelectorAll('.nav a');
        
        navLinks.forEach(link => {
            if (link.getAttribute('href') === currentPath) {
                link.classList.add('active');
            }
        });
    },
    
    // Setup auto-refresh functionality
    setupAutoRefresh: function() {
        const refreshToggle = document.getElementById('auto-refresh');
        if (refreshToggle) {
            refreshToggle.addEventListener('change', (e) => {
                this.state.autoRefresh = e.target.checked;
                if (this.state.autoRefresh) {
                    this.startAutoRefresh();
                } else {
                    this.stopAutoRefresh();
                }
            });
        }
    },
    
    // Detect current page
    detectCurrentPage: function() {
        const path = window.location.pathname;
        if (path === '/' || path === '/index') {
            this.state.currentPage = 'index';
        } else if (path === '/server_info') {
            this.state.currentPage = 'server_info';
        } else if (path === '/config_dump') {
            this.state.currentPage = 'config_dump';
        } else if (path === '/stats') {
            this.state.currentPage = 'stats';
        } else if (path === '/pprof') {
            this.state.currentPage = 'pprof';
        }
    },
    
    // Load page-specific data
    loadPageData: function() {
        switch (this.state.currentPage) {
            case 'index':
                this.loadSystemInfo();
                break;
            case 'server_info':
                this.loadServerInfo();
                break;
            case 'config_dump':
                this.loadConfigDump();
                break;
        }
    },
    
    // API call helper
    apiCall: function(endpoint, callback) {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.config.apiTimeout);
        
        fetch(endpoint, {
            signal: controller.signal,
            headers: {
                'Accept': 'application/json'
            }
        })
        .then(response => {
            clearTimeout(timeoutId);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => callback(null, data))
        .catch(error => {
            clearTimeout(timeoutId);
            callback(error, null);
        });
    },
    
    // Load system information
    loadSystemInfo: function() {
        this.showLoading('system-info');
        this.apiCall('/api/info', (error, data) => {
            this.hideLoading('system-info');
            if (error) {
                this.showError('system-info', 'Failed to load system information: ' + error.message);
                return;
            }
            this.updateSystemInfo(data);
        });
    },
    
    // Update system information display
    updateSystemInfo: function(data) {
        const container = document.getElementById('system-info');
        if (!container) return;

        container.innerHTML = `
            <div class="info-box">
                <div>
                    <strong>Version:</strong> ${data.version}<br>
                    <strong>Uptime:</strong> ${data.uptime}<br>
                    <strong>Go Version:</strong> ${data.goVersion}<br>
                    <strong>Platform:</strong> ${data.platform}
                </div>
                <div>
                    <small>Last updated: ${new Date(data.timestamp).toLocaleString()}</small>
                </div>
            </div>
        `;
    },



    // Update EnvoyGateway configuration display
    updateEnvoyGatewayConfig: function(data) {
        const container = document.getElementById('envoy-gateway-config');
        if (!container) return;

        container.innerHTML = `
            <div class="info-box">
                <div>
                    <strong>API Version:</strong> ${data.apiVersion || 'N/A'}<br>
                    <strong>Kind:</strong> ${data.kind || 'N/A'}<br>
                    <strong>Controller Name:</strong> ${data.gateway?.controllerName || 'N/A'}<br>
                    <strong>Provider Type:</strong> ${data.provider?.type || 'N/A'}
                </div>
                <div>
                    <button class="btn btn-secondary" onclick="EnvoyGatewayAdmin.toggleConfigDetails()" id="config-toggle-btn">
                        Show Details
                    </button>
                    <button class="btn btn-secondary" onclick="EnvoyGatewayAdmin.copyConfigToClipboard(event)" style="margin-left: 0.5rem;">
                        Copy JSON
                    </button>
                </div>
            </div>
            <div id="config-details" class="config-details collapsed">
                <div class="json-code" id="config-json">${this.formatJSON(data)}</div>
            </div>
        `;
    },

    // Toggle configuration details
    toggleConfigDetails: function() {
        const details = document.getElementById('config-details');
        const button = document.getElementById('config-toggle-btn');

        if (details && button) {
            if (details.classList.contains('collapsed')) {
                details.classList.remove('collapsed');
                details.classList.add('expanded');
                button.textContent = 'Hide Details';
            } else {
                details.classList.remove('expanded');
                details.classList.add('collapsed');
                button.textContent = 'Show Details';
            }
        }
    },

    // Copy configuration to clipboard
    copyConfigToClipboard: function(event) {
        const configElement = document.getElementById('config-json');
        if (configElement && event && event.target) {
            const text = configElement.textContent;
            const button = event.target;

            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(text).then(() => {
                    this.showCopySuccess(button);
                }).catch(err => {
                    console.error('Failed to copy to clipboard:', err);
                    this.fallbackCopyToClipboard(text, button);
                });
            } else {
                this.fallbackCopyToClipboard(text, button);
            }
        }
    },

    // Show copy success feedback
    showCopySuccess: function(button) {
        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.classList.add('btn-success');
        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('btn-success');
        }, 2000);
    },

    // Fallback copy method for older browsers
    fallbackCopyToClipboard: function(text, button) {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful && button) {
                this.showCopySuccess(button);
            }
        } catch (err) {
            console.error('Fallback copy failed:', err);
        }

        document.body.removeChild(textArea);
    },

    // Format JSON with syntax highlighting
    formatJSON: function(obj) {
        const jsonString = JSON.stringify(obj, null, 2);
        return this.syntaxHighlight(jsonString);
    },

    // Add syntax highlighting to JSON
    syntaxHighlight: function(json) {
        json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
            let cls = 'json-number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'json-key';
                } else {
                    cls = 'json-string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'json-boolean';
            } else if (/null/.test(match)) {
                cls = 'json-null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    },
    
    // Load server information
    loadServerInfo: function() {
        this.showLoading('server-info');
        this.apiCall('/api/server_info', (error, data) => {
            this.hideLoading('server-info');
            if (error) {
                this.showError('server-info', 'Failed to load server information: ' + error.message);
                return;
            }
            this.updateServerInfo(data);
        });
    },
    
    // Update server information display
    updateServerInfo: function(data) {
        const container = document.getElementById('server-info');
        if (!container) return;

        let componentsHtml = '';
        data.components.forEach(component => {
            const statusClass = component.status.toLowerCase() === 'running' ? 'running' : 'error';
            componentsHtml += `
                <tr>
                    <td>${component.name}</td>
                    <td><span class="status ${statusClass}">${component.status}</span></td>
                    <td>${component.message}</td>
                </tr>
            `;
        });

        container.innerHTML = `
            <div class="info-box">
                <div>
                    <strong>State:</strong> <span class="status running">${data.state}</span><br>
                    <strong>Version:</strong> ${data.version}<br>
                    <strong>Uptime:</strong> ${data.uptime}
                </div>
                <div>
                    <small>Last updated: ${new Date(data.lastUpdated).toLocaleString()}</small>
                </div>
            </div>
            <table class="table">
                <thead>
                    <tr>
                        <th>Component</th>
                        <th>Status</th>
                        <th>Message</th>
                    </tr>
                </thead>
                <tbody>
                    ${componentsHtml}
                </tbody>
            </table>
        `;

        // Update EnvoyGateway configuration if we're on server_info page
        if (data.envoyGatewayConfig) {
            this.updateEnvoyGatewayConfig(data.envoyGatewayConfig);
        }
    },
    
    // Load configuration dump
    loadConfigDump: function() {
        // Load config summary
        this.showLoading('config-summary');
        this.apiCall('/api/config_dump', (error, data) => {
            this.hideLoading('config-summary');
            if (error) {
                this.showError('config-summary', 'Failed to load configuration summary: ' + error.message);
                return;
            }
            this.updateConfigSummary(data);
        });

        // Load resources with new layout
        this.showLoading('config-resources-container');
        this.apiCall('/api/config_dump', (error, data) => {
            this.hideLoading('config-resources-container');
            if (error) {
                this.showError('config-resources-container', 'Failed to load resources: ' + error.message);
                return;
            }
            this.updateResourcesContainer(data);
        });
    },
    
    // Update configuration dump display
    updateConfigDump: function(data) {
        const container = document.getElementById('config-dump');
        if (container) {
            container.innerHTML = `
                <div class="info-box">
                    <div>
                        <strong>Gateways:</strong> ${data.gateways ? data.gateways.length : 0}<br>
                        <strong>HTTP Routes:</strong> ${data.httpRoutes ? data.httpRoutes.length : 0}<br>
                        <strong>Gateway Classes:</strong> ${data.gatewayClass ? data.gatewayClass.length : 0}
                    </div>
                    <div>
                        <small>Last updated: ${new Date(data.lastUpdated).toLocaleString()}</small>
                    </div>
                </div>
                <div class="code">
                    ${JSON.stringify(data, null, 2)}
                </div>
            `;
        }

        // Update individual resource lists
        this.updateGatewaysList(data.gateways || []);
        this.updateHTTPRoutesList(data.httpRoutes || []);
        this.updateGRPCRoutesList(data.grpcRoutes || []);
        this.updateTLSRoutesList(data.tlsRoutes || []);
        this.updateTCPRoutesList(data.tcpRoutes || []);
        this.updateUDPRoutesList(data.udpRoutes || []);
        this.updateGatewayClassesList(data.gatewayClass || []);

        // Update Envoy Gateway Policies
        this.updateResourcesList('clienttrafficpolicies-list', data.clientTrafficPolicies || [], 'Client Traffic Policies');
        this.updateResourcesList('backendtrafficpolicies-list', data.backendTrafficPolicies || [], 'Backend Traffic Policies');
        this.updateResourcesList('backendtlspolicies-list', data.backendTLSPolicies || [], 'Backend TLS Policies');
        this.updateResourcesList('securitypolicies-list', data.securityPolicies || [], 'Security Policies');
        this.updateResourcesList('envoypatchpolicies-list', data.envoyPatchPolicies || [], 'Envoy Patch Policies');
        this.updateResourcesList('envoyextensionpolicies-list', data.envoyExtensionPolicies || [], 'Envoy Extension Policies');

        // Update Kubernetes Resources
        this.updateResourcesList('services-list', data.services || [], 'Services');
        this.updateResourcesList('secrets-list', data.secrets || [], 'Secrets');
        this.updateResourcesList('configmaps-list', data.configMaps || [], 'ConfigMaps');
        this.updateResourcesList('namespaces-list', data.namespaces || [], 'Namespaces');
        this.updateResourcesList('endpointslices-list', data.endpointSlices || [], 'Endpoint Slices');

        // Update Other Resources
        this.updateResourcesList('referencegrants-list', data.referenceGrants || [], 'Reference Grants');
        this.updateResourcesList('httproutefilters-list', data.httpRouteFilters || [], 'HTTP Route Filters');
        this.updateResourcesList('envoyproxies-list', data.envoyProxies || [], 'Envoy Proxies');
        this.updateResourcesList('backends-list', data.backends || [], 'Backends');
        this.updateResourcesList('serviceimports-list', data.serviceImports || [], 'Service Imports');

        // Update config summary
        this.updateConfigSummary(data);
    },

    // Update gateways list
    updateGatewaysList: function(gateways) {
        const container = document.getElementById('gateways-list');
        if (!container) return;

        if (gateways.length === 0) {
            container.innerHTML = '<p class="text-muted">No gateways found</p>';
            return;
        }

        const gatewaysHtml = gateways.map(gw => `
            <div class="resource-item">
                <div class="resource-header">
                    <div class="resource-name">${gw.name}</div>
                    <div class="resource-namespace">${gw.namespace}</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = gatewaysHtml;
    },

    // Update HTTP routes list
    updateHTTPRoutesList: function(httpRoutes) {
        this.updateRoutesList('httproutes-list', httpRoutes, 'HTTP routes');
    },

    // Update GRPC routes list
    updateGRPCRoutesList: function(grpcRoutes) {
        this.updateRoutesList('grpcroutes-list', grpcRoutes, 'GRPC routes');
    },

    // Update TLS routes list
    updateTLSRoutesList: function(tlsRoutes) {
        this.updateRoutesList('tlsroutes-list', tlsRoutes, 'TLS routes');
    },

    // Update TCP routes list
    updateTCPRoutesList: function(tcpRoutes) {
        this.updateRoutesList('tcproutes-list', tcpRoutes, 'TCP routes');
    },

    // Update UDP routes list
    updateUDPRoutesList: function(udpRoutes) {
        this.updateRoutesList('udproutes-list', udpRoutes, 'UDP routes');
    },

    // Generic method to update any routes list
    updateRoutesList: function(containerId, routes, routeType) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (routes.length === 0) {
            container.innerHTML = `<p class="text-muted">No ${routeType} found</p>`;
            return;
        }

        const routesHtml = routes.map(route => `
            <div class="resource-item">
                <div class="resource-header">
                    <div class="resource-name">${route.name}</div>
                    <div class="resource-namespace">${route.namespace}</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = routesHtml;
    },

    // Update gateway classes list
    updateGatewayClassesList: function(gatewayClasses) {
        const container = document.getElementById('gatewayclasses-list');
        if (!container) return;

        if (gatewayClasses.length === 0) {
            container.innerHTML = '<p class="text-muted">No gateway classes found</p>';
            return;
        }

        const classesHtml = gatewayClasses.map(gc => `
            <div class="resource-item">
                <div class="resource-header">
                    <div class="resource-name">${gc.name}</div>
                    <div class="resource-scope">Cluster-scoped</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = classesHtml;
    },

    // Generic method to update any resource list
    updateResourcesList: function(containerId, resources, resourceType) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (resources.length === 0) {
            container.innerHTML = `<p class="text-muted">No ${resourceType.toLowerCase()} found</p>`;
            return;
        }

        const resourcesHtml = resources.map(resource => `
            <div class="resource-item">
                <div class="resource-header">
                    <div class="resource-name">${resource.name}</div>
                    <div class="resource-namespace">${resource.namespace || 'Cluster-scoped'}</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = resourcesHtml;
    },

    // Update config summary
    updateConfigSummary: function(data) {
        const container = document.getElementById('config-summary');
        if (!container) return;

        const summary = `
            <div class="config-summary-grid">
                <!-- Gateway API Core Resources -->
                <div class="summary-item">
                    <h3>${data.gateways ? data.gateways.length : 0}</h3>
                    <p>Gateways</p>
                </div>
                <div class="summary-item">
                    <h3>${data.httpRoutes ? data.httpRoutes.length : 0}</h3>
                    <p>HTTP Routes</p>
                </div>
                <div class="summary-item">
                    <h3>${data.grpcRoutes ? data.grpcRoutes.length : 0}</h3>
                    <p>GRPC Routes</p>
                </div>
                <div class="summary-item">
                    <h3>${data.tlsRoutes ? data.tlsRoutes.length : 0}</h3>
                    <p>TLS Routes</p>
                </div>
                <div class="summary-item">
                    <h3>${data.tcpRoutes ? data.tcpRoutes.length : 0}</h3>
                    <p>TCP Routes</p>
                </div>
                <div class="summary-item">
                    <h3>${data.udpRoutes ? data.udpRoutes.length : 0}</h3>
                    <p>UDP Routes</p>
                </div>
                <div class="summary-item">
                    <h3>${data.gatewayClass ? data.gatewayClass.length : 0}</h3>
                    <p>Gateway Classes</p>
                </div>

                <!-- Envoy Gateway Policies -->
                <div class="summary-item">
                    <h3>${data.clientTrafficPolicies ? data.clientTrafficPolicies.length : 0}</h3>
                    <p>Client Traffic Policies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.backendTrafficPolicies ? data.backendTrafficPolicies.length : 0}</h3>
                    <p>Backend Traffic Policies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.backendTLSPolicies ? data.backendTLSPolicies.length : 0}</h3>
                    <p>Backend TLS Policies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.securityPolicies ? data.securityPolicies.length : 0}</h3>
                    <p>Security Policies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.envoyPatchPolicies ? data.envoyPatchPolicies.length : 0}</h3>
                    <p>Envoy Patch Policies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.envoyExtensionPolicies ? data.envoyExtensionPolicies.length : 0}</h3>
                    <p>Envoy Extension Policies</p>
                </div>

                <!-- Kubernetes Resources -->
                <div class="summary-item">
                    <h3>${data.services ? data.services.length : 0}</h3>
                    <p>Services</p>
                </div>
                <div class="summary-item">
                    <h3>${data.secrets ? data.secrets.length : 0}</h3>
                    <p>Secrets</p>
                </div>
                <div class="summary-item">
                    <h3>${data.configMaps ? data.configMaps.length : 0}</h3>
                    <p>ConfigMaps</p>
                </div>
                <div class="summary-item">
                    <h3>${data.namespaces ? data.namespaces.length : 0}</h3>
                    <p>Namespaces</p>
                </div>
                <div class="summary-item">
                    <h3>${data.endpointSlices ? data.endpointSlices.length : 0}</h3>
                    <p>Endpoint Slices</p>
                </div>

                <!-- Other Resources -->
                <div class="summary-item">
                    <h3>${data.referenceGrants ? data.referenceGrants.length : 0}</h3>
                    <p>Reference Grants</p>
                </div>
                <div class="summary-item">
                    <h3>${data.httpRouteFilters ? data.httpRouteFilters.length : 0}</h3>
                    <p>HTTP Route Filters</p>
                </div>
                <div class="summary-item">
                    <h3>${data.envoyProxies ? data.envoyProxies.length : 0}</h3>
                    <p>Envoy Proxies</p>
                </div>
                <div class="summary-item">
                    <h3>${data.backends ? data.backends.length : 0}</h3>
                    <p>Backends</p>
                </div>
                <div class="summary-item">
                    <h3>${data.serviceImports ? data.serviceImports.length : 0}</h3>
                    <p>Service Imports</p>
                </div>
            </div>
        `;

        container.innerHTML = summary;
    },
    
    // Show loading indicator
    showLoading: function(containerId) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = '<div class="loading"></div> Loading...';
        }
    },
    
    // Hide loading indicator
    hideLoading: function(containerId) {
        // Loading will be replaced by content
    },
    
    // Show error message
    showError: function(containerId, message) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = `<div class="info-box error">${message}</div>`;
        }
    },
    
    // Start auto-refresh
    startAutoRefresh: function() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
        }
        this.refreshTimer = setInterval(() => {
            this.loadPageData();
        }, this.config.refreshInterval);
    },
    
    // Stop auto-refresh
    stopAutoRefresh: function() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
            this.refreshTimer = null;
        }
    },
    
    // Update configuration summary
    updateConfigSummary: function(data) {
        const container = document.getElementById('config-summary');
        if (!container) return;

        // Collect all resource types and their counts
        const resourceStats = [
            { name: 'Gateways', count: data.gateways?.length || 0 },
            { name: 'HTTP Routes', count: data.httpRoutes?.length || 0 },
            { name: 'GRPC Routes', count: data.grpcRoutes?.length || 0 },
            { name: 'TLS Routes', count: data.tlsRoutes?.length || 0 },
            { name: 'TCP Routes', count: data.tcpRoutes?.length || 0 },
            { name: 'UDP Routes', count: data.udpRoutes?.length || 0 },
            { name: 'Gateway Classes', count: data.gatewayClass?.length || 0 },
            { name: 'Client Traffic Policies', count: data.clientTrafficPolicies?.length || 0 },
            { name: 'Backend Traffic Policies', count: data.backendTrafficPolicies?.length || 0 },
            { name: 'Backend TLS Policies', count: data.backendTLSPolicies?.length || 0 },
            { name: 'Security Policies', count: data.securityPolicies?.length || 0 },
            { name: 'Envoy Patch Policies', count: data.envoyPatchPolicies?.length || 0 },
            { name: 'Envoy Extension Policies', count: data.envoyExtensionPolicies?.length || 0 },
            { name: 'Services', count: data.services?.length || 0 },
            { name: 'Secrets', count: data.secrets?.length || 0 },
            { name: 'ConfigMaps', count: data.configMaps?.length || 0 },
            { name: 'Namespaces', count: data.namespaces?.length || 0 },
            { name: 'Endpoint Slices', count: data.endpointSlices?.length || 0 },
            { name: 'Reference Grants', count: data.referenceGrants?.length || 0 },
            { name: 'HTTP Route Filters', count: data.httpRouteFilters?.length || 0 },
            { name: 'Envoy Proxies', count: data.envoyProxies?.length || 0 },
            { name: 'Backends', count: data.backends?.length || 0 },
            { name: 'Service Imports', count: data.serviceImports?.length || 0 }
        ];

        // Filter out resources with 0 count and sort by count (descending)
        const filteredStats = resourceStats
            .filter(stat => stat.count > 0)
            .sort((a, b) => b.count - a.count);

        // Calculate total resources
        const totalResources = resourceStats.reduce((sum, stat) => sum + stat.count, 0);

        // Generate bar chart HTML
        const maxCount = filteredStats.length > 0 ? filteredStats[0].count : 1;
        const chartHTML = this.generateBarChart(filteredStats, maxCount);

        container.innerHTML = `
            <div class="config-summary-chart">
                ${chartHTML}
            </div>
            <div class="config-summary-total">
                <strong>Total Resources: ${totalResources}</strong>
            </div>
            <div style="margin-top: 1rem; text-align: center; color: #6c757d; font-size: 0.875rem;">
                Last updated: ${new Date(data.lastUpdated).toLocaleString()}
            </div>
        `;
    },

    // Generate bar chart HTML
    generateBarChart: function(stats, maxCount) {
        if (stats.length === 0) {
            return '<div class="empty-chart">No resources found</div>';
        }

        // Calculate grid lines (5 horizontal lines for Y-axis)
        const gridLines = [];
        const step = maxCount / 5;
        for (let i = 1; i <= 5; i++) {
            const value = Math.ceil(step * i);
            const percentage = (value / maxCount) * 100;
            gridLines.push({ value, percentage: 100 - percentage });
        }

        let chartHTML = '<div class="bar-chart">';

        // Add grid lines and Y-axis labels
        chartHTML += '<div class="bar-chart-grid">';
        gridLines.forEach(line => {
            chartHTML += `
                <div class="grid-line" style="bottom: ${100 - line.percentage}%"></div>
                <div class="y-axis-label" style="bottom: ${100 - line.percentage}%">${line.value}</div>
            `;
        });
        // Add zero line
        chartHTML += `
            <div class="grid-line" style="bottom: 0%"></div>
            <div class="y-axis-label" style="bottom: 0%">0</div>
        `;
        chartHTML += '</div>';

        // Add bars (X-axis items)
        stats.forEach(stat => {
            const heightPercentage = (stat.count / maxCount) * 100;
            const maxHeight = 250; // Maximum height in pixels
            const actualHeight = (heightPercentage / 100) * maxHeight;

            chartHTML += `
                <div class="bar-item">
                    <div class="bar-container" style="height: ${actualHeight}px;">
                        <div class="bar-value">${stat.count}</div>
                    </div>
                    <div class="bar-label">${stat.name}</div>
                </div>
            `;
        });

        chartHTML += '</div>';
        return chartHTML;
    },

    // Update resources container with new layout
    updateResourcesContainer: function(data) {
        const container = document.getElementById('config-resources-container');
        if (!container) return;

        const resourceCategories = [
            {
                title: 'Gateway API Resources',
                icon: '🚪',
                resources: [
                    { name: 'Gateways', data: data.gateways || [], id: 'gateways' },
                    { name: 'Gateway Classes', data: data.gatewayClass || [], id: 'gatewayclass' },
                    { name: 'HTTP Routes', data: data.httpRoutes || [], id: 'httproutes' },
                    { name: 'GRPC Routes', data: data.grpcRoutes || [], id: 'grpcroutes' },
                    { name: 'TLS Routes', data: data.tlsRoutes || [], id: 'tlsroutes' },
                    { name: 'TCP Routes', data: data.tcpRoutes || [], id: 'tcproutes' },
                    { name: 'UDP Routes', data: data.udpRoutes || [], id: 'udproutes' },
                    { name: 'Reference Grants', data: data.referenceGrants || [], id: 'referencegrants' }
                ]
            },
            {
                title: 'Envoy Gateway Policies',
                icon: '🛡️',
                resources: [
                    { name: 'Client Traffic Policies', data: data.clientTrafficPolicies || [], id: 'clienttrafficpolicies' },
                    { name: 'Backend Traffic Policies', data: data.backendTrafficPolicies || [], id: 'backendtrafficpolicies' },
                    { name: 'Backend TLS Policies', data: data.backendTLSPolicies || [], id: 'backendtlspolicies' },
                    { name: 'Security Policies', data: data.securityPolicies || [], id: 'securitypolicies' },
                    { name: 'Envoy Patch Policies', data: data.envoyPatchPolicies || [], id: 'envoypatchpolicies' },
                    { name: 'Envoy Extension Policies', data: data.envoyExtensionPolicies || [], id: 'envoyextensionpolicies' }
                ]
            },
            {
                title: 'Envoy Gateway Resources',
                icon: '⚙️',
                resources: [
                    { name: 'HTTP Route Filters', data: data.httpRouteFilters || [], id: 'httproutefilters' },
                    { name: 'Envoy Proxies', data: data.envoyProxies || [], id: 'envoyproxies' },
                    { name: 'Backends', data: data.backends || [], id: 'backends' }
                ]
            },
            {
                title: 'Kubernetes Resources',
                icon: '☸️',
                resources: [
                    { name: 'Services', data: data.services || [], id: 'services' },
                    { name: 'Secrets', data: data.secrets || [], id: 'secrets' },
                    { name: 'ConfigMaps', data: data.configMaps || [], id: 'configmaps' },
                    { name: 'Namespaces', data: data.namespaces || [], id: 'namespaces' },
                    { name: 'Endpoint Slices', data: data.endpointSlices || [], id: 'endpointslices' },
                    { name: 'Service Imports', data: data.serviceImports || [], id: 'serviceimports' }
                ]
            }
        ];

        let html = '';
        resourceCategories.forEach(category => {
            const totalCount = category.resources.reduce((sum, resource) => sum + resource.data.length, 0);

            html += `
                <div class="resource-category collapsed" data-category="${category.id}">
                    <div class="resource-category-header">
                        <div class="resource-category-title">
                            <span>${category.icon}</span>
                            <span>${category.title}</span>
                        </div>
                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                            <span class="resource-count-badge">${totalCount}</span>
                            <span class="resource-category-toggle">▼</span>
                        </div>
                    </div>
                    <div class="resource-category-content">
                        ${this.renderResourceSubcategories(category.resources)}
                    </div>
                </div>
            `;
        });

        container.innerHTML = html;
    },

    // Render resource subcategories
    renderResourceSubcategories: function(resources) {
        let html = '';

        resources.forEach(resource => {
            if (resource.data.length > 0) {
                html += `
                    <div class="resource-subcategory" data-resource-type="${resource.id}" style="margin-bottom: 1.5rem;">
                        <h4 class="resource-subcategory-title" style="margin-bottom: 0.75rem; color: #495057; font-size: 1rem; font-weight: 600;">
                            ${resource.name} (<span class="resource-count">${resource.data.length}</span>)
                        </h4>
                        <div class="resource-grid">
                            ${resource.data.map(item => this.renderResourceCard(item)).join('')}
                        </div>
                    </div>
                `;
            } else {
                html += `
                    <div class="resource-subcategory" data-resource-type="${resource.id}" style="margin-bottom: 1.5rem;">
                        <h4 class="resource-subcategory-title" style="margin-bottom: 0.75rem; color: #6c757d; font-size: 1rem; font-weight: 600;">
                            ${resource.name} (<span class="resource-count">0</span>)
                        </h4>
                        <div class="empty-state">No ${resource.name.toLowerCase()} found</div>
                    </div>
                `;
            }
        });

        return html;
    },

    // Render individual resource card
    renderResourceCard: function(resource) {
        const namespaceDisplay = resource.namespace
            ? `<span class="resource-card-namespace">ns: ${resource.namespace}</span>`
            : `<span class="resource-card-cluster-scoped">cluster-scoped</span>`;

        return `
            <div class="resource-card">
                <div class="resource-card-name">${resource.name}</div>
                ${namespaceDisplay}
            </div>
        `;
    },

    // Manual refresh
    refresh: function() {
        this.loadPageData();
    }
};

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    window.EnvoyGatewayAdmin.init();
});

// Cleanup on page unload
window.addEventListener('beforeunload', function() {
    window.EnvoyGatewayAdmin.stopAutoRefresh();
});
