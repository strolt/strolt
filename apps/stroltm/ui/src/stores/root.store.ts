import { managerStore } from "./manager.store";
import { authStore } from "./auth.store";

export type RootStoreModel = typeof RootStore;

export const RootStore = {
	authStore,
	managerStore,
};
