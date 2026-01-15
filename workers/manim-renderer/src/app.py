import os
import uuid
import subprocess
import threading
import time
from flask import Flask, request, jsonify
from datetime import datetime

app = Flask(__name__)

JOBS_DIR = os.environ.get('JOBS_DIR', '/tmp/manim_jobs')
os.makedirs(JOBS_DIR, exist_ok=True)

jobs: dict[str, dict] = {}
job_locks: dict[str, threading.Lock] = {}


def cleanup_job(job_id: str):
    time.sleep(3600)
    job_dir = os.path.join(JOBS_DIR, job_id)
    if os.path.exists(job_dir):
        import shutil
        shutil.rmtree(job_dir)
    jobs.pop(job_id, None)


@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy', 'timestamp': datetime.utcnow().isoformat()})


@app.route('/render', methods=['POST'])
def submit_render():
    data = request.json

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

        quality_map = {
            'low': '480p15',
            'medium': '720p30',
            'high': '1080p60',
        }
        quality_flag = quality_map.get(job.get('quality', 'medium'), '720p30')

        output_file = os.path.join(job['job_dir'], f'scene.{job["output_format"]}')

        try:
            env = os.environ.copy()
            env['MANIM_PREVIEW'] = 'false'

            quality = job.get('quality', 'medium')

            cmd = [
                'manim',
                '-ql',
                '--format', job['output_format'],
                '--quality', quality_flag,
                '-o', output_file,
                job['code_file'],
                'Scene',
            ]

            if quality == 'high':
                cmd = ['manim', '-qh', '--format', job['output_format'], '-o', output_file, job['code_file'], 'Scene']
            elif quality == 'low':
                cmd = ['manim', '-ql', '--format', job['output_format'], '-o', output_file, job['code_file'], 'Scene']

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=600,
                env=env,
            )

            if result.returncode == 0 and os.path.exists(output_file):
                job['status'] = 'complete'
                job['video_path'] = output_file
                job['completed_at'] = datetime.utcnow().isoformat()

                file_size = os.path.getsize(output_file)
                job['file_size'] = file_size

            else:
                job['status'] = 'failed'
                job['error'] = result.stderr or 'Unknown rendering error'

        except subprocess.TimeoutExpired:
            job['status'] = 'failed'
            job['error'] = 'Rendering timeout (600s)'
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
    elif job['status'] == 'complete':
        response['video_url'] = f'/download/{job_id}'
        response['completed_at'] = job.get('completed_at')
        response['file_size'] = job.get('file_size')
    elif job['status'] == 'failed':
        response['error'] = job.get('error')

    return jsonify(response)


@app.route('/download/<job_id>', methods=['GET'])
def download_video(job_id: str):
    job = jobs.get(job_id)
    if not job or job['status'] != 'complete':
        return jsonify({'error': 'Video not found'}), 404

    from flask import send_file
    return send_file(
        job['video_path'],
        mimetype=f'video/{job["output_format"]}',
        as_attachment=True,
        download_name=f'{job_id}.{job["output_format"]}',
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
    app.run(host='0.0.0.0', port=8080)
