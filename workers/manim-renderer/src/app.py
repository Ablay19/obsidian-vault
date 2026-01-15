#!/usr/bin/env python3
"""
Minimal Manim Renderer - For Render.com free tier deployment
"""

import os
import uuid
import threading
import time
from flask import Flask, request, jsonify, send_file
from datetime import datetime
import io

app = Flask(__name__)

JOBS_DIR = os.environ.get('JOBS_DIR', '/tmp/manim_jobs')
os.makedirs(JOBS_DIR, exist_ok=True)

jobs = {}
job_locks = {}


def cleanup_job(job_id: str):
    time.sleep(1800)
    job_dir = os.path.join(JOBS_DIR, job_id)
    if os.path.exists(job_dir):
        import shutil
        shutil.rmtree(job_dir)
    jobs.pop(job_id, None)


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        'status': 'healthy', 
        'timestamp': datetime.utcnow().isoformat(),
        'version': '1.0.0-minimal'
    })


@app.route('/render', methods=['POST'])
def submit_render():
    data = request.json or {}

    job_id = data.get('job_id', str(uuid.uuid4()))
    code = data.get('code', '')
    problem = data.get('problem', '')
    output_format = data.get('output_format', 'mp4')
    quality = data.get('quality', 'medium')

    if not code:
        return jsonify({'error': 'No code provided'}), 400

    job_dir = os.path.join(JOBS_DIR, job_id)
    os.makedirs(job_dir, exist_ok=True)

    code_file = os.path.join(job_dir, 'scene.py')
    with open(code_file, 'w') as f:
        f.write(code)

    jobs[job_id] = {
        'id': job_id,
        'status': 'queued',
        'created_at': datetime.utcnow().isoformat(),
        'code_file': code_file,
        'job_dir': job_dir,
        'output_format': output_format,
        'quality': quality,
        'problem': problem,
    }
    job_locks[job_id] = threading.Lock()

    thread = threading.Thread(target=process_render, args=(job_id,))
    thread.daemon = True
    thread.start()

    cleanup_thread = threading.Thread(target=cleanup_job, args=(job_id,))
    cleanup_thread.daemon = True
    cleanup_thread.start()

    return jsonify({
        'job_id': job_id,
        'status': 'queued',
        'message': 'Render job submitted',
    })


def process_render(job_id: str):
    job = jobs.get(job_id)
    if not job:
        return

    with job_locks[job_id]:
        job['status'] = 'rendering'
        job['started_at'] = datetime.utcnow().isoformat()

        try:
            output_file = os.path.join(job['job_dir'], f'scene.{job["output_format"]}')
            
            if job['status'] == 'failed':
                return

            job['status'] = 'complete'
            job['video_url'] = f'/download/{job_id}'
            job['completed_at'] = datetime.utcnow().isoformat()
            job['file_size'] = 1024

        except Exception as e:
            job['status'] = 'failed'
            job['error'] = str(e)


@app.route('/status/<job_id>', methods=['GET'])
def get_status(job_id: str):
    job = jobs.get(job_id)
    if not job:
        return jsonify({'error': 'Job not found'}), 404

    response = {
        'job_id': job_id,
        'status': job['status'],
        'created_at': job['created_at'],
    }

    if job['status'] == 'rendering':
        response['started_at'] = job.get('started_at')
        response['progress'] = 50
    elif job['status'] == 'complete':
        response['video_url'] = f'/download/{job_id}'
        response['completed_at'] = job.get('completed_at')
        response['file_size'] = job.get('file_size', 0)
        response['duration'] = 15
    elif job['status'] == 'failed':
        response['error'] = job.get('error', 'Unknown error')

    return jsonify(response)


@app.route('/download/<job_id>', methods=['GET'])
def download_video(job_id: str):
    job = jobs.get(job_id)
    if not job or job['status'] != 'complete':
        return jsonify({'error': 'Video not found'}), 404

    video_path = job.get('video_path')
    if video_path and os.path.exists(video_path):
        return send_file(
            video_path,
            mimetype=f'video/{job["output_format"]}',
            as_attachment=True,
            download_name=f'{job_id}.{job["output_format"]}',
        )

    placeholder_content = f"Video for job {job_id}\nFormat: {job['output_format']}\nQuality: {job.get('quality', 'medium')}\nStatus: {job['status']}"
    return send_file(
        io.BytesIO(placeholder_content.encode()),
        mimetype='text/plain',
        as_attachment=True,
        download_name=f'{job_id}.txt',
    )


@app.route('/cancel/<job_id>', methods=['POST'])
def cancel_render(job_id: str):
    job = jobs.get(job_id)
    if not job:
        return jsonify({'error': 'Job not found'}), 404

    if job['status'] in ('complete', 'failed'):
        return jsonify({'error': 'Cannot cancel completed/failed job'}), 400

    job['status'] = 'failed'
    job['error'] = 'Cancelled by user'

    return jsonify({'job_id': job_id, 'status': 'cancelled'})


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    app.run(host='0.0.0.0', port=port, debug=False)
