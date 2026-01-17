import type { ExecutionContext } from '@cloudflare/workers-types';
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import type { Env } from './types';
import { createLogger } from './utils/logger';
import { SessionService } from './services/session';
import { AIFallbackService } from './services/fallback';
import { ManimRendererService, MockRendererService, RendererService } from './services/manim';
import { TelegramHandler } from './handlers/telegram';
import { VideoHandler } from './handlers/video';
import { CodeHandler } from './handlers/code';
import { WhatsAppHandler } from './handlers/whatsapp';
import { DebugHandler } from './handlers/debug';

export interface ProcessingJob {
  id: string;
  userId: string;
  chatId: number;
  problem: string;
  status: 'queued' | 'ai_generating' | 'code_validating' | 'rendering' | 'uploading' | 'ready' | 'delivered' | 'failed';
  createdAt: number;
  updatedAt: number;
  error?: string;
  videoUrl?: string;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const logger = createLogger({
      level: (env.LOG_LEVEL as 'debug' | 'info' | 'warn' | 'error') || 'info',
      component: 'ai-manim-worker'
    });

    logger.info('Incoming request', {
      method: request.method,
      url: url.pathname,
    });

    try {
      const app = createApp(env, logger);
      return app.fetch(request, env, ctx);
    } catch (error) {
      logger.error('Request failed', error as Error, {
        method: request.method,
        url: url.pathname,
      });

      return new Response(
        JSON.stringify({
          status: 'error',
          message: (error as Error).message,
        }),
        { status: 500, headers: { 'Content-Type': 'application/json' } }
      );
    }
  },
};

