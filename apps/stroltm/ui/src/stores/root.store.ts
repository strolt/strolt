import { authStore } from "./auth.store";
import { infoStore } from "./info.store";
import { managerStore } from "./manager.store";

export type RootStoreModel = typeof RootStore;

export const RootStore = {
  authStore,
  managerStore,
  infoStore,
};
