import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

const LEVEL_ORDER: Record<string, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

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

function createLogger(config: { level: string; component: string }) {
  const { level: minLevel, component } = config;

  function formatEntry(entry: Record<string, unknown>): string {
    const json = JSON.stringify(entry);
    return jqColorize(json);
  }

  function shouldLog(level: string): boolean {
    return LEVEL_ORDER[level] >= LEVEL_ORDER[minLevel];
  }

  return {
    info: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("info")) return;
      const entry = { level: "info", message: msg, component, ts: new Date().toISOString(), ...data };
      console.log(formatEntry(entry));
    },
    error: (msg: string, error?: Error, data?: Record<string, unknown>) => {
      if (!shouldLog("error")) return;
      const entry = { level: "error", message: msg, component, ts: new Date().toISOString(), error: error?.message, ...data };
      console.log(formatEntry(entry));
    },
    debug: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("debug")) return;
      const entry = { level: "debug", message: msg, component, ts: new Date().toISOString(), ...data };
      console.log(formatEntry(entry));
    },
    warn: (msg: string, data?: Record<string, unknown>) => {
      if (!shouldLog("warn")) return;
      const entry = { level: "warn", message: msg, component, ts: new Date().toISOString(), ...data };
      console.log(formatEntry(entry));
    },
  };
}

describe("Logger", () => {
  let consoleSpy: vi.SpyInstance;

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  describe("log levels", () => {
    it("should log info messages when level is info", () => {
      const logger = createLogger({ level: "info", component: "test" });
      logger.info("test message", { key: "value" });

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("test message");
      expect(output).toContain("info");
      expect(output).toContain("test");
    });

    it("should not log debug messages when level is info", () => {
      const logger = createLogger({ level: "info", component: "test" });
      logger.debug("debug message");

      expect(consoleSpy).not.toHaveBeenCalled();
    });

    it("should log error messages", () => {
      const logger = createLogger({ level: "debug", component: "test" });
      const error = new Error("test error");
      logger.error("error message", error, { extra: "data" });

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("error message");
      expect(output).toContain("test error");
    });
  });

  describe("jqColorize output format", () => {
    it("should colorize strings in orange", () => {
      const logger = createLogger({ level: "debug", component: "test" });
      logger.info("hello");

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("hello");
    });

    it("should colorize numbers in green", () => {
      const logger = createLogger({ level: "debug", component: "test" });
      logger.info("number", { value: 42 });

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("42");
    });

    it("should colorize booleans in yellow", () => {
      const logger = createLogger({ level: "debug", component: "test" });
      logger.info("bool", { flag: true });

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("true");
    });

    it("should include timestamp", () => {
      const logger = createLogger({ level: "debug", component: "test" });
      logger.info("timestamp test");

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("ts");
      expect(output).toContain("2026");
    });

    it("should include component name", () => {
      const logger = createLogger({ level: "debug", component: "my-component" });
      logger.info("component test");

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(output).toContain("my-component");
    });
  });
});