function createApp(env: Env, logger: ReturnType<typeof createLogger>): Hono {
  const app = new Hono();

  app.use('/*', cors({
    origin: (origin) => origin,
    allowMethods: ['GET', 'POST', 'OPTIONS'],
    allowHeaders: ['Content-Type', 'Authorization', 'X-Telegram-Bot-Api-Secret-Token'],
  }));

  const sessionService = new SessionService(env);
  const aiService = new AIFallbackService(env);

  const useMockRenderer = env.USE_MOCK_RENDERER === 'true';
  let manimService: RendererService;

  if (useMockRenderer) {
    logger.info('Using mock renderer service');
    manimService = new MockRendererService();
  } else {
    const rendererUrl = env.MANIM_RENDERER_URL || 'http://localhost:8080';
    logger.info('Using real renderer service', { url: rendererUrl });
    manimService = new ManimRendererService({
      endpoint: rendererUrl,
      timeout: 300000,
      maxRetries: 3,
    });
  }

  const telegramHandler = new TelegramHandler(sessionService, aiService, manimService, env.TELEGRAM_SECRET);
  const videoHandler = new VideoHandler();
  const codeHandler = new CodeHandler(env);
  const whatsappHandler = new WhatsAppHandler(sessionService, aiService, manimService as ManimRendererService, env.WHATSAPP_WEBHOOK_SECRET || 'default_secret');
  const debugHandler = new DebugHandler();

  app.route('/telegram', telegramHandler.getApp());
  app.route('/video', videoHandler.getApp());
  app.route('/api/v1', codeHandler.getRouter());
  app.route('/webhook', whatsappHandler.getApp());
  app.route('/debug', debugHandler.getApp());

  // Dashboard routes
  app.get('/dashboard', (c) => {
    const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Manim Video Generator</title>
    <link rel="stylesheet" href="/styles/dashboard.css">
</head>
<body>
    <div class="container">
        <header class="header">
            <h1>🎬 AI Manim Video Generator</h1>
            <p class="subtitle">Mathematical visualizations powered by AI</p>
        </header>

        <main class="main">
            <!-- Job Input Section -->
            <section class="input-section">
                <h2>Submit a Problem</h2>
                <form id="problemForm" class="problem-form">
                    <textarea
                        id="problemInput"
                        placeholder="Describe a mathematical problem or concept you'd like visualized..."
                        rows="4"
                        maxlength="5000"
                        required
                    ></textarea>
                    <div class="form-footer">
                        <span class="char-count"><span id="charCount">0</span> / 5000</span>
                        <button type="submit" id="submitBtn" class="btn-primary">
                            <span class="btn-text">Generate Video</span>
                            <span class="btn-loader" style="display: none;">Generating...</span>
                        </button>
                    </div>
                </form>
            </section>

            <!-- Jobs List Section -->
            <section class="jobs-section">
                <h2>Your Videos</h2>
                <div id="jobsList" class="jobs-list">
                    <div class="empty-state">
                        <div class="empty-icon">📹</div>
                        <p>No videos yet. Submit a problem above to get started!</p>
                    </div>
                </div>
            </section>

            <!-- Video Player Section -->
            <section id="videoSection" class="video-section" style="display: none;">
                <h2>Video Preview</h2>
                <div class="video-container">
                    <video id="videoPlayer" controls preload="metadata">
                        Your browser does not support the video tag.
                    </video>
                    <div class="video-controls">
                        <button id="downloadBtn" class="btn-secondary">⬇️ Download</button>
                        <button id="closeVideoBtn" class="btn-secondary">✕ Close</button>
                    </div>
                </div>
                <div class="video-info">
                    <h3 id="videoTitle">Problem Title</h3>
                    <p id="videoDescription">Problem description</p>
                </div>
            </section>
        </main>

        <!-- Status Messages -->
        <div id="statusMessage" class="status-message" style="display: none;">
            <span class="status-text"></span>
            <button class="status-close" onclick="hideStatusMessage()">✕</button>
        </div>

        <!-- Job Status Modal -->
        <div id="jobModal" class="modal" style="display: none;">
            <div class="modal-content">
                <span class="modal-close" onclick="closeJobModal()">&times;</span>
                <h3>Job Status</h3>
                <div class="job-info">
                    <p><strong>Job ID:</strong> <span id="modalJobId"></span></p>
                    <p><strong>Problem:</strong> <span id="modalProblem"></span></p>
                    <p><strong>Status:</strong> <span id="modalStatus" class="status-badge"></span></p>
                    <p><strong>Created:</strong> <span id="modalCreated"></span></p>
                </div>
                <div class="job-progress">
                    <div class="progress-bar">
                        <div id="progressFill" class="progress-fill"></div>
                    </div>
                    <p id="progressText">Initializing...</p>
                </div>
            </div>
        </div>
    </div>

    <script src="/scripts/dashboard.js"></script>
</body>
</html>`;
    return c.html(html);
  });

  app.get('/styles/dashboard.css', (c) => {
    const css = `/* AI Manim Video Generator Dashboard Styles */

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

:root {
    --primary-color: #6366f1;
    --primary-hover: #4f46e5;
    --secondary-color: #64748b;
    --success-color: #10b981;
    --warning-color: #f59e0b;
    --error-color: #ef4444;
    --background: #ffffff;
    --surface: #f8fafc;
    --text-primary: #1e293b;
    --text-secondary: #64748b;
    --border: #e2e8f0;
    --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    background-color: var(--background);
    color: var(--text-primary);
    line-height: 1.6;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
}

/* Header */
.header {
    text-align: center;
    padding: 2rem 0;
    border-bottom: 1px solid var(--border);
    margin-bottom: 2rem;
}

.header h1 {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
}

.subtitle {
    font-size: 1.125rem;
    color: var(--text-secondary);
}

/* Main Content */
.main {
    display: grid;
    gap: 2rem;
}

/* Sections */
.input-section,
.jobs-section,
.video-section {
    background: var(--surface);
    border-radius: 12px;
    padding: 2rem;
    border: 1px solid var(--border);
}

.input-section h2,
.jobs-section h2,
.video-section h2 {
    font-size: 1.5rem;
    font-weight: 600;
    margin-bottom: 1.5rem;
    color: var(--text-primary);
}

/* Problem Form */
.problem-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

#problemInput {
    width: 100%;
    padding: 1rem;
    border: 2px solid var(--border);
    border-radius: 8px;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
    min-height: 120px;
    transition: border-color 0.2s;
}

#problemInput:focus {
    outline: none;
    border-color: var(--primary-color);
}

.form-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
}

.char-count {
    color: var(--text-secondary);
    font-size: 0.875rem;
}

/* Buttons */
.btn-primary,
.btn-secondary {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
}

