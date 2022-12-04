import { createContext } from "react";

import { RootStoreModel } from "../stores/root.store";

export const StoreContext = createContext<RootStoreModel>({} as RootStoreModel);
export const StoreProvider = StoreContext.Provider;
