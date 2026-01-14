import type { LogLevel } from "../types";

export interface LoggerConfig {
  level: LogLevel;
  component: string;
}

export type LogLevel = "debug" | "info" | "warn" | "error";

interface LogEntry {
  level: string;
  message: string;
  component: string;
  ts: string;
  [key: string]: unknown;
}

function jqColorize(json: string): string {
  let result = "";
  for (let i = 0; i < json.length; i++) {
    const c = json[i];
    if (c === '"') {
      let str = "";
      let j = i + 1;
      while (j < json.length && (json[j] !== '"' || json[j - 1] === "\\")) {
        str += json[j];
        j++;
      }
      result += '\x1b[38;5;214m"' + str + '"\x1b[0m';
      i = j;
    } else if (c === "{" || c === "[") {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === "}" || c === "]") {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === ":") {
      result += '\x1b[38;5;39m:\x1b[0m';
    } else if (c === ",") {
      result += '\x1b[38;5;39m,\x1b[0m';
    } else if (json.slice(i, i + 4) === "true") {
      result += '\x1b[38;5;220mtrue\x1b[0m';
      i += 3;
    } else if (json.slice(i, i + 5) === "false") {
      result += '\x1b[38;5;220mfalse\x1b[0m';
      i += 4;
    } else if (json.slice(i, i + 4) === "null") {
      result += '\x1b[38;5;220mnull\x1b[0m';
      i += 3;
    } else if ((c >= "0" && c <= "9") || c === "-") {
      let num = "";
      let j = i;
      while (
        j < json.length &&
        ((json[j] >= "0" && json[j] <= "9") ||
          json[j] === "." ||
          json[j] === "e" ||
          json[j] === "E" ||
          json[j] === "+" ||
          json[j] === "-")
      ) {
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

const LEVEL_ORDER: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

export function createLogger(config: LoggerConfig) {
  const { level: minLevel, component } = config;

  function formatEntry(entry: LogEntry): string {
    const json = JSON.stringify(entry);
    return jqColorize(json);
  }

  function shouldLog(level: LogLevel): boolean {
    return LEVEL_ORDER[level] >= LEVEL_ORDER[minLevel];
  }

  return {
    info: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("info")) return;
      const entry: LogEntry = {
        level: "info",
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(formatEntry(entry));
    },

    error: (msg: string, error?: Error, data?: Record<string, unknown>) => {
      if (!shouldLog("error")) return;
      const entry: LogEntry = {
        level: "error",
        message: msg,
        component,
        ts: new Date().toISOString(),
        error: error?.message,
        stack: error?.stack,
        ...data,
      };
      console.log(formatEntry(entry));
    },

    debug: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("debug")) return;
      const entry: LogEntry = {
        level: "debug",
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(formatEntry(entry));
    },

    warn: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("warn")) return;
      const entry: LogEntry = {
        level: "warn",
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(formatEntry(entry));
    },

    log: (level: LogLevel, msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog(level)) return;
      const entry: LogEntry = {
        level,
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(formatEntry(entry));
    },
  };
}

export type Logger = ReturnType<typeof createLogger>;
