import { AxiosResponse } from "axios";
import { makeAutoObservable, runInAction } from "mobx";

import { fromPromise, IPromiseBasedObservable } from "mobx-utils";

import * as api from "../../api";
import * as apiGenerated from "../../api/generated";

import { getTaskList } from "./taskList";

export class ManagerStore {
  constructor() {
    makeAutoObservable(this);
  }

  instances: apiGenerated.ManagerhManagerPreparedInstance[] = [];
  instancesStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ManagerhManagerPreparedInstance[], any>
  > | null = null;
  async fetchInstances() {
    this.instancesStatus = fromPromise(api.manager.getInstances());

    const { data } = await this.instancesStatus;

    runInAction(() => {
      data?.forEach((instance) => {
        instance.taskStatus?.tasks?.forEach((taskStatus) => {
          if (instance.name && taskStatus.serviceName && taskStatus.taskName) {
            this.taskStatusMap.set(
              this.getTaskStatusMapKey(
                instance.name,
                taskStatus.serviceName,
                taskStatus.taskName,
                instance.proxyName,
              ),
              taskStatus,
            );
          }
        });
      });

      this.instances = data.sort((a, b) => (a.name || "").localeCompare(b.name || ""));
    });

    return data;
  }
  resetInstances() {
    this.instancesStatus = null;
    this.instances = [];
  }

	get taskList(){
		return getTaskList(this.instances)
	}

  backupAllStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ManagerhBackupAllResponse, any>
  > | null = null;
  async backupAll() {
    this.backupAllStatus = fromPromise(api.manager.backupAll());

    runInAction(() => {
      this.instances.forEach((instance) => {
        Object.entries(instance.config?.services || {}).forEach(([serviceName, service]) => {
          Object.entries(service || {}).forEach(([taskName]) => {
            if (instance.name) {
              this.taskStatusMapStart(instance.name, serviceName, taskName, instance.proxyName);
            }
          });
        });
      });
    });

    const { data } = await this.backupAllStatus;

    return data;
  }
  resetBackupAll() {
    this.backupAllStatus = null;
  }

  backupStatusMapKey(
    instanceName: string,
    serviceName: string,
    taskName: string,
    proxyName?: string,
  ) {
    return [proxyName, instanceName, serviceName, taskName].join("_");
  }

  backupStatusMap = new Map<
    string,
    IPromiseBasedObservable<AxiosResponse<apiGenerated.ApiuResultSuccess, any>>
  >();
  async backup(instanceName: string, serviceName: string, taskName: string, proxyName?: string) {
    const request = !!proxyName
      ? fromPromise(api.manager.backupProxy(proxyName, instanceName, serviceName, taskName))
      : fromPromise(api.manager.backup(instanceName, serviceName, taskName));

    runInAction(() => {
      this.taskStatusMapStart(instanceName, serviceName, taskName, proxyName);
    });

    this.backupStatusMap.set(
      this.backupStatusMapKey(instanceName, serviceName, taskName, proxyName),
      request,
    );

    const { data } = await request;

    return data;
  }
  resetBackup() {
    this.backupStatusMap.clear();
  }

  snapshots: apiGenerated.ServicesGetSnapshotsResult = {
    items: [],
  };
  snapshotsStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ServicesGetSnapshotsResult, any>
  > | null = null;
  async fetchSnapshots(
    instanceName: string,
    serviceName: string,
    taskName: string,
    destinationName: string,
    proxyName?: string,
  ) {
    this.snapshotsStatus = !!proxyName
      ? fromPromise(
          api.manager.getSnapshotsProxy(
            proxyName,
            instanceName,
            serviceName,
            taskName,
            destinationName,
          ),
        )
      : fromPromise(api.manager.getSnapshots(instanceName, serviceName, taskName, destinationName));
    runInAction(() => {
      this.taskStatusMapStart(instanceName, serviceName, taskName);
    });

    const { data } = await this.snapshotsStatus;

    runInAction(() => {
      this.snapshots.items = (data.items || []).sort((a, b) => {
        return new Date(b.time || "").getTime() - new Date(a.time || "").getTime();
      });
    });

    return data;
  }
  resetSnapshots() {
    this.snapshotsForPrune = null;
    this.snapshots = { items: [] };
  }

  snapshotsForPrune: apiGenerated.ServicesGetPruneResult | null = null;
  snapshotsForPruneStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ServicesGetPruneResult, any>
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
    runInAction(() => {
      this.taskStatusMapStart(instanceName, serviceName, taskName);
    });

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

  prune: apiGenerated.ServicesGetPruneResult | null = null;
  pruneStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ServicesGetPruneResult, any>
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

    runInAction(() => {
      this.taskStatusMapStart(instanceName, serviceName, taskName);
    });

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

  stats: apiGenerated.ServicesGetStatsResult | null = null;
  statsStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ServicesGetStatsResult, any>
  > | null = null;
  async fetchStats(
    instanceName: string,
    serviceName: string,
    taskName: string,
    destinationName: string,
  ) {
    this.statsStatus = fromPromise(
      api.manager.getStats(instanceName, serviceName, taskName, destinationName),
    );
    runInAction(() => {
      this.taskStatusMapStart(instanceName, serviceName, taskName);
    });

    const { data } = await this.statsStatus;

    runInAction(() => {
      this.stats = data;
    });

    return data;
  }
  resetStats() {
    this.statsStatus = null;
    this.stats = null;
  }

  taskStatusMap = new Map<string, apiGenerated.ManagerTaskItem>();

  getTaskStatusMapKey(
    instanceName?: string,
    serviceName?: string,
    taskName?: string,
    proxyName?: string,
  ) {
    return [proxyName, instanceName, serviceName, taskName].join("_");
  }

  taskStatusMapStart(
    instanceName: string,
    serviceName: string,
    taskName: string,
    proxyName?: string,
  ) {
    const key = this.getTaskStatusMapKey(instanceName, serviceName, taskName, proxyName);

    let task = this.taskStatusMap.get(key);
    if (task) {
      task.isRunning = true;
    } else {
      task = {
        serviceName,
        taskName,
        isRunning: true,
      };
    }

    this.taskStatusMap.set(key, task);
  }
}

export const managerStore = new ManagerStore();
