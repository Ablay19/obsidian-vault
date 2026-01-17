// AI Manim Video Generator Dashboard JavaScript

class DashboardApp {
    constructor() {
        this.currentJobs = new Map();
        this.pollingIntervals = new Map();
        this.currentPlatform = 'web';
        this.currentMode = 'problem';
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadJobs();
        this.setupCharacterCounter();
        this.updateCodeStats();
    }

    setupEventListeners() {
        // Form submission
        const form = document.getElementById('problemForm');
        form.addEventListener('submit', this.handleProblemSubmit.bind(this));

        // Video controls
        const downloadBtn = document.getElementById('downloadBtn');
        const closeVideoBtn = document.getElementById('closeVideoBtn');
        downloadBtn.addEventListener('click', this.handleDownload.bind(this));
        closeVideoBtn.addEventListener('click', this.closeVideoSection.bind(this));

        // Modal close
        const modalClose = document.querySelector('.modal-close');
        modalClose.addEventListener('click', this.closeJobModal.bind(this));

        // Close modal on background click
        const modal = document.getElementById('jobModal');
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeJobModal();
            }
        });

        // Platform selector
        const platformBtns = document.querySelectorAll('.platform-btn');
        platformBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.handlePlatformSelect(e.target.dataset.platform);
            });
        });

        // Mode selector
        const modeBtns = document.querySelectorAll('.mode-btn');
        modeBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.handleModeSelect(e.target.dataset.mode);
            });
        });

        // Code input controls
        const codeInput = document.getElementById('codeInput');
        codeInput.addEventListener('input', this.updateCodeStats.bind(this));

        const validateBtn = document.getElementById('validateBtn');
        validateBtn.addEventListener('click', this.handleCodeValidation.bind(this));

        const codeForm = document.getElementById('codeForm');
        codeForm.addEventListener('submit', this.handleCodeSubmit.bind(this));
    }

    setupCharacterCounter() {
        const problemInput = document.getElementById('problemInput');
        const charCount = document.getElementById('charCount');

        problemInput.addEventListener('input', () => {
            const count = problemInput.value.length;
            charCount.textContent = count;

            if (count > 4500) {
                charCount.style.color = 'var(--warning-color)';
            } else if (count > 5000) {
                charCount.style.color = 'var(--error-color)';
            } else {
                charCount.style.color = 'var(--text-secondary)';
            }
        });
    }

    handlePlatformSelect(platform) {
        // Update active button
        const platformBtns = document.querySelectorAll('.platform-btn');
        platformBtns.forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`.platform-btn[data-platform="${platform}"]`);
        activeBtn.classList.add('active');

        // Update instruction panel
        const instructionPanels = document.querySelectorAll('.instruction-panel');
        instructionPanels.forEach(panel => {
            panel.classList.remove('active');
        });
        const activePanel = document.querySelector(`.instruction-panel[data-platform="${platform}"]`);
        activePanel.classList.add('active');

        // Store current platform for form submissions
        this.currentPlatform = platform;
    }

    handleModeSelect(mode) {
        // Update active button
        const modeBtns = document.querySelectorAll('.mode-btn');
        modeBtns.forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`.mode-btn[data-mode="${mode}"]`);
        activeBtn.classList.add('active');

        // Update input section
        const inputSections = document.querySelectorAll('.input-section');
        inputSections.forEach(section => {
            section.classList.remove('active');
        });
        const activeSection = document.querySelector(`.input-section[data-mode="${mode}"]`);
        activeSection.classList.add('active');

        // Store current mode
        this.currentMode = mode;
    }

    updateCodeStats() {
        const codeInput = document.getElementById('codeInput');
        const codeLines = document.getElementById('codeLines');
        const codeChars = document.getElementById('codeChars');

        const code = codeInput.value;
        const lines = code.split('\n').length;
        const chars = code.length;

        codeLines.textContent = `${lines} lines`;
        codeChars.textContent = `${chars} characters`;
    }

    async handleCodeValidation() {
        const codeInput = document.getElementById('codeInput');
        const validateBtn = document.getElementById('validateBtn');
        const validationResult = document.getElementById('validationResult');
        const validationContent = document.getElementById('validationContent');

        const code = codeInput.value.trim();

        if (!code) {
            this.showStatusMessage('Please enter some Manim code to validate', 'error');
            return;
        }

        // Show loading state
        validateBtn.disabled = true;
        validateBtn.textContent = 'Validating...';

        try {
            const response = await fetch('/api/v1/validate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ code }),
            });

            const result = await response.json();

            if (response.ok) {
                validationContent.innerHTML = `
                    <div style="color: var(--success-color);">
                        <p>✅ Code validation successful!</p>
                        <p>Scene classes found: ${result.scene_classes?.join(', ') || 'None detected'}</p>
                    </div>
                `;
            } else {
                validationContent.innerHTML = `
                    <div style="color: var(--error-color);">
                        <p>❌ Validation failed:</p>
                        <pre style="background: #fee2e2; padding: 1rem; border-radius: 4px; margin-top: 0.5rem;">${result.error || 'Unknown error'}</pre>
                    </div>
                `;
            }

            validationResult.style.display = 'block';
        } catch (error) {
            console.error('Validation error:', error);
            validationContent.innerHTML = `
                <div style="color: var(--error-color);">
                    <p>❌ Network error during validation</p>
                </div>
            `;
            validationResult.style.display = 'block';
            this.showStatusMessage('Network error during validation', 'error');
        } finally {
            validateBtn.disabled = false;
            validateBtn.textContent = 'Validate Code';
        }
    }

    async handleCodeSubmit(event) {
        event.preventDefault();

        const codeInput = document.getElementById('codeInput');
        const qualitySelect = document.getElementById('codeQuality');
        const formatSelect = document.getElementById('codeFormat');
        const submitBtn = document.getElementById('submitCodeBtn');

        const code = codeInput.value.trim();
        const quality = qualitySelect.value;
        const format = formatSelect.value;

        if (!code) {
            this.showStatusMessage('Please enter Manim code', 'error');
            return;
        }

        if (code.length < 50) {
            this.showStatusMessage('Code must be at least 50 characters', 'error');
            return;
        }

        // Disable form and show loading
        this.setFormLoading(true, submitBtn);

        try {
            const response = await this.submitCode(code, quality, format);

            if (response.ok) {
                const result = await response.json();
                this.showStatusMessage('Code submitted successfully! Video rendering started.', 'success');
                codeInput.value = '';

                // Add job to list and start polling
                this.addJob(result.job_id, 'Direct Manim code submission', 'queued');
                this.startJobPolling(result.job_id);

                // Show job modal
                this.showJobModal(result.job_id, 'Direct Manim code submission', 'queued');
            } else {
                const error = await response.json();
                this.showStatusMessage(error.error || 'Failed to submit code', 'error');
            }
        } catch (error) {
            console.error('Submit error:', error);
            this.showStatusMessage('Network error. Please try again.', 'error');
        } finally {
            this.setFormLoading(false, submitBtn);
        }
    }

    async submitCode(code, quality, format) {
        const response = await fetch('/api/v1/code', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ code, quality, format }),
        });
        return response;
    }

    async handleProblemSubmit(event) {
        event.preventDefault();
        
        const problemInput = document.getElementById('problemInput');
        const submitBtn = document.getElementById('submitBtn');
        const problem = problemInput.value.trim();

        if (!problem) {
            this.showStatusMessage('Please enter a problem description', 'error');
            return;
        }

        if (problem.length < 10) {
            this.showStatusMessage('Problem description must be at least 10 characters', 'error');
            return;
        }

        // Disable form and show loading
        this.setFormLoading(true, submitBtn);
        
        try {
            const response = await this.submitProblem(problem);
            
            if (response.ok) {
                const result = await response.json();
                this.showStatusMessage('Problem submitted successfully! Video generation started.', 'success');
                problemInput.value = '';
                document.getElementById('charCount').textContent = '0';
                
                // Add job to list and start polling
                this.addJob(result.job_id, problem, 'queued');
                this.startJobPolling(result.job_id);
                
                // Show job modal
                this.showJobModal(result.job_id, problem, 'queued');
            } else {
                const error = await response.json();
                this.showStatusMessage(error.error || 'Failed to submit problem', 'error');
            }
        } catch (error) {
            console.error('Submit error:', error);
            this.showStatusMessage('Network error. Please try again.', 'error');
        } finally {
            this.setFormLoading(false, submitBtn);
        }
    }

    async submitProblem(problem) {
        const response = await fetch('/api/v1/jobs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ problem }),
        });
        return response;
    }

    setFormLoading(loading, button) {
        const btnText = button.querySelector('.btn-text');
        const btnLoader = button.querySelector('.btn-loader');
        
        button.disabled = loading;
        btnText.style.display = loading ? 'none' : 'inline';
        btnLoader.style.display = loading ? 'inline-flex' : 'none';
    }

    async loadJobs() {
        try {
            const response = await fetch('/api/v1/jobs');
            if (response.ok) {
                const jobs = await response.json();
                jobs.forEach(job => {
                    this.addJob(job.job_id, job.problem_preview || job.problem, job.status);
                    if (job.status === 'queued' || job.status === 'processing' || job.status === 'ai_generating') {
                        this.startJobPolling(job.job_id);
                    }
                });
            }
        } catch (error) {
            console.error('Load jobs error:', error);
        }
    }

    addJob(jobId, problem, status) {
        const job = {
            id: jobId,
            problem: problem.length > 50 ? problem.substring(0, 50) + '...' : problem,
            fullProblem: problem,
            status: status,
            created: new Date().toISOString(),
            videoUrl: null,
        };
        
        this.currentJobs.set(jobId, job);
        this.renderJobsList();
    }

    updateJobStatus(jobId, status, videoUrl = null) {
        const job = this.currentJobs.get(jobId);
        if (job) {
            job.status = status;
            if (videoUrl) {
                job.videoUrl = videoUrl;
            }
            this.renderJobsList();
            
            // Update modal if open
            const modalJobId = document.getElementById('modalJobId').textContent;
            if (modalJobId === jobId) {
                this.updateJobModal(job);
            }
        }
    }

    renderJobsList() {
        const jobsList = document.getElementById('jobsList');
        const jobs = Array.from(this.currentJobs.values()).sort((a, b) => 
            new Date(b.created) - new Date(a.created)
        );

        if (jobs.length === 0) {
            jobsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📹</div>
                    <p>No videos yet. Submit a problem above to get started!</p>
                </div>
            `;
            return;
        }

        jobsList.innerHTML = jobs.map(job => this.renderJobCard(job)).join('');
        
        // Add click handlers
        jobsList.querySelectorAll('.job-card').forEach(card => {
            card.addEventListener('click', () => {
                const jobId = card.dataset.jobId;
                this.handleJobClick(jobId);
            });
        });
    }

    renderJobCard(job) {
        const statusClass = this.getStatusClass(job.status);
        const statusText = this.getStatusText(job.status);
        
        return `
            <div class="job-card ${statusClass}" data-job-id="${job.id}">
                <div class="job-info">
                    <div class="job-title">${job.problem}</div>
                    <div class="job-meta">
                        <span>Job: ${job.id.substring(0, 8)}</span>
                        <span>${this.formatDate(job.created)}</span>
                        <span class="job-status">
                            <span class="status-badge ${job.status}">${statusText}</span>
                        </span>
                    </div>
                </div>
            </div>
        `;
    }

    getStatusClass(status) {
        switch (status) {
            case 'ready':
            case 'delivered':
                return 'completed';
            case 'failed':
                return 'failed';
            case 'queued':
            case 'ai_generating':
            case 'code_validating':
            case 'rendering':
            case 'uploading':
                return 'processing';
            default:
                return '';
        }
    }

    getStatusText(status) {
        const statusMap = {
            'queued': 'Queued',
            'ai_generating': 'AI Generating',
            'code_validating': 'Validating',
            'rendering': 'Rendering',
            'uploading': 'Uploading',
            'ready': 'Ready',
            'delivered': 'Delivered',
            'failed': 'Failed',
        };
        return statusMap[status] || status;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) return 'Just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
        return date.toLocaleDateString();
    }

    handleJobClick(jobId) {
        const job = this.currentJobs.get(jobId);
        if (!job) return;

        if (job.status === 'ready' && job.videoUrl) {
            this.showVideo(job);
        } else {
            this.showJobModal(jobId, job.fullProblem, job.status);
        }
    }

    showVideo(job) {
        const videoSection = document.getElementById('videoSection');
        const videoPlayer = document.getElementById('videoPlayer');
        const videoTitle = document.getElementById('videoTitle');
        const videoDescription = document.getElementById('videoDescription');

        videoPlayer.src = job.videoUrl;
        videoTitle.textContent = job.problem;
        videoDescription.textContent = job.fullProblem;
        
        videoSection.style.display = 'block';
        videoSection.scrollIntoView({ behavior: 'smooth' });
    }

    closeVideoSection() {
        const videoSection = document.getElementById('videoSection');
        const videoPlayer = document.getElementById('videoPlayer');
        
        videoSection.style.display = 'none';
        videoPlayer.pause();
        videoPlayer.src = '';
    }

    handleDownload() {
        const videoPlayer = document.getElementById('videoPlayer');
        const a = document.createElement('a');
        a.href = videoPlayer.src;
        a.download = `manim-video-${Date.now()}.mp4`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    }

    showJobModal(jobId, problem, status) {
        const modal = document.getElementById('jobModal');
        const modalJobId = document.getElementById('modalJobId');
        const modalProblem = document.getElementById('modalProblem');
        const modalStatus = document.getElementById('modalStatus');
        const modalCreated = document.getElementById('modalCreated');
        
        modalJobId.textContent = jobId.substring(0, 8);
        modalProblem.textContent = problem.length > 100 ? problem.substring(0, 100) + '...' : problem;
        modalStatus.textContent = this.getStatusText(status);
        modalStatus.className = `status-badge ${status}`;
        modalCreated.textContent = new Date().toLocaleString();
        
        modal.style.display = 'block';
        this.updateProgress(status);
    }

    updateJobModal(job) {
        const modalStatus = document.getElementById('modalStatus');
        modalStatus.textContent = this.getStatusText(job.status);
        modalStatus.className = `status-badge ${job.status}`;
        this.updateProgress(job.status);
    }

    updateProgress(status) {
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');
        
        const progressMap = {
            'queued': { width: '10%', text: 'Queued for processing...' },
            'ai_generating': { width: '30%', text: 'AI is generating Manim code...' },
            'code_validating': { width: '40%', text: 'Validating generated code...' },
            'rendering': { width: '70%', text: 'Rendering video animation...' },
            'uploading': { width: '90%', text: 'Uploading video...' },
            'ready': { width: '100%', text: 'Video ready!' },
            'delivered': { width: '100%', text: 'Video delivered' },
            'failed': { width: '0%', text: 'Processing failed' },
        };
        
        const progress = progressMap[status] || { width: '0%', text: 'Unknown status' };
        progressFill.style.width = progress.width;
        progressText.textContent = progress.text;
    }

    closeJobModal() {
        const modal = document.getElementById('jobModal');
        modal.style.display = 'none';
    }

    async startJobPolling(jobId) {
        // Clear existing interval for this job
        if (this.pollingIntervals.has(jobId)) {
            clearInterval(this.pollingIntervals.get(jobId));
        }

        const poll = async () => {
            try {
                const response = await fetch(`/api/v1/jobs/${jobId}`);
                if (response.ok) {
                    const job = await response.json();
                    this.updateJobStatus(jobId, job.status, job.video_url);
                    
                    // Stop polling if job is complete
                    if (job.status === 'ready' || job.status === 'delivered' || job.status === 'failed') {
                        clearInterval(this.pollingIntervals.get(jobId));
                        this.pollingIntervals.delete(jobId);
                        
                        if (job.status === 'ready') {
                            this.showStatusMessage('Your video is ready! Click to view.', 'success');
                        } else if (job.status === 'failed') {
                            this.showStatusMessage('Video generation failed. Please try again.', 'error');
                        }
                    }
                } else {
                    // Stop polling on error
                    clearInterval(this.pollingIntervals.get(jobId));
                    this.pollingIntervals.delete(jobId);
                }
            } catch (error) {
                console.error('Polling error:', error);
                // Continue polling on network errors
            }
        };

        // Start polling immediately
        poll();
        
        // Then poll every 5 seconds
        const interval = setInterval(poll, 5000);
        this.pollingIntervals.set(jobId, interval);
    }

    showStatusMessage(message, type = 'info') {
        const statusMessage = document.getElementById('statusMessage');
        const statusText = statusMessage.querySelector('.status-text');
        
        statusMessage.className = `status-message ${type}`;
        statusText.textContent = message;
        statusMessage.style.display = 'flex';
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            this.hideStatusMessage();
        }, 5000);
    }

    hideStatusMessage() {
        const statusMessage = document.getElementById('statusMessage');
        statusMessage.style.display = 'none';
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new DashboardApp();
});

// Global functions for inline event handlers
function hideStatusMessage() {
    window.dashboardApp?.hideStatusMessage();
}

function closeJobModal() {
    window.dashboardApp?.closeJobModal();
}