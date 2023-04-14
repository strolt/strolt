import { AxiosResponse } from "axios";
import { makeAutoObservable, reaction, runInAction } from "mobx";

import { fromPromise, IPromiseBasedObservable } from "mobx-utils";

import * as api from "../api";
import * as apiGenerated from "../api/generated";
import { authStore } from "./auth.store";
import { managerStore } from "./manager.store";

export class InfoStore {
  constructor() {
    makeAutoObservable(this);
  }

  latestVersion = "";
  version = "";
  updatedAt = new Date(0);
  map = new Map<string, apiGenerated.ApiInfoInstance>();

  requestFetchInfo: IPromiseBasedObservable<AxiosResponse<apiGenerated.ApiInfo, any>> | null = null;
  async fetchInfo() {
    this.requestFetchInfo = fromPromise(api.global.getInfo());
    const { data } = await this.requestFetchInfo;

    runInAction(() => {
      const newUpdatedAt = new Date(data.updatedAt || "");

      if (newUpdatedAt.getTime() != this.updatedAt.getTime()) {
        managerStore.fetchInstances();
      }

      this.latestVersion = data.latestVersion || "";
      this.version = data.version || "";
      this.updatedAt = newUpdatedAt;
      this.map.clear();
      data.instances?.forEach((instance) => {
        if (instance && instance.name) {
          this.map.set(this.getKey(instance.name, instance.proxyName), instance);
        }
      });
    });
  }

  getKey(instanceName: string, proxyName?: string) {
    return `${proxyName}_${instanceName}`;
  }
}

export const infoStore = new InfoStore();

{
  let intervalId: null | NodeJS.Timer = null;
  reaction(
    () => ({
      isAuthorized: authStore.isAuthorized,
    }),
    ({ isAuthorized }) => {
      if (isAuthorized) {
        infoStore.fetchInfo();
        intervalId = setInterval(() => infoStore.fetchInfo(), 5000);
      } else if (intervalId) {
        clearInterval(intervalId);
      }
    },
  );
}
