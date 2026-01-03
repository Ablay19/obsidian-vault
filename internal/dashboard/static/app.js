document.addEventListener('alpine:init', () => {
    // Application state store
    Alpine.store('app', {
        syncInterval: 5,
        countdown: 5,
        lastUpdate: new Date(),
        isPaused: false,
        sidebarOpen: false,

        init() {
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
        },

        // Helper to check if polling should occur (e.g., pause on input focus)
        shouldPoll() {
            const active = document.activeElement;
            const isInput = active.tagName === 'INPUT' || 
                           active.tagName === 'TEXTAREA' || 
                           active.tagName === 'SELECT';
            return !isInput;
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
