// Enhanced Web UI Utilities and Components
// Modern, accessible, and feature-rich dashboard utilities

class DashboardUtils {
    constructor() {
        this.theme = localStorage.getItem('dashboard-theme') || 'dark';
        this.notifications = [];
        this.shortcuts = new Map();
        this.plugins = new Map();
        this.metrics = {
            pageLoad: Date.now(),
            interactions: 0,
            errors: 0,
            lastActivity: Date.now()
        };
        
        this.init();
    }

    init() {
        this.setupTheme();
        this.setupShortcuts();
        this.setupNotifications();
        this.setupAccessibility();
        this.setupPerformanceMonitoring();
        this.setupErrorHandling();
        this.setupWebSocketEnhancements();
        this.setupLocalStorageSync();
        
        console.log('🚀 Dashboard initialized with enhanced utilities');
    }

    // Theme Management
    setupTheme() {
        document.documentElement.setAttribute('data-theme', this.theme);
        
        // Create theme toggle button
        const themeToggle = this.createElement('button', {
            className: 'theme-toggle',
            innerHTML: this.theme === 'dark' ? '🌙' : '☀️',
            onclick: () => this.toggleTheme(),
            title: 'Toggle theme (Ctrl+Shift+T)'
        });
        
        document.body.appendChild(themeToggle);
        
        // Keyboard shortcut
        this.addShortcut('ctrl+shift+t', () => this.toggleTheme());
    }

    toggleTheme() {
        this.theme = this.theme === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', this.theme);
        localStorage.setItem('dashboard-theme', this.theme);
        
        const toggle = document.querySelector('.theme-toggle');
        if (toggle) {
            toggle.innerHTML = this.theme === 'dark' ? '🌙' : '☀️';
        }
        
        this.showNotification(`Theme changed to ${this.theme}`, 'info');
        this.trackMetric('theme_change');
    }

    // Enhanced Notifications
    setupNotifications() {
        this.createNotificationContainer();
    }

    createNotificationContainer() {
        const container = this.createElement('div', {
            id: 'notifications',
            className: 'notifications-container'
        });
        document.body.appendChild(container);
    }

    showNotification(message, type = 'info', duration = 5000, actions = []) {
        const notification = this.createElement('div', {
            className: `notification notification-${type}`,
            innerHTML: `
                <div class="notification-content">
                    <span class="notification-icon">${this.getNotificationIcon(type)}</span>
                    <span class="notification-message">${message}</span>
                    <button class="notification-close" onclick="dashboardUtils.removeNotification(this)">×</button>
                </div>
                ${actions.length > 0 ? `
                    <div class="notification-actions">
                        ${actions.map(action => `
                            <button class="notification-action" onclick="${action.onclick}">
                                ${action.label}
                            </button>
                        `).join('')}
                    </div>
                ` : ''}
            `
        });

        const container = document.getElementById('notifications');
        container.appendChild(notification);

        // Auto-remove
        if (duration > 0) {
            setTimeout(() => this.removeNotification(notification.querySelector('.notification-close')), duration);
        }

        this.trackMetric('notification', { type, message });
    }

    removeNotification(button) {
        const notification = button.closest('.notification');
        notification.style.animation = 'slideOutRight 0.3s ease forwards';
        setTimeout(() => notification.remove(), 300);
    }

    getNotificationIcon(type) {
        const icons = {
            info: 'ℹ️',
            success: '✅',
            warning: '⚠️',
            error: '❌',
            loading: '⏳'
        };
        return icons[type] || icons.info;
    }

    // Keyboard Shortcuts
    setupShortcuts() {
        document.addEventListener('keydown', (e) => {
            const key = this.getShortcutKey(e);
            if (this.shortcuts.has(key)) {
                e.preventDefault();
                this.shortcuts.get(key)();
            }
        });
    }

    addShortcut(keys, callback, description = '') {
        const key = Array.isArray(keys) ? keys : [keys];
        key.forEach(k => this.shortcuts.set(k.toLowerCase(), callback));
        
        if (description) {
            console.log(`⌨️  Shortcut: ${key.join(' or ')} - ${description}`);
        }
    }

    getShortcutKey(e) {
        const parts = [];
        if (e.ctrlKey) parts.push('ctrl');
        if (e.shiftKey) parts.push('shift');
        if (e.altKey) parts.push('alt');
        if (e.metaKey) parts.push('meta');
        parts.push(e.key.toLowerCase());
        return parts.join('+');
    }

