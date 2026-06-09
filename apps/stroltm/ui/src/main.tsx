import React from "react";
import ReactDOM from "react-dom/client";

import "antd/dist/reset.css";

import App from "./boot/App";
import "./index.css";
import { StoreProvider } from "./contexts/storeProvider";
import { RootStore } from "./stores/root.store";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <StoreProvider value={RootStore}>
      <App />
    </StoreProvider>
  </React.StrictMode>,
);
