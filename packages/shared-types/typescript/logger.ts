export interface Env {
  LOG_LEVEL: string;
  API_GATEWAY_URL: string;
  AI_GATEWAY_URL: string;
}

export interface APIResponse {
  status: string;
  data: unknown;
  message: string;
}

export function jqColorize(json: string): string {
  let result = '';
  for (let i = 0; i < json.length; i++) {
    const c = json[i];
    if (c === '"') {
      let str = '';
      let j = i + 1;
      while (j < json.length && (json[j] !== '"' || json[j - 1] === '\\')) {
        str += json[j];
        j++;
      }
      result += '\x1b[38;5;214m"' + str + '"\x1b[0m';
      i = j;
    } else if (c === '{' || c === '[') {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === '}' || c === ']') {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === ':') {
      result += '\x1b[38;5;39m:\x1b[0m';
    } else if (c === ',') {
      result += '\x1b[38;5;39m,\x1b[0m';
    } else if (json.slice(i, i + 4) === 'true') {
      result += '\x1b[38;5;220mtrue\x1b[0m';
      i += 3;
    } else if (json.slice(i, i + 5) === 'false') {
      result += '\x1b[38;5;220mfalse\x1b[0m';
      i += 4;
    } else if (json.slice(i, i + 4) === 'null') {
      result += '\x1b[38;5;220mnull\x1b[0m';
      i += 3;
    } else if ((c >= '0' && c <= '9') || c === '-') {
      let num = '';
      let j = i;
      while (j < json.length && ((json[j] >= '0' && json[j] <= '9') || json[j] === '.' || json[j] === 'e' || json[j] === 'E' || json[j] === '+' || json[j] === '-')) {
        num += json[j];
        j++;
      }
      result += '\x1b[38;5;154m' + num + '\x1b[0m';
      i = j - 1;
    } else {
      result += c;
    }
  }
  return result;
}

export function createLogger(component: string) {
  return {
    info: (msg: string, data?: Record<string, unknown>) => {
      const entry = {
        level: 'info',
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
    error: (msg: string, error?: Error, data?: Record<string, unknown>) => {
      const entry = {
        level: 'error',
        message: msg,
        component,
        ts: new Date().toISOString(),
        error: error?.message,
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
    debug: (msg: string, data?: Record<string, unknown>) => {
      const entry = {
        level: 'debug',
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
  };
}

export const logger = createLogger('worker');
