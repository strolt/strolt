import { AxiosResponse } from "axios";
import { makeAutoObservable, reaction, runInAction } from "mobx";

import { fromPromise, IPromiseBasedObservable } from "mobx-utils";

import * as api from "../api";
import * as apiGenerated from "../api/generated";
import { authStore } from "./auth.store";

export class ManagerStore {
  constructor() {
    makeAutoObservable(this);
  }

  isAutoUpdateInstancesEnabled = false;
  setIsAutoUpdateInstancesEnabled(status: boolean) {
    this.isAutoUpdateInstancesEnabled = status;
  }

  instances: apiGenerated.ManagerhGetInstancesResultItem[] = [];
  instancesStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ManagerhGetInstancesResult, any>
  > | null = null;
  async fetchInstances() {
    this.instancesStatus = fromPromise(api.manager.getInstances());

    const { data } = await this.instancesStatus;

    const sortedInstances = (data.data || []).sort((a, b) =>
      (a.instanceName || "").localeCompare(b.instanceName || ""),
    );

    runInAction(() => {
      this.instances = sortedInstances;
    });

    data.data = sortedInstances;

    return data;
  }
  resetInstances() {
    this.instancesStatus = null;
    this.instances = [];
  }

  backupAllStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ManagerhBackupAllResponse, any>
  > | null = null;
  async backupAll() {
    this.backupAllStatus = fromPromise(api.manager.backupAll());

    const { data } = await this.backupAllStatus;

    return data;
  }
  resetBackupAll() {
    this.backupAllStatus = null;
  }

  backupStatusMapKey(instanceName: string, serviceName: string, taskName: string) {
    return `${instanceName}_${serviceName}_${taskName}`;
  }

  backupStatusMap = new Map<
    string,
    IPromiseBasedObservable<AxiosResponse<apiGenerated.ApiuResultSuccess, any>>
  >();
  async backup(instanceName: string, serviceName: string, taskName: string) {
    const request = fromPromise(api.manager.backup(instanceName, serviceName, taskName));

    this.backupStatusMap.set(this.backupStatusMapKey(instanceName, serviceName, taskName), request);

    const { data } = await request;

    return data;
  }
  resetBackup() {
    this.backupStatusMap.clear();
  }

  snapshots: apiGenerated.ModelsServicesGetSnapshotsResult | null = null;
  snapshotsStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ModelsServicesGetSnapshotsResult, any>
  > | null = null;
  async fetchSnapshots(
    instanceName: string,
    serviceName: string,
    taskName: string,
    destinationName: string,
  ) {
    this.snapshotsStatus = fromPromise(
      api.manager.getSnapshots(instanceName, serviceName, taskName, destinationName),
    );

    const { data } = await this.snapshotsStatus;

    const sortedSnapshots = (data.data || []).sort((a, b) => {
      return new Date(b.time || "").getTime() - new Date(a.time || "").getTime();
    });

    data.data = sortedSnapshots;

    runInAction(() => {
      this.snapshots = data;
    });

    return data;
  }
  resetSnapshots() {
    this.snapshotsForPrune = null;
    this.snapshots = null;
  }

  snapshotsForPrune: apiGenerated.ModelsServicesGetPruneResult | null = null;
  snapshotsForPruneStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ModelsServicesGetPruneResult, any>
  > | null = null;
  async fetchSnapshotsForPrune(
    instanceName: string,
    serviceName: string,
    taskName: string,
    destinationName: string,
  ) {
    this.snapshotsForPruneStatus = fromPromise(
      api.manager.getSnapshotsForPrune(instanceName, serviceName, taskName, destinationName),
    );

    const { data } = await this.snapshotsForPruneStatus;

    const sortedSnapshots = (data.data || []).sort((a, b) => {
      return new Date(b.time || "").getTime() - new Date(a.time || "").getTime();
    });

    data.data = sortedSnapshots;

    runInAction(() => {
      this.snapshotsForPrune = data;
    });

    return data;
  }
  resetSnapshotsForPrune() {
    this.snapshotsForPruneStatus = null;
    this.snapshotsForPrune = null;
  }

  prune: apiGenerated.ModelsServicesGetPruneResult | null = null;
  pruneStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ModelsServicesGetPruneResult, any>
  > | null = null;
  async fetchPrune(
    instanceName: string,
    serviceName: string,
    taskName: string,
    destinationName: string,
  ) {
    this.pruneStatus = fromPromise(
      api.manager.prune(instanceName, serviceName, taskName, destinationName),
    );

    const { data } = await this.pruneStatus;

    const sortedSnapshots = (data.data || []).sort((a, b) => {
      return new Date(b.time || "").getTime() - new Date(a.time || "").getTime();
    });

    data.data = sortedSnapshots;

    runInAction(() => {
      this.prune = data;
    });

    return data;
  }
  resetPrune() {
    this.pruneStatus = null;
    this.prune = null;
  }
}

export const managerStore = new ManagerStore();

{
  let intervalId: null | NodeJS.Timer = null;
  reaction(
    () => ({
      isAuthorized: authStore.isAuthorized,
      isAutoUpdateInstancesEnabled: managerStore.isAutoUpdateInstancesEnabled,
    }),
    ({ isAuthorized, isAutoUpdateInstancesEnabled }) => {
      if (isAuthorized && isAutoUpdateInstancesEnabled) {
        managerStore.fetchInstances();
        intervalId = setInterval(() => managerStore.fetchInstances(), 5000);
      } else if (intervalId) {
        clearInterval(intervalId);
      }
    },
  );
}