.btn-primary {
    background-color: var(--primary-color);
    color: white;
}

.btn-primary:hover:not(:disabled) {
    background-color: var(--primary-hover);
}

.btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.btn-secondary {
    background-color: var(--secondary-color);
    color: white;
}

.btn-secondary:hover {
    background-color: #475569;
}

.btn-loader {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
}

.btn-loader::before {
    content: '';
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}

/* Jobs List */
.jobs-list {
    display: grid;
    gap: 1rem;
}

.job-card {
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    transition: all 0.2s;
    cursor: pointer;
}

.job-card:hover {
    border-color: var(--primary-color);
    box-shadow: var(--shadow);
}

.job-card.completed {
    border-left: 4px solid var(--success-color);
}

.job-card.failed {
    border-left: 4px solid var(--error-color);
}

.job-card.processing {
    border-left: 4px solid var(--warning-color);
}

.job-info {
    flex: 1;
}

.job-title {
    font-weight: 600;
    margin-bottom: 0.25rem;
    color: var(--text-primary);
}

.job-meta {
    font-size: 0.875rem;
    color: var(--text-secondary);
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
}

.job-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
}

.status-badge.queued {
    background-color: #dbeafe;
    color: #1e40af;
}

.status-badge.processing {
    background-color: #fef3c7;
    color: #92400e;
}

.status-badge.ready {
    background-color: #d1fae5;
    color: #065f46;
}

.status-badge.failed {
    background-color: #fee2e2;
    color: #991b1b;
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 3rem 1rem;
    color: var(--text-secondary);
}

.empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
}

/* Video Section */
.video-section {
    border: 2px solid var(--success-color);
    background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
}

.video-container {
    position: relative;
    margin-bottom: 1.5rem;
    border-radius: 8px;
    overflow: hidden;
    background: #000;
}

#videoPlayer {
    width: 100%;
    height: auto;
    max-height: 500px;
    display: block;
}

.video-controls {
    position: absolute;
    bottom: 1rem;
    right: 1rem;
    display: flex;
    gap: 0.5rem;
}

.video-controls button {
    background-color: rgba(0, 0, 0, 0.7);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    backdrop-filter: blur(4px);
}

.video-info h3 {
    margin-bottom: 0.5rem;
    color: var(--text-primary);
}

.video-info p {
    color: var(--text-secondary);
    line-height: 1.5;
}

/* Status Messages */
.status-message {
    position: fixed;
    top: 20px;
    right: 20px;
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1rem 1.5rem;
    box-shadow: var(--shadow-lg);
    display: flex;
    align-items: center;
    gap: 1rem;
    z-index: 1000;
    max-width: 400px;
}

.status-message.success {
    border-color: var(--success-color);
    background-color: #f0fdf4;
}

.status-message.error {
    border-color: var(--error-color);
    background-color: #fef2f2;
}

.status-message.warning {
    border-color: var(--warning-color);
    background-color: #fffbeb;
}

.status-close {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    color: var(--text-secondary);
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
}

/* Modal */
.modal {
    position: fixed;
    z-index: 1000;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
}

.modal-content {
    background-color: var(--background);
    margin: 5% auto;
    padding: 2rem;
    border-radius: 12px;
    border: 1px solid var(--border);
    width: 90%;
    max-width: 600px;
    position: relative;
    animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
    from {
        transform: translateY(-50px);
        opacity: 0;
    }
    to {
        transform: translateY(0);
        opacity: 1;
    }
}

.modal-close {
    position: absolute;
    right: 1rem;
    top: 1rem;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--text-secondary);
    background: none;
    border: none;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal h3 {
    margin-bottom: 1.5rem;
    color: var(--text-primary);
}

.job-info p {
    margin-bottom: 0.5rem;
    display: flex;
    justify-content: space-between;
}

.job-progress {
    margin-top: 1.5rem;
}

.progress-bar {
    width: 100%;
    height: 8px;
    background-color: var(--border);
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 0.5rem;
}

.progress-fill {
    height: 100%;
    background-color: var(--primary-color);
    transition: width 0.3s ease;
    width: 0%;
}

#progressText {
    font-size: 0.875rem;
    color: var(--text-secondary);
    text-align: center;
}

