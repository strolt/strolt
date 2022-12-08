import React from "react";

import ReactDOM from "react-dom/client";

import "antd/dist/reset.css";

import { RootStore } from "./stores/root.store";

import "./index.css";

import App from "./boot/App";
import { StoreProvider } from "./contexts/storeProvider";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <StoreProvider value={RootStore}>
      <App />
    </StoreProvider>
  </React.StrictMode>,
);
