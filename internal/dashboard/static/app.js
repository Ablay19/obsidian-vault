document.addEventListener('alpine:init', () => {
    // Application state store
    Alpine.store('app', {
        syncInterval: 5,
        countdown: 5,
        lastUpdate: new Date(),
        isPaused: false,
        sidebarOpen: false,
        activeTab: window.initialTab || 'overview',

        init() {
            // Persist initial tab not needed as much with direct routes, 
            // but we can keep it for any history restoration if we wanted.
            
            // WebSocket Connection
            this.initWebSocket();

            // Polling interval logic
            setInterval(() => {
                if (!this.isPaused && this.shouldPoll()) {
                    this.countdown--;
                    if (this.countdown < 0) {
                        this.countdown = this.syncInterval;
                        this.triggerSync();
                    }
                }
            }, 1000);
            
            // Handle back/forward buttons
            window.addEventListener('hashchange', () => {
                const newTab = window.location.hash.replace('#', '') || 'overview';
                if (newTab !== this.activeTab) {
                    this.activeTab = newTab;
                    localStorage.setItem('activeTab', newTab);
                    this.syncUI();
                }
            });

            // HTMX integration for state updates
            document.addEventListener('htmx:afterSettle', (evt) => {
                // When content is swapped, ensure Alpine components are initialized
                // and UI state reflects current store values.
            });
        },

        initWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            console.log('Connecting to WebSocket...', wsUrl);
            const socket = new WebSocket(wsUrl);

            socket.onmessage = (event) => {
                const data = JSON.parse(event.data);
                console.debug('WS Event:', data.type, data.payload);
                
                // Dispatch custom event for specific components
                window.dispatchEvent(new CustomEvent(`ws:${data.type}`, { detail: data.payload }));
                
                // Global reactions
                if (data.type === 'notification') {
                    $store.notifications.show(data.payload.message, data.payload.level);
                }
            };

            socket.onclose = () => {
                console.warn('WebSocket disconnected. Reconnecting in 5s...');
                setTimeout(() => this.initWebSocket(), 5000);
            };

            socket.onerror = (err) => {
                console.error('WebSocket error:', err);
            };
        },

        setActiveTab(tab) {
            this.activeTab = tab;
            window.location.hash = tab;
            localStorage.setItem('activeTab', tab);
        },

        // Helper to force UI to match activeTab (triggers HTMX)
        syncUI() {
            const link = document.querySelector(`[hx-target="#main-content"][href="#${this.activeTab}"]`);
            if (link) {
                htmx.trigger(link, 'click');
            } else {
                // Fallback for direct URL access or missing elements
                const urlMap = {
                    'overview': '/dashboard/panels/overview',
                    'status': '/dashboard/panels/service_status',
                    'providers': '/dashboard/panels/ai_providers',
                    'keys': '/dashboard/panels/api_keys',
                    'history': '/dashboard/panels/chat_history',
                    'chat': '/dashboard/panels/qa_console',
                    'env': '/dashboard/panels/environment',
                    'whatsapp': '/dashboard/panels/whatsapp'
                };
                const targetUrl = urlMap[this.activeTab] || urlMap['overview'];
                htmx.ajax('GET', targetUrl, '#main-content');
            }
        },

        // Helper to check if polling should occur (e.g., pause on input focus)
        shouldPoll() {
            const active = document.activeElement;
            const isInput = active.tagName === 'INPUT' || 
                           active.tagName === 'TEXTAREA' || 
                           active.tagName === 'SELECT' ||
                           active.isContentEditable;
            return !isInput && !this.isPaused;
        },

        // Manually trigger a UI sync if needed
        triggerSync() {
            // HTMX handles the actual fetching, this is just for UI sync state
            this.lastUpdate = new Date();
        },

        resetCountdown() {
            this.countdown = this.syncInterval;
            this.lastUpdate = new Date();
        },

        togglePause() {
            this.isPaused = !this.isPaused;
        },
        
        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen;
        }
    });

    // QA Store for persistence
    Alpine.store('qa', {
        draft: '',
        messages: [], // Store recent local messages to prevent loss on swap
        
        addMessage(direction, text) {
            this.messages.push({
                direction: direction,
                text: text,
                timestamp: new Date()
            });
            // Keep only last 20 messages in local memory
            if (this.messages.length > 20) this.messages.shift();
        },
        
        clear() {
            this.messages = [];
            this.draft = '';
        }
    });

    // Notification store using Toastify
    Alpine.store('notifications', {
        show(message, type = 'success') {
            if (typeof Toastify === 'undefined') {
                console.warn('Toastify not loaded', message);
                return;
            }
            Toastify({
                text: message,
                duration: 3000,
                close: true,
                gravity: "top",
                position: "right",
                stopOnFocus: true,
                className: "font-mono text-xs",
                style: {
                    background: type === 'success' ? "#059669" : "#dc2626",
                    boxShadow: "0 4px 12px rgba(0,0,0,0.5)",
                    borderRadius: "8px",
                    border: "1px solid rgba(255,255,255,0.1)"
                }
            }).showToast();
        }
    });

    // Chart component logic
    Alpine.data('systemLoadChart', (initialData) => ({
        chart: null,
        init() {
            if (!this.$refs.canvas) return;
            
            this.chart = new Chart(this.$refs.canvas, {
                type: 'line',
                data: {
                    labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'],
                    datasets: [{
                        label: 'Processed',
                        data: initialData || [0, 0, 0, 0, 0, 0, 0], // Default data
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.4,
                        fill: true,
                        pointRadius: 0
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    animation: false, // Disable animation for performance on updates
                    plugins: { legend: { display: false } },
                    scales: {
                        y: { display: false },
                        x: { 
                            grid: { display: false }, 
                            ticks: { color: '#4b5563', font: { size: 8 } } 
                        }
                    }
                }
            });

            // Clean up chart instance when component is destroyed (e.g. HTMX swap)
            // Alpine automatically handles cleanup for registered listeners, 
            // but chart instances need manual destruction if the DOM element is removed.
        },
        destroy() {
            if (this.chart) {
                this.chart.destroy();
                this.chart = null;
            }
        }
    }));
});