    // Accessibility Features
    setupAccessibility() {
        // Focus management
        this.setupFocusManagement();
        
        // Screen reader support
        this.setupScreenReaderSupport();
        
        // High contrast mode
        this.addShortcut('ctrl+shift+h', () => this.toggleHighContrast());
        
        // Reduce motion
        this.addShortcut('ctrl+shift+m', () => this.toggleReducedMotion());
    }

    setupFocusManagement() {
        // Skip to main content
        const skipLink = this.createElement('a', {
            href: '#main-content',
            className: 'skip-link',
            innerHTML: 'Skip to main content',
            onclick: (e) => {
                e.preventDefault();
                document.getElementById('main-content')?.focus();
            }
        });
        document.body.insertBefore(skipLink, document.body.firstChild);
        
        // Focus traps for modals
        this.setupFocusTraps();
    }

    setupScreenReaderSupport() {
        // Announce dynamic content changes
        const announcer = this.createElement('div', {
            'aria-live': 'polite',
            'aria-atomic': 'true',
            className: 'sr-only',
            id: 'screen-reader-announcer'
        });
        document.body.appendChild(announcer);
        
        this.announceToScreenReader = (message) => {
            announcer.textContent = message;
            setTimeout(() => (announcer.textContent = ''), 1000);
        };
    }

    toggleHighContrast() {
        const isHighContrast = document.documentElement.toggleAttribute('data-high-contrast');
        this.showNotification(`High contrast ${isHighContrast ? 'enabled' : 'disabled'}`, 'info');
        this.announceToScreenReader(`High contrast mode ${isHighContrast ? 'enabled' : 'disabled'}`);
    }

    toggleReducedMotion() {
        const isReduced = document.documentElement.toggleAttribute('data-reduced-motion');
        this.showNotification(`Reduced motion ${isReduced ? 'enabled' : 'disabled'}`, 'info');
        this.announceToScreenReader(`Reduced motion ${isReduced ? 'enabled' : 'disabled'}`);
    }

    // Performance Monitoring
    setupPerformanceMonitoring() {
        // Track page load performance
        if ('performance' in window) {
            window.addEventListener('load', () => {
                const perfData = performance.getEntriesByType('navigation')[0];
                this.metrics.pageLoadTime = perfData.loadEventEnd - perfData.fetchStart;
                console.log(`📊 Page load time: ${this.metrics.pageLoadTime}ms`);
            });
        }

        // Track interaction performance
        this.setupInteractionTracking();
        
        // Monitor memory usage
        this.monitorMemoryUsage();
    }

    setupInteractionTracking() {
        const trackInteraction = (event) => {
            this.metrics.interactions++;
            this.metrics.lastActivity = Date.now();
            
            // Track slow interactions
            const start = performance.now();
            requestAnimationFrame(() => {
                const duration = performance.now() - start;
                if (duration > 100) {
                    console.warn(`🐌 Slow interaction detected: ${duration.toFixed(2)}ms on ${event.target}`);
                }
            });
        };

        document.addEventListener('click', trackInteraction);
        document.addEventListener('keydown', trackInteraction);
    }

    monitorMemoryUsage() {
        if ('memory' in performance) {
            setInterval(() => {
                const memory = performance.memory;
                const used = (memory.usedJSHeapSize / 1048576).toFixed(2);
                const total = (memory.totalJSHeapSize / 1048576).toFixed(2);
                const limit = (memory.jsHeapSizeLimit / 1048576).toFixed(2);
                
                // Update memory display if exists
                const memoryDisplay = document.getElementById('memory-usage');
                if (memoryDisplay) {
                    memoryDisplay.innerHTML = `Memory: ${used}MB / ${total}MB (${limit}MB limit)`;
                }
                
                // Warning if memory usage is high
                if (parseFloat(used) / parseFloat(limit) > 0.8) {
                    this.showNotification('High memory usage detected', 'warning');
                }
            }, 10000);
        }
    }

    // Enhanced Error Handling
    setupErrorHandling() {
        // Global error handler
        window.addEventListener('error', (e) => {
            this.handleError(e.error || new Error(e.message), 'javascript');
        });

        // Unhandled promise rejections
        window.addEventListener('unhandledrejection', (e) => {
            this.handleError(e.reason, 'promise');
        });

        // Network error monitoring
        this.setupNetworkErrorMonitoring();
    }

