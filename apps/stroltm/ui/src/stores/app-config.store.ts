import { makeAutoObservable, runInAction } from "mobx";
import { themeMode } from "utils/storage";

class AppConfigStore {
  mode: "dark" | "light" = "light";

  constructor() {
    makeAutoObservable(this);
  }

  toggleMode() {
    if (this.mode === "dark") {
      this.mode = "light";
    } else {
      this.mode = "dark";
    }
    themeMode.setItem(this.mode);
  }
}

export const appConfigStore = new AppConfigStore();

if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
  runInAction(() => {
    appConfigStore.mode = "dark";
  });

  themeMode.getItem().then((v) => {
    if (v) {
      runInAction(() => {
        appConfigStore.mode = v;
      });
    }
  });
}
