import { StoreProvider } from "contexts/storeProvider";

import { useRoutes } from "react-router";
import { BrowserRouter } from "react-router-dom";

import { observer, useStores } from "stores";
import { RootStore } from "stores/root.store";

import { routes } from "./routes";

const Router = observer(() => {
  const { authStore } = useStores();
  return useRoutes(routes(authStore.isAuthorized));
});

const Renderer = () => {
  return (
    <BrowserRouter>
      <Router />
    </BrowserRouter>
  );
};

const App = observer(() => {
  return (
    <StoreProvider value={RootStore}>
      <Renderer />
    </StoreProvider>
  );
});

export default App;
