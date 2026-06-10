import { createContext } from "react";

import type { RootStoreModel } from "../stores/root.store";

export const StoreContext = createContext<RootStoreModel>({} as RootStoreModel);
export const StoreProvider = StoreContext.Provider;