/* Responsive Design */
@media (max-width: 768px) {
    .container {
        padding: 0 1rem;
    }

    .header h1 {
        font-size: 2rem;
    }

    .input-section,
    .jobs-section,
    .video-section {
        padding: 1.5rem;
    }

    .form-footer {
        flex-direction: column;
        align-items: stretch;
    }

    .job-card {
        flex-direction: column;
        align-items: stretch;
        gap: 1rem;
    }

    .job-status {
        justify-content: space-between;
    }

    .video-controls {
        bottom: 0.5rem;
        right: 0.5rem;
    }

    .video-controls button {
        padding: 0.375rem 0.75rem;
        font-size: 0.75rem;
    }

    .modal-content {
        margin: 2rem auto;
        width: 95%;
        padding: 1.5rem;
    }

    .status-message {
        top: 10px;
        right: 10px;
        left: 10px;
        max-width: none;
    }
}

@media (max-width: 480px) {
    .header {
        padding: 1.5rem 0;
    }

    .header h1 {
        font-size: 1.75rem;
    }

    .input-section,
    .jobs-section,
    .video-section {
        padding: 1rem;
    }

    .job-meta {
        flex-direction: column;
        gap: 0.25rem;
    }

    .video-controls {
        flex-direction: column;
        gap: 0.25rem;
        bottom: 0.5rem;
        right: 0.5rem;
    }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
    * {
        animation-duration: 0.01ms !important;
        animation-iteration-count: 1 !important;
        transition-duration: 0.01ms !important;
    }
}

/* Focus styles for keyboard navigation */
.btn-primary:focus,
.btn-secondary:focus,
#problemInput:focus,
button:focus {
    outline: 2px solid var(--primary-color);
    outline-offset: 2px;
}`;
    return c.text(css, 200, { 'Content-Type': 'text/css' });
  });

  app.get('/scripts/dashboard.js', (c) => {
    const js = `// AI Manim Video Generator Dashboard JavaScript

class DashboardApp {
    constructor() {
        this.currentJobs = new Map();
        this.pollingIntervals = new Map();
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadJobs();
        this.setupCharacterCounter();
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
            jobsList.innerHTML = \`
                <div class="empty-state">
                    <div class="empty-icon">📹</div>
                    <p>No videos yet. Submit a problem above to get started!</p>
                </div>
            \`;
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

        return \`
            <div class="job-card \${statusClass}" data-job-id="\${job.id}">
                <div class="job-info">
                    <div class="job-title">\${job.problem}</div>
                    <div class="job-meta">
                        <span>Job: \${job.id.substring(0, 8)}</span>
                        <span>\${this.formatDate(job.created)}</span>
                        <span class="job-status">
                            <span class="status-badge \${job.status}">\${statusText}</span>
                        </span>
                    </div>
                </div>
            </div>
        \`;
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
        if (diff < 3600000) return \`\${Math.floor(diff / 60000)}m ago\`;
        if (diff < 86400000) return \`\${Math.floor(diff / 3600000)}h ago\`;
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
        a.download = \`manim-video-\${Date.now()}.mp4\`;
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
        modalStatus.className = \`status-badge \${status}\`;
        modalCreated.textContent = new Date().toLocaleString();

        modal.style.display = 'block';
        this.updateProgress(status);
    }

    updateJobModal(job) {
        const modalStatus = document.getElementById('modalStatus');
        modalStatus.textContent = this.getStatusText(job.status);
        modalStatus.className = \`status-badge \${job.status}\`;
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
                const response = await fetch(\`/api/v1/jobs/\${jobId}\`);
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

        statusMessage.className = \`status-message \${type}\`;
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
}`;
    return c.text(js, 200, { 'Content-Type': 'application/javascript' });
  });

  app.get('/health', async (c) => {
    return c.json({
      status: 'healthy',
      version: '1.0.0',
      timestamp: new Date().toISOString(),
      providers: {
        cloudflare: 'configured',
        groq: 'configured',
        huggingface: 'configured',
      },
    });
  });

  app.get('/ready', async (c) => {
    return c.json({
      ready: true,
      timestamp: new Date().toISOString(),
    });
  });

  return app;
}

export { createApp };
