import { createLogger } from '../utils/logger';
import { metric } from '../utils/metrics';

const logger = createLogger({ level: 'info', component: 'health-checks' });

interface HealthCheckResult {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  message?: string;
  responseTimeMs?: number;
}

interface ComponentHealth {
  telegram: HealthCheckResult;
  ai: HealthCheckResult;
  renderer: HealthCheckResult;
  r2: HealthCheckResult;
  kv: HealthCheckResult;
}

export async function checkComponentHealth(
  name: string,
  checkFn: () => Promise<boolean>
): Promise<HealthCheckResult> {
  const start = Date.now();

  try {
    const isHealthy = await checkFn();
    const responseTimeMs = Date.now() - start;

    metric.histogram('health_check_duration_ms', responseTimeMs, { component: name });

    return {
      name,
      status: isHealthy ? 'healthy' : 'unhealthy',
      responseTimeMs,
    };
  } catch (error) {
    const responseTimeMs = Date.now() - start;

    metric.increment('health_check_failure', 1, { component: name });

    logger.error(`Health check failed for ${name}`, error as Error);

    return {
      name,
      status: 'unhealthy',
      message: (error as Error).message,
      responseTimeMs,
    };
  }
}

export async function checkRelays(
  components: {
    telegram: () => Promise<boolean>;
    ai: () => Promise<boolean>;
    renderer: () => Promise<boolean>;
    r2: () => Promise<boolean>;
    kv: () => Promise<boolean>;
  }
): Promise<{
  telegram: HealthCheckResult;
  ai: HealthCheckResult;
  renderer: HealthCheckResult;
  r2: HealthCheckResult;
  kv: HealthCheckResult;
}> {
  logger.info('Starting health checks for all relays');

  const telegramPromise = checkComponentHealth('telegram', components.telegram);
  const aiPromise = checkComponentHealth('ai', components.ai);
  const rendererPromise = checkComponentHealth('renderer', components.renderer);
  const r2Promise = checkComponentHealth('r2', components.r2);
  const kvPromise = checkComponentHealth('kv', components.kv);

  const results = await Promise.allSettled([
    telegramPromise,
    aiPromise,
    rendererPromise,
    r2Promise,
    kvPromise,
  ]);

  const telegramResult =
    results[0].status === 'fulfilled' ? results[0].value : { name: 'telegram', status: 'unhealthy' as 'healthy' | 'unhealthy' | 'degraded', message: 'Check failed' };
  const aiResult =
    results[1].status === 'fulfilled' ? results[1].value : { name: 'ai', status: 'unhealthy' as 'healthy' | 'unhealthy' | 'degraded', message: 'Check failed' };
  const rendererResult =
    results[2].status === 'fulfilled' ? results[2].value : { name: 'renderer', status: 'unhealthy' as 'healthy' | 'unhealthy' | 'degraded', message: 'Check failed' };
  const r2Result =
    results[3].status === 'fulfilled' ? results[3].value : { name: 'r2', status: 'unhealthy' as 'healthy' | 'unhealthy' | 'degraded', message: 'Check failed' };
  const kvResult =
    results[4].status === 'fulfilled' ? results[4].value : { name: 'kv', status: 'unhealthy' as 'healthy' | 'unhealthy' | 'degraded', message: 'Check failed' };

  const health: ComponentHealth = {
    telegram: telegramResult,
    ai: aiResult,
    renderer: rendererResult,
    r2: r2Result,
    kv: kvResult,
  };

  const allHealthy = Object.values(health).every((h) => h.status === 'healthy');
  const someHealthy = Object.values(health).some((h) => h.status === 'healthy');
  const overallStatus = allHealthy ? 'healthy' : someHealthy ? 'degraded' : 'unhealthy';

  logger.info('Health checks completed', { overallStatus, components: Object.keys(health) });

  metric.gauge('system_health', overallStatus === 'healthy' ? 1 : overallStatus === 'degraded' ? 0.5 : 0);

  return health;
}

export function getOverallHealth(
  components: ComponentHealth
): { status: 'healthy' | 'degraded' | 'unhealthy'; details: ComponentHealth } {
  const allHealthy = Object.values(components).every((h) => h.status === 'healthy');
  const someHealthy = Object.values(components).some((h) => h.status === 'healthy');
  const status = allHealthy ? 'healthy' : someHealthy ? 'degraded' : 'unhealthy';

  logger.info('Overall health status', { status, allHealthy, someHealthy });

  return {
    status,
    details: components,
  };
}
