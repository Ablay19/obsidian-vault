export interface WorkerModule {
  id: string;
  name: string;
  version: string;
  description: string;
  entryPoint: string;
  dependencies: string[];
  environment: Record<string, string>;
  routes: RouteMapping[];
  permissions: string[];
  resources: ResourceLimits;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface RouteMapping {
  path: string;
  method: string;
  target: string;
  timeout: number;
  headers: Record<string, string>;
}

export interface ResourceLimits {
  cpu: string;
  memory: string;
  storage: string;
}

export interface GoApplication {
  id: string;
  name: string;
  version: string;
  description: string;
  modulePath: string;
  entryPoint: string;
  port: number;
  database: DatabaseConfig;
  apis: APIEndpoint[];
  dependencies: GoDependency[];
  resources: ResourceLimits;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface DatabaseConfig {
  type: string;
  host: string;
  port: number;
  name: string;
  migrations: string;
}

export interface APIEndpoint {
  path: string;
  method: string;
  description: string;
  authentication: boolean;
  rateLimit: RateLimitConfig;
}

export interface RateLimitConfig {
  requests: number;
  window: string;
}

export interface GoDependency {
  name: string;
  version: string;
}

export interface SharedPackage {
  id: string;
  name: string;
  version: string;
  type: string;
  languages: string[];
  goModule?: GoPackageConfig;
  npmPackage?: NpmPackageConfig;
  dependencies: string[];
  exports: string[];
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface GoPackageConfig {
  modulePath: string;
  goVersion: string;
  imports: string[];
}

export interface NpmPackageConfig {
  packageName: string;
  main: string;
  types: string;
  scripts: Record<string, string>;
}

export interface APIGateway {
  id: string;
  name: string;
  version: string;
  routes: Route[];
  middlewares: Middleware[];
  rateLimit: RateLimitConfig;
  cors: CORSConfig;
  auth: AuthConfig;
}

export interface Route {
  method: string;
  path: string;
  target: string;
  timeout: number;
  retry: RetryConfig;
}

export interface Middleware {
  name: string;
  config: Record<string, unknown>;
}

export interface RetryConfig {
  maxAttempts: number;
  backoff: number;
  strategy: string;
}

export interface CORSConfig {
  allowedOrigins: string[];
  allowedMethods: string[];
  allowedHeaders: string[];
  maxAge: number;
}

export interface AuthConfig {
  enabled: boolean;
  type: string;
  rateLimit: boolean;
}

export interface DeploymentPipeline {
  id: string;
  name: string;
  componentId: string;
  componentType: string;
  triggers: TriggerConfig[];
  stages: PipelineStage[];
  environment: EnvironmentConfig;
  notifications: NotificationConfig[];
  lastDeployment?: DeploymentStatus;
  createdAt: string;
  updatedAt: string;
}

export interface TriggerConfig {
  type: string;
  config: Record<string, unknown>;
}

export interface PipelineStage {
  name: string;
  type: string;
  commands: string[];
  timeout: number;
  onFailure: string;
}

export interface EnvironmentConfig {
  name: string;
  type: string;
  variables: Record<string, string>;
}

export interface NotificationConfig {
  type: string;
  target: string;
  events: string[];
}

export interface DeploymentStatus {
  id: string;
  status: string;
  startTime: string;
  endTime?: string;
  version: string;
  environment: string;
}

export interface APIResponse {
  status: string;
  data: unknown;
  message: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface LogConfig {
  level: string;
  format: string;
  output: string;
  timeFormat: string;
}

export class StructuredLogger {
  constructor(private component: string, private level: string = 'info') {}

  info(message: string, data?: Record<string, unknown>): void {
    this.log('info', message, data);
  }

  warn(message: string, data?: Record<string, unknown>): void {
    this.log('warn', message, data);
  }

  error(message: string, error?: Error, data?: Record<string, unknown>): void {
    this.log('error', message, { ...data, error: error?.message, stack: error?.stack });
  }

  debug(message: string, data?: Record<string, unknown>): void {
    this.log('debug', message, data);
  }

  private log(level: string, message: string, data?: Record<string, unknown>): void {
    const logEntry: Record<string, unknown> = {
      level,
      message,
      component: this.component,
      timestamp: new Date().toISOString(),
      ...data,
    };

    console.log(JSON.stringify(logEntry));
  }
}

export function createLogger(component: string, level?: string): StructuredLogger {
  return new StructuredLogger(component, level);
}