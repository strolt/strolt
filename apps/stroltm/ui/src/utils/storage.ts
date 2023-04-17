const encode = (v: unknown) => JSON.stringify(v);
const decode = (v: string) => {
  try {
    return JSON.parse(v);
  } catch (e) {
    return null;
  }
};

interface Options<T> {
  defaultValue?: T;
  validate: (value: T) => Promise<void>;
}

const storage = <T = unknown>(key: string, options?: Options<T>) => {
  return {
    setItem: async (value: T) => {
      await options?.validate(value);

      localStorage.setItem(key, encode(value));
    },
    getItem: async (): Promise<T | null> => {
      const v = localStorage.getItem(key);
      if (!v) {
        return options?.defaultValue || null;
      }

      const decoded = decode(v);

      if (decoded) {
        await options?.validate(decoded);
      }

      return decoded || options?.defaultValue || null;
    },
    removeItem: async () => {
      localStorage.removeItem(key);
    },
  };
};

export const themeMode = storage<"light" | "dark">("theme:mode", {
  validate: async (value) => {
    if (!["light", "dark"].includes(value)) {
      throw new Error(`invalid theme mode '${value}'`);
    }
  },
});
