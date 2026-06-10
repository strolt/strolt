import type { AxiosResponse } from "axios";
import type { IPromiseBasedObservable } from "mobx-utils";

import { makeAutoObservable, reaction, runInAction } from "mobx";
import { fromPromise } from "mobx-utils";

import type * as apiGenerated from "../api/generated";

import * as api from "../api";
import { authStore } from "./auth.store";
import { managerStore } from "./manager.store";

export class InfoStore {
  latestVersion = "";

  map = new Map<string, apiGenerated.ManagerInfoInstance>();
  requestFetchInfo: IPromiseBasedObservable<AxiosResponse<apiGenerated.ApiInfo, any>> | null = null;
  updatedAt = new Date(0);
  version = "";

  constructor() {
    makeAutoObservable(this);
  }
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
  let intervalId: NodeJS.Timeout | null = null;
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
