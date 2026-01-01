document.addEventListener('DOMContentLoaded', () => {
    const servicesStatusContainer = document.getElementById('services-status');
    const pauseBotButton = document.getElementById('pause-bot');
    const resumeBotButton = document.getElementById('resume-bot');
    const aiProviderSelect = document.getElementById('ai-provider-select');
    const setProviderBtn = document.getElementById('set-provider-btn');

    const fetchServiceStatus = async () => {
        try {
            const response = await fetch('/api/services/status');
            if (!response.ok) {
                throw new Error('Failed to fetch service status');
            }
            const statuses = await response.json();
            renderServiceCards(statuses);
        } catch (error) {
            console.error('Error fetching service status:', error);
            servicesStatusContainer.innerHTML = '<p class="error">Could not load service statuses.</p>';
        }
    };

    const fetchAIProviders = async () => {
        try {
            const response = await fetch('/api/ai/providers');
            if (!response.ok) {
                throw new Error('Failed to fetch AI providers');
            }
            const providers = await response.json();
            populateProviderSelect(providers.available, providers.active);
        } catch (error) {
            console.error('Error fetching AI providers:', error);
        }
    };

    const populateProviderSelect = (available, active) => {
        aiProviderSelect.innerHTML = '';
        available.forEach(provider => {
            const option = document.createElement('option');
            option.value = provider;
            option.textContent = provider;
            if (provider === active) {
                option.selected = true;
            }
            aiProviderSelect.appendChild(option);
        });
    };

    const setAIProvider = async () => {
        const selectedProvider = aiProviderSelect.value;
        try {
            const response = await fetch('/api/ai/provider/set', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ provider: selectedProvider }),
            });
            if (!response.ok) {
                throw new Error('Failed to set AI provider');
            }
            fetchAIProviders(); // Refresh provider list
            fetchServiceStatus(); // Refresh service statuses
        } catch (error) {
            console.error('Error setting AI provider:', error);
        }
    };


    const renderServiceCards = (statuses) => {
        if (!statuses || statuses.length === 0) {
            servicesStatusContainer.innerHTML = '<p>No services to display.</p>';
            return;
        }

        servicesStatusContainer.innerHTML = ''; // Clear existing cards

        statuses.forEach(service => {
            const card = document.createElement('div');
            card.className = 'service-card';
            card.innerHTML = `
                <div class="service-header">
                    <div class="service-name">${service.name}</div>
                    <div class="status ${service.status}">${service.status}</div>
                </div>
                <div class="details">${service.details || ''}</div>
            `;
            servicesStatusContainer.appendChild(card);
        });
    };

    const controlBot = async (action) => {
        try {
            const response = await fetch(`/${action}`, { method: 'POST' });
            if (!response.ok) {
                throw new Error(`Failed to ${action} the bot`);
            }
            console.log(`Bot ${action} request sent successfully.`);
            fetchServiceStatus(); // Refresh statuses after action
        } catch (error) {
            console.error(`Error sending ${action} request:`, error);
        }
    };

    pauseBotButton.addEventListener('click', () => controlBot('pause'));
    resumeBotButton.addEventListener('click', () => controlBot('resume'));
    setProviderBtn.addEventListener('click', setAIProvider);

    // Initial fetches
    fetchServiceStatus();
    fetchAIProviders();

    // Set intervals for periodic fetching
    setInterval(fetchServiceStatus, 30000);
    setInterval(fetchAIProviders, 60000);
});