    handleError(error, type = 'general') {
        this.metrics.errors++;
        console.error(`❌ ${type} error:`, error);
        
        this.showNotification(
            `An error occurred: ${error.message || 'Unknown error'}`,
            'error',
            0,
            [
                { label: 'Report', onclick: 'dashboardUtils.reportError()' },
                { label: 'Dismiss', onclick: 'dashboardUtils.removeNotification(this.parentElement.parentElement)' }
            ]
        );
        
        // Track error for analytics
        this.trackMetric('error', { type, message: error.message, stack: error.stack });
    }

    reportError() {
        const errorData = {
            userAgent: navigator.userAgent,
            url: window.location.href,
            timestamp: new Date().toISOString(),
            metrics: this.metrics
        };
        
        // Send error report (implement endpoint)
        fetch('/api/error-report', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(errorData)
        }).catch(() => {
            this.showNotification('Failed to send error report', 'error');
        });
        
        this.showNotification('Error report sent', 'success');
    }

    setupNetworkErrorMonitoring() {
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            try {
                const response = await originalFetch(...args);
                
                if (!response.ok) {
                    this.handleError(
                        new Error(`HTTP ${response.status}: ${response.statusText}`),
                        'network'
                    );
                }
                
                return response;
            } catch (error) {
                this.handleError(error, 'network');
                throw error;
            }
        };
    }

    // WebSocket Enhancements
    setupWebSocketEnhancements() {
        // Enhanced reconnection logic
        this.setupRobustWebSocket();
        
        // Connection quality monitoring
        this.setupConnectionQualityMonitoring();
    }

    setupRobustWebSocket() {
        const originalWebSocket = window.WebSocket;
        
        window.WebSocket = class extends originalWebSocket {
            constructor(url, protocols) {
                super(url, protocols);
                this.setupEnhancedFeatures();
            }
            
            setupEnhancedFeatures() {
                this.reconnectAttempts = 0;
                this.maxReconnectAttempts = 10;
                this.reconnectDelay = 1000;
                
                const originalOnClose = this.onclose;
                this.onclose = (event) => {
                    if (originalOnClose) originalOnClose(event);
                    this.handleReconnect();
                };
                
                const originalOnError = this.onerror;
                this.onerror = (event) => {
                    if (originalOnError) originalOnError(event);
                    dashboardUtils.showNotification('WebSocket connection error', 'warning');
                };
            }
            
            handleReconnect() {
                if (this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.reconnectAttempts++;
                    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
                    
                    dashboardUtils.showNotification(
                        `Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`,
                        'info',
                        2000
                    );
                    
                    setTimeout(() => {
                        // Create new WebSocket connection
                        const newSocket = new WebSocket(this.url, this.protocols);
                        Object.assign(this, newSocket);
                    }, delay);
                } else {
                    dashboardUtils.showNotification('Failed to reconnect after multiple attempts', 'error');
                }
            }
        };
    }

    setupConnectionQualityMonitoring() {
        let lastPingTime = 0;
        let pings = [];
        
        const measurePing = () => {
            if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
                const start = Date.now();
                this.websocket.send(JSON.stringify({ type: 'ping' }));
                
                lastPingTime = start;
                setTimeout(() => {
                    const latency = Date.now() - lastPingTime;
                    pings.push(latency);
                    
                    // Keep only last 10 pings
                    if (pings.length > 10) pings.shift();
                    
                    const avgLatency = pings.reduce((a, b) => a + b, 0) / pings.length;
                    this.updateConnectionStatus(avgLatency);
                }, 1000);
            }
        };
        
        setInterval(measurePing, 30000);
    }

    updateConnectionStatus(latency) {
        const status = document.getElementById('connection-status');
        if (status) {
            const quality = latency < 100 ? 'good' : latency < 300 ? 'fair' : 'poor';
            status.innerHTML = `
                <span class="connection-quality connection-${quality}">
                    📡 ${quality} (${latency}ms)
                </span>
            `;
        }
    }

    // Local Storage Sync
    setupLocalStorageSync() {
        // Sync across tabs
        window.addEventListener('storage', (e) => {
            if (e.key === 'dashboard-theme') {
                this.theme = e.newValue;
                document.documentElement.setAttribute('data-theme', this.theme);
            }
        });
        
        // Periodic backup of important data
        setInterval(() => this.backupToLocalStorage(), 60000);
    }

    backupToLocalStorage() {
        const backup = {
            theme: this.theme,
            lastTab: document.querySelector('[data-active="true"]')?.dataset.tab,
            settings: this.getUserSettings(),
            timestamp: Date.now()
        };
        
        localStorage.setItem('dashboard-backup', JSON.stringify(backup));
    }

    getUserSettings() {
        return {
            syncInterval: parseInt(localStorage.getItem('sync-interval') || '5'),
            notifications: localStorage.getItem('notifications') !== 'disabled',
            autoRefresh: localStorage.getItem('auto-refresh') === 'enabled'
        };
    }

    // Plugin System
    registerPlugin(name, plugin) {
        if (this.plugins.has(name)) {
            console.warn(`Plugin ${name} already registered`);
            return false;
        }
        
        this.plugins.set(name, plugin);
        
        if (typeof plugin.init === 'function') {
            plugin.init();
        }
        
        console.log(`🔌 Plugin registered: ${name}`);
        return true;
    }

    unregisterPlugin(name) {
        if (!this.plugins.has(name)) return false;
        
        const plugin = this.plugins.get(name);
        if (typeof plugin.cleanup === 'function') {
            plugin.cleanup();
        }
        
        this.plugins.delete(name);
        console.log(`🔌 Plugin unregistered: ${name}`);
        return true;
    }

    // Utility Functions
    createElement(tag, options = {}) {
        const element = document.createElement(tag);
        Object.keys(options).forEach(key => {
            if (key === 'innerHTML' || key === 'textContent') {
                element[key] = options[key];
            } else if (key === 'onclick') {
                element.addEventListener('click', options[key]);
            } else {
                element.setAttribute(key, options[key]);
            }
        });
        return element;
    }

    trackMetric(event, data = {}) {
        this.metrics[event] = (this.metrics[event] || 0) + 1;
        
        // Send to analytics (implement endpoint)
        if (Math.random() < 0.1) { // 10% sampling
            fetch('/api/analytics', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ event, data, timestamp: Date.now() })
            }).catch(() => {}); // Ignore errors for analytics
        }
    }

    // Public API
    getMetrics() {
        return {
            ...this.metrics,
            uptime: Date.now() - this.metrics.pageLoad,
            plugins: Array.from(this.plugins.keys()),
            memoryUsage: performance.memory ? {
                used: performance.memory.usedJSHeapSize,
                total: performance.memory.totalJSHeapSize,
                limit: performance.memory.jsHeapSizeLimit
            } : null
        };
    }
}

