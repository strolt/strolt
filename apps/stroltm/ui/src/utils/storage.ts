const encode = (v: unknown) => JSON.stringify(v);
const decode = (v: string) => {
  try {
    return JSON.parse(v);
  } catch {
    return null;
  }
};

interface Options<T> {
  defaultValue?: T;
  validate: (value: T) => Promise<void>;
}

const storage = <T = unknown>(key: string, options?: Options<T>) => {
  return {
    getItem: async (): Promise<null | T> => {
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
    setItem: async (value: T) => {
      await options?.validate(value);

      localStorage.setItem(key, encode(value));
    },
  };
};

export const themeMode = storage<"dark" | "light">("theme:mode", {
  validate: async (value) => {
    if (!["dark", "light"].includes(value)) {
      throw new Error(`invalid theme mode '${value}'`);
    }
  },
});
