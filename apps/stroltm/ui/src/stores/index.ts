import { StoreContext } from "../contexts/storeProvider";

import { useContext } from "react";

export const useStores = () => {
  return useContext(StoreContext);
};

export { observer } from "mobx-react-lite";
