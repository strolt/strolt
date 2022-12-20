import { useContext } from "react";

import { StoreContext } from "../contexts/storeProvider";

export const useStores = () => {
  return useContext(StoreContext);
};

export { observer } from "mobx-react-lite";