// Initialize dashboard utilities
const dashboardUtils = new DashboardUtils();

// Export for global access
window.dashboardUtils = dashboardUtils;

// Default plugins
dashboardUtils.registerPlugin('shortcuts-help', {
    init() {
        dashboardUtils.addShortcut('ctrl+shift+/', () => {
            dashboardUtils.showNotification(`
                <strong>Keyboard Shortcuts:</strong><br>
                Ctrl+Shift+T: Toggle theme<br>
                Ctrl+Shift+H: High contrast<br>
                Ctrl+Shift+M: Reduced motion<br>
                Ctrl+Shift+/: Show this help<br>
                Ctrl+Shift+S: Settings<br>
                Ctrl+Shift+R: Refresh data
            `, 'info', 10000);
        });
    }
});

dashboardUtils.registerPlugin('performance-panel', {
    init() {
        const panel = dashboardUtils.createElement('div', {
            className: 'performance-panel',
            innerHTML: `
                <h4>Performance</h4>
                <div id="performance-stats"></div>
            `,
            style: 'position: fixed; bottom: 20px; right: 20px; background: rgba(0,0,0,0.8); color: white; padding: 10px; border-radius: 5px; font-size: 12px; z-index: 1000;'
        });
        
        document.body.appendChild(panel);
        
        setInterval(() => {
            const metrics = dashboardUtils.getMetrics();
            const stats = document.getElementById('performance-stats');
            if (stats) {
                stats.innerHTML = `
                    Interactions: ${metrics.interactions}<br>
                    Errors: ${metrics.errors}<br>
                    Uptime: ${Math.floor(metrics.uptime / 1000)}s<br>
                    ${metrics.memoryUsage ? `Memory: ${Math.round(metrics.memoryUsage.used / 1048576)}MB` : ''}
                `;
            }
        }, 1000);
    },
    
    cleanup() {
        document.querySelector('.performance-panel')?.remove();
    }
});

// Enhanced Alpine.js integration
document.addEventListener('alpine:init', () => {
    // Extend Alpine store with our utilities
    Alpine.store('utils', {
        theme: dashboardUtils.theme,
        showNotification: dashboardUtils.showNotification.bind(dashboardUtils),
        trackMetric: dashboardUtils.trackMetric.bind(dashboardUtils),
        
        toggleTheme() {
            dashboardUtils.toggleTheme();
            this.theme = dashboardUtils.theme;
        }
    });
    
    console.log('🔧 Enhanced Alpine.js store initialized');
});