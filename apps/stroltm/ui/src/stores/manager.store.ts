import { AxiosResponse } from "axios";
import { makeAutoObservable, reaction, runInAction } from "mobx";
import * as api from "../api";
import * as apiGenerated from "../api/generated";
import { fromPromise, IPromiseBasedObservable } from "mobx-utils";
import { authStore } from "./auth.store";

export class ManagerStore {
	constructor() {
		makeAutoObservable(this);
	}

	instances: apiGenerated.ManagerhGetInstancesResultItem[] = [];
	instancesStatus: IPromiseBasedObservable<
		AxiosResponse<apiGenerated.ManagerhGetInstancesResult, any>
	> | null = null;
	async fetchInstances() {
		this.instancesStatus = fromPromise(api.manager.getInstances());

		const { data } = await this.instancesStatus;

		const sortedInstances = (data.data || []).sort((a, b) =>
			(a.instanceName || "").localeCompare(b.instanceName || "")
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

	backupStatus: IPromiseBasedObservable<
		AxiosResponse<apiGenerated.ApiuResultSuccess, any>
	> | null = null;
	async backup(instanceName: string, serviceName: string, taskName: string) {
		this.backupStatus = fromPromise(
			api.manager.backup(instanceName, serviceName, taskName)
		);

		const { data } = await this.backupStatus;

		return data;
	}
	resetBackup() {
		this.backupStatus = null;
	}

	snapshots: apiGenerated.ModelsServicesGetSnapshotsResult | null = null;
	snapshotsStatus: IPromiseBasedObservable<
		AxiosResponse<apiGenerated.ModelsServicesGetSnapshotsResult, any>
	> | null = null;
	async fetchSnapshots(
		instanceName: string,
		serviceName: string,
		taskName: string,
		destinationName: string
	) {
		this.snapshotsStatus = fromPromise(
			api.manager.getSnapshots(
				instanceName,
				serviceName,
				taskName,
				destinationName
			)
		);

		const { data } = await this.snapshotsStatus;

		const sortedSnapshots = (data.data || []).sort((a, b) => {
			return (
				new Date(b.time || "").getTime() - new Date(a.time || "").getTime()
			);
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
		destinationName: string
	) {
		this.snapshotsForPruneStatus = fromPromise(
			api.manager.getSnapshotsForPrune(
				instanceName,
				serviceName,
				taskName,
				destinationName
			)
		);

		const { data } = await this.snapshotsForPruneStatus;

		const sortedSnapshots = (data.data || []).sort((a, b) => {
			return (
				new Date(b.time || "").getTime() - new Date(a.time || "").getTime()
			);
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
		destinationName: string
	) {
		this.pruneStatus = fromPromise(
			api.manager.prune(instanceName, serviceName, taskName, destinationName)
		);

		const { data } = await this.pruneStatus;

		const sortedSnapshots = (data.data || []).sort((a, b) => {
			return (
				new Date(b.time || "").getTime() - new Date(a.time || "").getTime()
			);
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

reaction(
	() => authStore.isAuthorized,
	(isAuthorized) => {
		if (isAuthorized) {
			managerStore.fetchInstances();
			setInterval(() => managerStore.fetchInstances(), 5000);
		}
	}
);
