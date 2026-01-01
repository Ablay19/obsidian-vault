// Fetch and update dashboard data
async function fetchStatus() {
    try {
        const response = await fetch('/api/services/status');
        const data = await response.json();
        
        // Assuming data is an array of service statuses
        const botStatusData = data.find(s => s.name === 'Bot Core');

        // Update status indicator
        const statusDot = document.getElementById('status-indicator');
        const statusText = document.getElementById('status-text');
        
        if (botStatusData && botStatusData.status === 'up') {
            statusDot.classList.remove('error');
            statusText.textContent = 'Online';
        } else {
            statusDot.classList.add('error');
            statusText.textContent = 'Offline';
        }
        
        // Update stats
        document.getElementById('bot-status').textContent = botStatusData?.status || '--';
        document.getElementById('uptime').textContent = formatUptime(botStatusData?.details.match(/Uptime: (.*?),/)?.[1]);
        document.getElementById('last-activity').textContent = formatTime(botStatusData?.details.match(/Last Activity: (.*)/)?.[1]);
        
        // Update system info - these are not in the current /api/services/status response directly
        // The Go backend needs to expose these via /api/services/status or a new endpoint
        document.getElementById('os').textContent = botStatusData?.os || '--'; // Placeholder
        document.getElementById('arch').textContent = botStatusData?.arch || '--'; // Placeholder
        document.getElementById('go-version').textContent = botStatusData?.go_version || '--'; // Placeholder
        document.getElementById('pid').textContent = botStatusData?.pid || '--'; // Placeholder
        
    } catch (error) {
        console.error('Error fetching status:', error);
        document.getElementById('status-text').textContent = 'Error';
        document.getElementById('status-indicator').classList.add('error');
    }
}

// Fetch AI providers
async function fetchProviders() {
    try {
        const response = await fetch('/api/ai/providers');
        const data = await response.json();
        
        // Update current provider
        document.getElementById('ai-provider').textContent = data.active || 'None';
        
        // Populate select dropdown
        const select = document.getElementById('provider-select');
        select.innerHTML = '';
        
        if (data.available && data.available.length > 0) {
            data.available.forEach(provider => {
                const option = document.createElement('option');
                option.value = provider;
                option.textContent = provider;
                if (provider === data.active) {
                    option.selected = true;
                }
                select.appendChild(option);
            });
        } else {
            select.innerHTML = '<option>No providers available</option>';
        }
        
        // Render provider cards
        renderProviders(data.available, data.active);
        
    } catch (error) {
        console.error('Error fetching providers:', error);
        document.getElementById('provider-select').innerHTML = '<option>Error loading</option>';
    }
}

// Render provider cards
function renderProviders(providers, active) {
    const container = document.getElementById('providers-list');
    
    if (!providers || providers.length === 0) {
        container.innerHTML = '<div class="loading">No providers configured</div>';
        return;
    }
    
    container.innerHTML = '';
    
    providers.forEach(provider => {
        const card = document.createElement('div');
        card.className = `provider-card ${provider === active ? 'active' : ''}`;
        
        card.innerHTML = `
            <span class="provider-name">${provider}</span>
            <span class="provider-badge ${provider === active ? 'active' : 'inactive'}">
                ${provider === active ? 'Active' : 'Available'}
            </span>
        `;
        
        container.appendChild(card);
    });
}

// Set AI provider
async function setProvider() {
    const select = document.getElementById('provider-select');
    const provider = select.value;
    const statusDiv = document.getElementById('provider-status');
    
    if (!provider) {
        showMessage(statusDiv, 'Please select a provider', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/ai/provider/set', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ provider })
        });
        
        if (response.ok) {
            showMessage(statusDiv, `Provider set to ${provider}`, 'success');
            // Refresh provider list
            setTimeout(fetchProviders, 500);
        } else {
            const error = await response.text();
            showMessage(statusDiv, `Error: ${error}`, 'error');
        }
    } catch (error) {
        console.error('Error setting provider:', error);
        showMessage(statusDiv, 'Failed to set provider', 'error');
    }
}

// Utility functions
function formatUptime(uptime) {
    if (!uptime) return '--';
    
    // Parse duration string like "2m45.843492578s"
    const match = uptime.match(/^(?:(\d+)h)?(?:(\d+)m)?(?:([\d.]+)s)?$/);
    if (!match) return uptime;
    
    const hours = parseInt(match[1]) || 0;
    const minutes = parseInt(match[2]) || 0;
    const seconds = Math.floor(parseFloat(match[3]) || 0);
    
    const parts = [];
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (seconds > 0 || parts.length === 0) parts.push(`${seconds}s`);
    
    return parts.join(' ');
}

function formatTime(timestamp) {
    if (!timestamp) return '--';
    
    const date = new Date(timestamp);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000);
    
    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    
    return date.toLocaleString();
}

function showMessage(element, message, type) {
    element.className = `info-message ${type}`;
    element.textContent = message;
    element.style.display = 'block';
    
    setTimeout(() => {
        element.style.display = 'none';
    }, 5000);
}

// Event listeners
document.getElementById('set-provider-btn').addEventListener('click', setProvider);

// Initial load and auto-refresh
fetchStatus();
fetchProviders();

setInterval(fetchStatus, 10000); // Refresh every 10 seconds
setInterval(fetchProviders, 30000); // Refresh providers every 30 seconds
